package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

// purgeDir removes files in the provided directory that are older than the provided timestamp filter.
func purgeDir(dir string, filter uint64, logger *zerolog.Logger) {
	files, err := os.ReadDir(dir)
	if err != nil {
		logger.Error().Err(err).Str("path", dir).Msg("Reading directory")
		return
	}

	for _, file := range files {
		// Use the file's modification time to determine if it should be deleted.
		fileName := file.Name()
		info, err := file.Info()
		if err != nil {
			logger.Error().Err(err).Str("file", fileName).Msg("Getting file info")
			continue
		}

		modTime := uint64(info.ModTime().UnixMilli())

		// If the file's modification timestamp is older than the filter, delete the file.
		if modTime < filter {
			logger.Info().Uint64("modification time", modTime).Uint64("filter", filter).
				Str("file", fileName).Msg("file is older than filter, removing")
			err = os.Remove(filepath.Join(dir, fileName))
			if err != nil {
				logger.Error().Err(err).Str("file", fileName).Msg("Removing old file")
				continue
			}
		}
	}
}

// zipDir zips contents of the provided directory into a zip file at the provided path.
func zipDir(dir string, zipPath string, logger *zerolog.Logger) {
	// Create the destination zip file.
	zipFile, err := os.Create(zipPath)
	if err != nil {
		logger.Error().Err(err).Str("path", zipPath).Msg("Creating zip file")
		return
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Walk the directory and add each file to the zip.
	err = filepath.WalkDir(dir, fs.WalkDirFunc(func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories.
		if d.IsDir() {
			return nil
		}

		// Get the relative path of the file.
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		// Create a new zip file for the current file.
		zipFile, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		// Open the current file.
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Copy the file into the zip.
		_, err = io.Copy(zipFile, file)
		if err != nil {
			return err
		}

		return nil
	}))
	if err != nil {
		logger.Error().Err(err).Str("path", dir).Msg("Walking directory")
		return
	}

	err = zipWriter.Close()
	if err != nil {
		logger.Error().Err(err).Str("path", zipPath).Msg("Closing zip writer")
		return
	}
}

// uploadZip uploads the zip file at the provided path to the provided S3 or S3-compatible bucket.
func uploadZip(ctx context.Context, zipPath string, cfg *s3Config, logger *zerolog.Logger) {
	// Upload the zip file to an S3 or S3-compatible bucket.
	mnc, err := minio.New(cfg.Endpoint, cfg.Options)
	if err != nil {
		logger.Error().Err(err).Msg("Creating minio client")
		return
	}

	bucketName := cfg.Bucket
	contentType := "application/zip"
	objectName := filepath.Base(zipPath)

	info, err := mnc.FPutObject(ctx, bucketName, objectName, zipPath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		logger.Error().Err(err).Str("bucket", bucketName).Str("object", objectName).Msg("Uploading zip file")
		return
	}

	logger.Info().Str("bucket", bucketName).Str("object", objectName).Int64("size", info.Size).Msg("Uploaded zip file")

	// Remove the zip file after uploading.
	err = os.Remove(zipPath)
	if err != nil {
		logger.Error().Err(err).Str("path", zipPath).Msg("Removing zip file")
	}
}

// archive archives the contents of the provided directory by purging old files and zipping the
// recent files in the directory.
func archive(ctx context.Context, dir string, cfg *s3Config, logger *zerolog.Logger) {
	// The purge filter is set to 10 minutes before midnight of the previous day.
	now := time.Now()
	filter := time.Date(now.Year(), now.Month(), now.Day(), 23, 50, 0, 0, now.Location()).AddDate(0, 0, -1)

	// Purge the directory of old files.
	purgeDir(dir, uint64(filter.UnixMilli()), logger)

	// Zip the directory.
	zipPath := filepath.Join(dir, fmt.Sprintf("dump-%s.zip", now.Format("20060102150405")))
	zipDir(dir, zipPath, logger)

	// Upload the zip file to the S3/S3-compatible bucket.
	uploadZip(ctx, zipPath, cfg, logger)
}

// handleTermination processes context cancellation signals or interrupt signals from the OS.
func handleTermination(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()

	// Listen for interrupt signals.
	signals := []os.Signal{os.Interrupt}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, signals...)

	// Wait for the context to be cancelled or an interrupt signal.
	for {
		select {
		case <-ctx.Done():
			return

		case <-interrupt:
			cancel()
		}
	}
}

func main() {
	// Create the logger.
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	logger := log.With().Caller().Logger()

	cfg := Config{}
	err := loadConfig(&cfg, "")
	if err != nil {
		logger.Error().Msgf("Loading configuration: %s", strings.ReplaceAll(err.Error(), "\n", ". "))
		return
	}

	switch cfg.LogLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the S3 configuration.
	s3Cfg := &s3Config{
		Endpoint: cfg.Endpoint,
		Bucket:   cfg.Bucket,
		Options: &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
			Secure: true,
		},
	}

	// Create the cron scheduler.
	s, err := gocron.NewScheduler()
	if err != nil {
		logger.Error().Err(err).Msg("Creating scheduler")
		return
	}

	_, err = s.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(23, 50, 0))),
		gocron.NewTask(
			archive,
			cfg.SourceDir,
			s3Cfg,
			&logger,
		),
	)
	if err != nil {
		logger.Error().Err(err).Msg("Creating job")
		return
	}

	s.Start()

	logger.Info().Msg("Archiver started")

	wg.Add(1)
	go handleTermination(ctx, cancel, &wg)
	wg.Wait()
}
