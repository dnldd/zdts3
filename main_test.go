package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/peterldowns/testy/assert"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestPurgeDir(t *testing.T) {
	dir := t.TempDir()
	file, err := os.Create(filepath.Join(dir, "test.txt"))
	assert.NoError(t, err)
	file.Close()

	// Purge the directory.
	filter := uint64(time.Now().Add(time.Hour).UnixMilli())
	logger := zerolog.Nop()
	purgeDir(dir, filter, &logger)

	// Assert the directory is now empty.
	contents, err := os.ReadDir(dir)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(contents))
}

func TestZipDir(t *testing.T) {
	dir := t.TempDir()
	file, err := os.Create(filepath.Join(dir, "test.txt"))
	assert.NoError(t, err)
	file.Close()

	zipPath := filepath.Join(t.TempDir(), "test.zip")

	// Zip the directory.
	logger := zerolog.Nop()
	zipDir(dir, zipPath, &logger)

	// Assert the zip file exists.
	_, err = os.Stat(zipPath)
	assert.NoError(t, err)

	// Romove zip file.
	err = os.Remove(zipPath)
	assert.NoError(t, err)
}

func TestUploadZip(t *testing.T) {
	dir, err := os.MkdirTemp(filepath.Join(t.TempDir()), "tdir")
	assert.NoError(t, err)

	f, err := os.Create(filepath.Join(dir, "test.txt"))
	assert.NoError(t, err)

	_, err = f.WriteString("Hello!")
	assert.NoError(t, err)
	f.Close()

	zipPath := "test.zip"

	// Zip the directory.
	logger := log.With().Caller().Logger()
	ctx := context.Background()
	zipDir(dir, zipPath, &logger)

	// Assert the zip file exists.
	_, err = os.Stat(zipPath)
	assert.NoError(t, err)

	// Load the env file if it exists.
	_, err = os.Stat(".env")
	if err == nil {
		err = godotenv.Load()
		if err != nil {
			log.Fatal().Err(err).Msg("Loading environment variables")
		}
	}

	cfg := &s3Config{
		Bucket:   os.Getenv("BUCKET"),
		Endpoint: os.Getenv("ENDPOINT"),
		Options: &minio.Options{
			Creds:  credentials.NewStaticV4(os.Getenv("ACCESSKEYID"), os.Getenv("SECRETACCESSKEY"), ""),
			Secure: true,
		},
	}

	// Upload the zip file.
	uploadZip(ctx, zipPath, cfg, &logger)

	// Romove zip file.
	err = os.Remove(zipPath)
	assert.NoError(t, err)
}
