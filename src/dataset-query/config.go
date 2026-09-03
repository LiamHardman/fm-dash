package main

import "os"

// Config holds the Dataset Query Service's runtime configuration, sourced from
// environment variables.
type Config struct {
	Port              string
	DatasetStorageDir string
}

func loadConfig() Config {
	return Config{
		Port:              getEnv("PORT", "8092"),
		DatasetStorageDir: getEnv("DATASET_STORAGE_DIR", "./data/datasets"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
