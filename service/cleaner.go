package service

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// StartImageCleaner starts a background task to clean up old images
func StartImageCleaner() {
	// Run immediately on startup
	cleanOldImages()

	// Then run every 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			cleanOldImages()
		}
	}()
}

func cleanOldImages() {
	logDir := "./logs"
	if common.LogDir != nil && *common.LogDir != "" {
		logDir = *common.LogDir
	}
	saveDir := filepath.Join(logDir, "rejected_images")

	common.SysLog(fmt.Sprintf("Starting image cleanup in %s", saveDir))

	// Check if directory exists
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		return
	}

	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	count := 0
	errors := 0

	err := filepath.Walk(saveDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if info.ModTime().Before(cutoff) {
				if err := os.Remove(path); err != nil {
					common.SysError(fmt.Sprintf("Failed to remove old image %s: %v", path, err))
					errors++
				} else {
					count++
				}
			}
		}
		return nil
	})

	if err != nil {
		common.SysError(fmt.Sprintf("Error walking image directory: %v", err))
	}

	if count > 0 || errors > 0 {
		common.SysLog(fmt.Sprintf("Image cleanup finished: %d files removed, %d errors", count, errors))
	}
}
