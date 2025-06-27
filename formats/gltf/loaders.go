package gltf

import (
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg" // Import for side effects to register JPEG decoder
	_ "image/png"  // Import for side effects to register PNG decoder
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var _ BufferLoader = (*StandardLoader)(nil)
var _ ImageLoader = (*StandardLoader)(nil)

// StandardLoader implements both the BufferLoader and the ImageLoader interface for loading buffers and images from
// the file system or data URIs.
type StandardLoader struct {
	// BasePath is used to resolve relative file paths
	BasePath string
}

// LoadBuffer loads binary data for the given URI.
// It supports:
// - Data URIs (data:application/octet-stream;base64,...)
// - Relative file paths (resolved against BasePath)
// - Absolute file paths
func (l *StandardLoader) LoadBuffer(uri string) ([]byte, error) {
	// Handle data URIs
	if strings.HasPrefix(uri, "data:") {
		return decodeDataURI(uri)
	}

	// Handle file paths
	bufferPath := uri
	if !filepath.IsAbs(bufferPath) && l.BasePath != "" {
		bufferPath = filepath.Join(l.BasePath, bufferPath)
	}

	return os.ReadFile(bufferPath)
}

// LoadImage loads an image for the given URI.
// It supports:
// - Data URIs (data:image/png;base64,... or data:image/jpeg;base64,...)
// - file:// URIs
// - Relative file paths (resolved against BasePath)
// - Absolute file paths
func (l *StandardLoader) LoadImage(uri string) (image.Image, string, error) {
	// Handle data URIs
	if strings.HasPrefix(uri, "data:") {
		return loadImageFromDataURI(uri)
	}

	// Handle file:// URIs
	actualPath := uri
	if strings.HasPrefix(uri, "file://") {
		resolvedPath, err := resolveImagePath(uri)
		if err != nil {
			return nil, "", fmt.Errorf("failed to resolve image path %q: %w", uri, err)
		}
		actualPath = resolvedPath
	} else if !filepath.IsAbs(uri) && l.BasePath != "" {
		actualPath = filepath.Join(l.BasePath, uri)
	}

	// Open and decode the image file
	file, err := os.Open(actualPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open image file %q: %w", actualPath, err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image %q: %w", actualPath, err)
	}

	return img, format, nil
}

var _ ImageLoader = (*NoOpImageLoader)(nil)

// NoOpImageLoader implements the ImageLoader interface but does not actually load any images.
// This can be used to skip texture loading while still processing the rest of the GLTF file.
type NoOpImageLoader struct{}

// LoadImage always returns nil without error, effectively skipping image loading.
func (l *NoOpImageLoader) LoadImage(_ string) (image.Image, string, error) {
	return nil, "", nil
}

// resolveImagePath resolves a file:// URI to an absolute file path
func resolveImagePath(fileURI string) (string, error) {
	parsed, err := url.Parse(fileURI)
	if err != nil {
		return "", err
	}

	if parsed.Scheme != "file" {
		return "", fmt.Errorf("expected 'file' scheme, got %q", parsed.Scheme)
	}

	// URL path is already unescaped
	return parsed.Path, nil
}

// decodeDataURI extracts binary data from a data URI
func decodeDataURI(dataURI string) ([]byte, error) {
	if !strings.HasPrefix(dataURI, "data:") {
		return nil, fmt.Errorf("invalid data URI: must start with 'data:'")
	}

	// Find the comma that separates header from data
	commaIndex := strings.Index(dataURI, ",")
	if commaIndex == -1 {
		return nil, fmt.Errorf("invalid data URI: missing comma separator")
	}

	header := dataURI[5:commaIndex] // Skip "data:" prefix
	data := dataURI[commaIndex+1:]

	// Check if it's base64 encoded
	if !strings.Contains(header, "base64") {
		return nil, fmt.Errorf("only base64 encoded data URIs are supported")
	}

	// Decode base64 data
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %w", err)
	}

	return decoded, nil
}

// loadImageFromDataURI loads an image from a data URI
func loadImageFromDataURI(dataURI string) (image.Image, string, error) {
	if !strings.HasPrefix(dataURI, "data:") {
		return nil, "", fmt.Errorf("invalid data URI format")
	}

	// Find the comma that separates the header from the data
	commaIndex := strings.Index(dataURI, ",")
	if commaIndex == -1 {
		return nil, "", fmt.Errorf("invalid data URI: missing comma separator")
	}

	header := dataURI[5:commaIndex] // Skip "data:" prefix
	data := dataURI[commaIndex+1:]

	// Parse header to check content type and encoding
	headerParts := strings.Split(header, ";")
	if len(headerParts) == 0 {
		return nil, "", fmt.Errorf("invalid data URI: empty header")
	}

	// Check content type
	contentType := headerParts[0]
	if contentType != "image/jpeg" && contentType != "image/png" {
		return nil, "", fmt.Errorf("unsupported image type %q: only image/jpeg and image/png are supported", contentType)
	}

	// Check for base64 encoding
	hasBase64 := false
	for _, part := range headerParts[1:] {
		if part == "base64" {
			hasBase64 = true
			break
		}
	}

	if !hasBase64 {
		return nil, "", fmt.Errorf("only base64 encoded data URIs are supported")
	}

	// Decode base64 data
	imgData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode base64 image data: %w", err)
	}

	// Decode image
	img, format, err := image.Decode(strings.NewReader(string(imgData)))
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Verify format matches declared content type
	expectedFormat := ""
	switch contentType {
	case "image/jpeg":
		expectedFormat = "jpeg"
	case "image/png":
		expectedFormat = "png"
	}

	if format != expectedFormat {
		return nil, "", fmt.Errorf("image format mismatch: declared as %s but decoded as %s", contentType, format)
	}

	return img, format, nil
}
