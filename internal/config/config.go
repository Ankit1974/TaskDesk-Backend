// Package config handles loading and storing application configuration.
// It uses Viper to read from .env files and environment variables,
// and exposes a global Cfg variable for access across the application.
package config

import (
	"log"

	"github.com/spf13/viper"
)

// Cfg is the global configuration instance, set during LoadConfig().
// Accessible from any package as config.Cfg (e.g., config.Cfg.SupabaseJWTSecret).
var Cfg *Config

// Config holds all environment-specific settings for the application.
// Fields are mapped to environment variables via the `mapstructure` tag.
type Config struct {
	AppPort           string `mapstructure:"APP_PORT"`            // Server port (default: "8080")
	Env               string `mapstructure:"ENV"`                 // Environment: "development" or "production"
	SupabaseURL       string `mapstructure:"SUPABASE_URL"`        // Supabase project URL
	SupabaseKey       string `mapstructure:"SUPABASE_ANON_KEY"`   // Supabase anonymous/public API key
	SupabaseJWTSecret string `mapstructure:"SUPABASE_JWT_SECRET"` // Supabase JWT signing secret (used to verify access tokens)
	DatabaseURL       string `mapstructure:"DATABASE_URL"`        // PostgreSQL connection string (Supabase DB)
}

// LoadConfig reads configuration from the .env file and environment variables.
// It sets defaults for APP_PORT and ENV, then populates the global Cfg variable.
// Returns the loaded Config pointer.
func LoadConfig() *Config {
	// Set defaults for optional values
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("ENV", "development")

	// Read from .env file in the working directory
	viper.SetConfigFile(".env")
	// Also read from actual environment variables (takes precedence over .env)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: .env file not found, using default and env variables: %v", err)
	}

	// Unmarshal all values into the Config struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Error unmarshaling config: %v", err)
	}

	// Store globally so other packages can access it without passing config around
	Cfg = &config
	return &config
}
