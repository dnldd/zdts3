package main

import (
	"os"
	"testing"

	"github.com/peterldowns/testy/assert"
)

func TestLoadConfig(t *testing.T) {

	cfg := Config{}

	// Ensure loading the configuration fails if there are no env files
	// or command arguments provided.
	err := loadConfig(&cfg, "")
	assert.Error(t, err)

	// Ensure the configuration be read directly from set environment variables.
	os.Setenv("endpoint", "test-endpoint")
	os.Setenv("accesskeyid", "test-accesskeyid")
	os.Setenv("secretaccesskey", "test-secretaccesskey")
	os.Setenv("bucket", "test-bucket")
	os.Setenv("sourcedir", "test-sourcedir")
	os.Setenv("zipfile", "test-zipfile")
	os.Setenv("loglevel", "debug")

	err = loadConfig(&cfg, "")
	assert.NoError(t, err)

	// Validate the configuration.
	assert.Equal(t, "test-endpoint", cfg.Endpoint)
	assert.Equal(t, "test-accesskeyid", cfg.AccessKeyID)
	assert.Equal(t, "test-secretaccesskey", cfg.SecretAccessKey)
	assert.Equal(t, "test-bucket", cfg.Bucket)
	assert.Equal(t, "test-sourcedir", cfg.SourceDir)
	assert.Equal(t, "test-zipfile", cfg.Zipfile)
	assert.Equal(t, "debug", cfg.LogLevel)

	// Reset the environment variables set.
	os.Setenv("endpoint", "")
	os.Setenv("accesskeyid", "")
	os.Setenv("secretaccesskey", "")
	os.Setenv("bucket", "")
	os.Setenv("sourcedir", "")
	os.Setenv("zipfile", "")
	os.Setenv("loglevel", "")

	// Load the configuration from an .env file.
	err = loadConfig(&cfg, "data/.env")
	assert.NoError(t, err)

	// Validate the configuration.
	assert.Equal(t, "test-endpoint", cfg.Endpoint)
	assert.Equal(t, "test-accesskeyid", cfg.AccessKeyID)
	assert.Equal(t, "test-secretaccesskey", cfg.SecretAccessKey)
	assert.Equal(t, "test-bucket", cfg.Bucket)
	assert.Equal(t, "test-sourcedir", cfg.SourceDir)
	assert.Equal(t, "test-zipfile", cfg.Zipfile)
	assert.Equal(t, "debug", cfg.LogLevel)
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		hasError bool
	}{
		{
			name: "valid config",
			config: Config{
				Endpoint:        "test-endpoint",
				AccessKeyID:     "test-accesskeyid",
				SecretAccessKey: "test-secretaccesskey",
				Bucket:          "test-bucket",
				SourceDir:       "test-sourcedir",
				Zipfile:         "test-zipfile",
				LogLevel:        "debug",
			},
			hasError: false,
		},
		{
			name: "missing endpoint",
			config: Config{
				AccessKeyID:     "test-accesskeyid",
				SecretAccessKey: "test-secretaccesskey",
				Bucket:          "test-bucket",
				SourceDir:       "test-sourcedir",
				Zipfile:         "test-zipfile",
				LogLevel:        "debug",
			},
			hasError: true,
		},
		{
			name: "missing access key ID",
			config: Config{
				Endpoint:        "test-endpoint",
				SecretAccessKey: "test-secretaccesskey",
				Bucket:          "test-bucket",
				SourceDir:       "test-sourcedir",
				Zipfile:         "test-zipfile",
				LogLevel:        "debug",
			},
			hasError: true,
		},
		{
			name: "missing secret access key",
			config: Config{
				Endpoint:    "test-endpoint",
				AccessKeyID: "test-accesskeyid",
				Bucket:      "test-bucket",
				SourceDir:   "test-sourcedir",
				Zipfile:     "test-zipfile",
				LogLevel:    "debug",
			},
			hasError: true,
		},
		{
			name: "missing bucket",
			config: Config{
				Endpoint:        "test-endpoint",
				AccessKeyID:     "test-accesskeyid",
				SecretAccessKey: "test-secretaccesskey",
				SourceDir:       "test-sourcedir",
				Zipfile:         "test-zipfile",
				LogLevel:        "debug",
			},
			hasError: true,
		},
		{
			name: "missing source directory",
			config: Config{
				Endpoint:        "test-endpoint",
				AccessKeyID:     "test-accesskeyid",
				SecretAccessKey: "test-secretaccesskey",
				Bucket:          "test-bucket",
				Zipfile:         "test-zipfile",
				LogLevel:        "debug",
			},
			hasError: true,
		},
		{
			name: "missing zip file",
			config: Config{
				Endpoint:        "test-endpoint",
				AccessKeyID:     "test-accesskeyid",
				SecretAccessKey: "test-secretaccesskey",
				Bucket:          "test-bucket",
				SourceDir:       "test-sourcedir",
				LogLevel:        "debug",
			},
			hasError: true,
		},
		{
			name: "missing log level",
			config: Config{
				Endpoint:        "test-endpoint",
				AccessKeyID:     "test-accesskeyid",
				SecretAccessKey: "test-secretaccesskey",
				Bucket:          "test-bucket",
				SourceDir:       "test-sourcedir",
				Zipfile:         "test-zipfile",
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
