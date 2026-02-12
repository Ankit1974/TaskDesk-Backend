package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort     string `mapstructure:"APP_PORT"`
	Env         string `mapstructure:"ENV"`
	SupabaseURL string `mapstructure:"SUPABASE_URL"`
	SupabaseKey string `mapstructure:"SUPABASE_ANON_KEY"`
	DatabaseURL string `mapstructure:"DATABASE_URL"`
}

func LoadConfig() *Config {
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("ENV", "development")

	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: .env file not found, using default and env variables: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Error unmarshaling config: %v", err)
	}

	return &config
}
