package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
)

var registeredFlags = make(map[string]bool)

// registeredFlag registers command line arguments and tracks them to avoid reregistration.
func registerFlag(name string, value *string, usage string) {
	defaultValue := os.Getenv(name)

	if !registeredFlags[name] {
		flag.StringVar(value, name, defaultValue, usage)
		registeredFlags[name] = true
	}

	if registeredFlags[name] && defaultValue != "" {
		*value = defaultValue
	}
}

// s3Config is the access configuration for an S3 or S3-compatible bucket.
type s3Config struct {
	Endpoint string
	Bucket   string
	Options  *minio.Options
}

// Config is the configuration struct for the service.
type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	SourceDir       string
	LogLevel        string
}

// validate ensures that the configuration is valid.
func (c *Config) validate() error {
	var errs error

	if c.Endpoint == "" {
		errs = errors.Join(errs, fmt.Errorf("s3/s3-compatible endpoint required"))
	}

	if c.AccessKeyID == "" {
		errs = errors.Join(errs, fmt.Errorf("access key ID required"))
	}

	if c.SecretAccessKey == "" {
		errs = errors.Join(errs, fmt.Errorf("secret access key required"))
	}

	if c.Bucket == "" {
		errs = errors.Join(errs, fmt.Errorf("bucket required"))
	}

	if c.SourceDir == "" {
		errs = errors.Join(errs, fmt.Errorf("source directory required"))
	}

	if c.LogLevel == "" {
		errs = errors.Join(errs, fmt.Errorf("log level required"))
	}

	return errs
}

// loadConfig loads the configuration from environment variables and command line flags.
func loadConfig(cfg *Config, path string) error {
	if path == "" {
		path = ".env"
	}

	_, err := os.Stat(path)
	if err == nil {
		err := godotenv.Load(path)
		if err != nil {
			return fmt.Errorf("loading .env file: %w", err)
		}
	}

	// Register command line arguments using loaded environment variables as defaults.
	registerFlag("endpoint", &cfg.Endpoint, "S3 or S3-compatible endpoint")
	registerFlag("accesskeyid", &cfg.AccessKeyID, "S3 access key ID")
	registerFlag("secretaccesskey", &cfg.SecretAccessKey, "S3 secret access key")
	registerFlag("bucket", &cfg.Bucket, "S3 or S3-compatible bucket name")
	registerFlag("sourcedir", &cfg.SourceDir, "Source directory to archive")
	registerFlag("loglevel", &cfg.LogLevel, "Log level (debug, info, warn, error, fatal)")

	// Parse command-line flags.
	flag.Parse()

	return cfg.validate()
}
