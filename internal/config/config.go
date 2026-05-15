package config

import "os"
// Config holds the application configuration
type Config struct {
	Port           string
	DefaultHeaders map[string]string
}

// LoadConfig returns the configuration for the application
func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Port: ":" + port,
		DefaultHeaders: map[string]string{
			"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
			"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			"Accept-Language": "en-US,en;q=0.9",
			"Connection":      "keep-alive",
			"Cache-Control":   "max-age=0",
		},
	}
}
