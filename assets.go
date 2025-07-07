package main

import (
	"fmt"
	"mime"
	"os"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func (cfg apiConfig) getSupportedAssetType(contentType string) (string, error) {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", err
	}

	switch mediaType {
	case "image/png":
		return ".png", nil
	case "image/jpeg":
		return ".jpg", nil
	default:
		return "", fmt.Errorf("Not a supported file type")
	}
}
