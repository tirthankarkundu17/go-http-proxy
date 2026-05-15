package proxy

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	fhttp "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	tls_profiles "github.com/bogdanfinn/tls-client/profiles"
)

// Handler handles the proxy requests
type Handler struct {
	defaultHeaders map[string]string
	client         tls_client.HttpClient
}

// NewHandler creates a new proxy handler
func NewHandler(defaultHeaders map[string]string) *Handler {
	// Setup tls client to avoid Cloudflare/reCAPTCHA blocks
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(tls_profiles.Chrome_120),
		tls_client.WithNotFollowRedirects(),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		slog.Error("Failed to create tls client, falling back to default", "error", err)
		panic(err)
	}

	return &Handler{
		defaultHeaders: defaultHeaders,
		client:         client,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Determine the target URL from 'url' query parameter or 'X-Proxy-Url' header
	targetURLStr := r.Header.Get("X-Proxy-Url")
	if targetURLStr == "" {
		targetURLStr = r.URL.Query().Get("url")
	}

	if targetURLStr == "" {
		http.Error(w, "Missing target URL. Pass it via 'X-Proxy-Url' header or 'url' query parameter.", http.StatusBadRequest)
		return
	}

	targetURL, err := url.Parse(targetURLStr)
	if err != nil || targetURL.Scheme == "" || targetURL.Host == "" {
		http.Error(w, "Invalid target URL", http.StatusBadRequest)
		return
	}

	// Create a new HTTP request using fhttp
	proxyReq, err := fhttp.NewRequestWithContext(r.Context(), r.Method, targetURLStr, r.Body)
	if err != nil {
		slog.Error("Error creating request", "error", err, "url", targetURLStr)
		http.Error(w, fmt.Sprintf("Error creating request: %v", err), http.StatusInternalServerError)
		return
	}

	// Copy original headers to the new request
	for name, values := range r.Header {
		if name == "X-Proxy-Url" {
			continue
		}
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// Apply default headers if they are not already set by the user
	for key, defaultValue := range h.defaultHeaders {
		if proxyReq.Header.Get(key) == "" {
			proxyReq.Header.Set(key, defaultValue)
		}
	}

	// Send the request using tls-client
	resp, err := h.client.Do(proxyReq)
	if err != nil {
		slog.Error("Error forwarding request", "error", err, "url", targetURLStr)
		http.Error(w, fmt.Sprintf("Error forwarding request: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Headers to remove because tls-client auto-decompresses the body,
	// and we don't want to copy hop-by-hop headers.
	dropHeaders := map[string]bool{
		"Content-Encoding":  true,
		"Content-Length":    true,
		"Connection":        true,
		"Keep-Alive":        true,
		"Transfer-Encoding": true,
		"Te":                true,
		"Trailer":           true,
		"Upgrade":           true,
	}

	// Copy response headers back to the original response
	for name, values := range resp.Header {
		if dropHeaders[name] {
			continue
		}
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set the status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body back to the original response
	if _, err := io.Copy(w, resp.Body); err != nil {
		slog.Error("Error copying response body", "error", err)
	}

	slog.Info("Proxied request", "method", r.Method, "target", targetURLStr, "status", resp.StatusCode)
}
