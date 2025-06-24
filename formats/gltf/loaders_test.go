package gltf

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// StandardBufferLoader Tests
// =============================================================================

func TestStandardBufferLoader_LoadBuffer(t *testing.T) {
	tests := []struct {
		name        string
		setupFiles  map[string][]byte
		basePath    string
		uri         string
		expectError bool
		errorMsg    string
		expectedLen int
	}{
		{
			name: "valid_data_uri",
			uri:  "data:application/octet-stream;base64,SGVsbG8gV29ybGQ=", // "Hello World"
			expectedLen: 11,
			expectError: false,
		},
		{
			name: "relative_file_path",
			setupFiles: map[string][]byte{
				"buffer.bin": []byte("test buffer data"),
			},
			basePath:    "",
			uri:         "buffer.bin",
			expectedLen: 16,
			expectError: false,
		},
		{
			name: "relative_file_path_with_base_path",
			setupFiles: map[string][]byte{
				"assets/buffer.bin": []byte("test buffer data"),
			},
			basePath:    "assets",
			uri:         "buffer.bin",
			expectedLen: 16,
			expectError: false,
		},
		{
			name: "absolute_file_path",
			setupFiles: map[string][]byte{
				"buffer.bin": []byte("absolute path test"),
			},
			basePath:    "",
			uri:         "", // Will be set to absolute path in test
			expectedLen: 18,
			expectError: false,
		},
		{
			name:        "missing_file",
			basePath:    "",
			uri:         "nonexistent.bin",
			expectError: true,
			errorMsg:    "no such file",
		},
		{
			name:        "invalid_data_uri_no_comma",
			uri:         "data:application/octet-stream;base64SGVsbG8=",
			expectError: true,
			errorMsg:    "missing comma separator",
		},
		{
			name:        "invalid_data_uri_bad_base64",
			uri:         "data:application/octet-stream;base64,invalid_base64_data!!!",
			expectError: true,
			errorMsg:    "failed to decode base64 data",
		},
		{
			name:        "invalid_data_uri_no_base64",
			uri:         "data:application/octet-stream,plain_text",
			expectError: true,
			errorMsg:    "only base64 encoded data URIs are supported",
		},
		{
			name:        "invalid_data_uri_no_data_prefix",
			uri:         "application/octet-stream;base64,SGVsbG8=",
			expectError: true,
			errorMsg:    "no such file",
		},
		{
			name:        "empty_uri",
			uri:         "",
			expectError: true,
			errorMsg:    "no such file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			
			// Setup files if needed
			for filename, content := range tt.setupFiles {
				fullPath := filepath.Join(tempDir, filename)
				dir := filepath.Dir(fullPath)
				err := os.MkdirAll(dir, 0755)
				require.NoError(t, err)
				err = os.WriteFile(fullPath, content, 0644)
				require.NoError(t, err)
			}

			// Set up loader
			basePath := tt.basePath
			if basePath != "" && !filepath.IsAbs(basePath) {
				basePath = filepath.Join(tempDir, basePath)
			} else if basePath == "" {
				basePath = tempDir
			}

			loader := &StandardBufferLoader{
				BasePath: basePath,
			}

			// Handle absolute path test case
			uri := tt.uri
			if tt.name == "absolute_file_path" {
				uri = filepath.Join(tempDir, "buffer.bin")
			}

			// Test loading
			data, err := loader.LoadBuffer(uri)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				return
			}

			require.NoError(t, err)
			assert.Len(t, data, tt.expectedLen)
		})
	}
}

func TestStandardBufferLoader_DataURIEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		uri         string
		expectError bool
		errorMsg    string
		expectedLen int
	}{
		{
			name:        "data_uri_with_charset",
			uri:         "data:application/octet-stream;charset=utf-8;base64,SGVsbG8=",
			expectedLen: 5,
			expectError: false,
		},
		{
			name:        "data_uri_with_multiple_params",
			uri:         "data:application/octet-stream;param1=value1;base64;param2=value2,SGVsbG8=",
			expectedLen: 5,
			expectError: false,
		},
		{
			name:        "data_uri_case_sensitive_base64",
			uri:         "data:application/octet-stream;BASE64,SGVsbG8=",
			expectError: true,
			errorMsg:    "only base64 encoded data URIs are supported",
		},
		{
			name:        "data_uri_with_whitespace",
			uri:         "data: application/octet-stream ; base64 , SGVsbG8=",
			expectError: true,
			errorMsg:    "only base64 encoded data URIs are supported",
		},
		{
			name:        "empty_data_uri",
			uri:         "data:,",
			expectError: true,
			errorMsg:    "only base64 encoded data URIs are supported",
		},
		{
			name:        "malformed_data_uri_scheme",
			uri:         "dat:application/octet-stream;base64,SGVsbG8=",
			expectError: true,
			errorMsg:    "no such file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := &StandardBufferLoader{BasePath: ""}
			
			data, err := loader.LoadBuffer(tt.uri)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				return
			}

			require.NoError(t, err)
			assert.Len(t, data, tt.expectedLen)
		})
	}
}

// =============================================================================
// StandardImageLoader Tests
// =============================================================================

func TestStandardImageLoader_LoadImage(t *testing.T) {
	// Create test images
	testPNGImg := createTestImage(t, 16, 16, color.RGBA{255, 0, 0, 255})
	testPNGDataURI := imageToDataURI(t, testPNGImg, "png")

	tests := []struct {
		name        string
		setupFiles  map[string]image.Image
		basePath    string
		uri         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid_png_data_uri",
			uri:         testPNGDataURI,
			expectError: false,
		},
		{
			name: "relative_image_file",
			setupFiles: map[string]image.Image{
				"image.png": testPNGImg,
			},
			basePath:    "",
			uri:         "image.png",
			expectError: false,
		},
		{
			name: "relative_image_with_base_path",
			setupFiles: map[string]image.Image{
				"assets/image.png": testPNGImg,
			},
			basePath:    "assets",
			uri:         "image.png",
			expectError: false,
		},
		{
			name: "absolute_image_path",
			setupFiles: map[string]image.Image{
				"image.png": testPNGImg,
			},
			basePath:    "",
			uri:         "", // Will be set to absolute path in test
			expectError: false,
		},
		{
			name: "file_uri_scheme",
			setupFiles: map[string]image.Image{
				"image.png": testPNGImg,
			},
			basePath:    "",
			uri:         "", // Will be set to file:// URI in test
			expectError: false,
		},
		{
			name:        "missing_image_file",
			basePath:    "",
			uri:         "nonexistent.png",
			expectError: true,
			errorMsg:    "failed to open image file",
		},
		{
			name:        "invalid_data_uri_format",
			uri:         "data:invalid_format",
			expectError: true,
			errorMsg:    "missing comma separator",
		},
		{
			name:        "unsupported_image_type",
			uri:         "data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7",
			expectError: true,
			errorMsg:    "unsupported image type",
		},
		{
			name:        "invalid_file_uri",
			uri:         "file:///nonexistent/path/image.png",
			expectError: true,
			errorMsg:    "failed to open image file",
		},
		{
			name:        "corrupted_image_file",
			setupFiles: map[string]image.Image{
				"corrupted.png": nil, // Will write corrupted data
			},
			uri:         "corrupted.png",
			expectError: true,
			errorMsg:    "failed to decode image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			
			// Setup image files if needed
			for filename, img := range tt.setupFiles {
				fullPath := filepath.Join(tempDir, filename)
				dir := filepath.Dir(fullPath)
				err := os.MkdirAll(dir, 0755)
				require.NoError(t, err)
				
				if img == nil && strings.Contains(filename, "corrupted") {
					// Write corrupted PNG data
					err = os.WriteFile(fullPath, []byte("not a valid image"), 0644)
				} else {
					file, err := os.Create(fullPath)
					require.NoError(t, err)
					err = png.Encode(file, img)
					require.NoError(t, err)
					file.Close()
				}
				require.NoError(t, err)
			}

			// Set up loader
			basePath := tt.basePath
			if basePath != "" && !filepath.IsAbs(basePath) {
				basePath = filepath.Join(tempDir, basePath)
			} else if basePath == "" {
				basePath = tempDir
			}

			loader := &StandardImageLoader{
				BasePath: basePath,
			}

			// Handle special URI cases
			uri := tt.uri
			if tt.name == "absolute_image_path" {
				uri = filepath.Join(tempDir, "image.png")
			} else if tt.name == "file_uri_scheme" {
				uri = "file://" + filepath.Join(tempDir, "image.png")
			}

			// Test loading
			img, err := loader.LoadImage(uri)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, img)
			
			// Verify image dimensions
			bounds := img.Bounds()
			assert.Equal(t, 16, bounds.Dx())
			assert.Equal(t, 16, bounds.Dy())
		})
	}
}

func TestStandardImageLoader_DataURIValidation(t *testing.T) {
	// Create a valid test image for format validation tests
	testImg := createTestImage(t, 8, 8, color.RGBA{0, 255, 0, 255})
	validPNGData := imageToBase64(t, testImg, "png")
	
	tests := []struct {
		name        string
		uri         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid_png_explicit",
			uri:         "data:image/png;base64," + validPNGData,
			expectError: false,
		},
		{
			name:        "format_mismatch_jpeg_declared_png_actual",
			uri:         "data:image/jpeg;base64," + validPNGData,
			expectError: true,
			errorMsg:    "image format mismatch",
		},
		{
			name:        "missing_content_type",
			uri:         "data:;base64," + validPNGData,
			expectError: true,
			errorMsg:    "unsupported image type",
		},
		{
			name:        "invalid_base64_in_image",
			uri:         "data:image/png;base64,invalid_base64_data!!!",
			expectError: true,
			errorMsg:    "failed to decode base64 image data",
		},
		{
			name:        "no_base64_encoding",
			uri:         "data:image/png,raw_data",
			expectError: true,
			errorMsg:    "only base64 encoded data URIs are supported",
		},
		{
			name:        "empty_image_data",
			uri:         "data:image/png;base64,",
			expectError: true,
			errorMsg:    "failed to decode image",
		},
		{
			name:        "case_sensitive_content_type",
			uri:         "data:IMAGE/PNG;base64," + validPNGData,
			expectError: true,
			errorMsg:    "unsupported image type",
		},
		{
			name:        "invalid_data_uri_no_comma",
			uri:         "data:image/png;base64" + validPNGData,
			expectError: true,
			errorMsg:    "missing comma separator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := &StandardImageLoader{BasePath: ""}
			
			img, err := loader.LoadImage(tt.uri)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, img)
		})
	}
}

// =============================================================================
// NoOpImageLoader Tests  
// =============================================================================

func TestNoOpImageLoader_LoadImage(t *testing.T) {
	tests := []struct {
		name string
		uri  string
	}{
		{
			name: "any_uri",
			uri:  "any_uri_here",
		},
		{
			name: "data_uri",
			uri:  "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==",
		},
		{
			name: "file_path",
			uri:  "/path/to/nonexistent/image.png",
		},
		{
			name: "file_uri",
			uri:  "file:///path/to/image.png",
		},
		{
			name: "empty_uri",
			uri:  "",
		},
		{
			name: "invalid_uri",
			uri:  "invalid://uri/format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := &NoOpImageLoader{}
			
			img, err := loader.LoadImage(tt.uri)
			
			// NoOpImageLoader should always return (nil, nil)
			require.NoError(t, err)
			assert.Nil(t, img)
		})
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// createTestImage creates a simple test image with a solid color
func createTestImage(t *testing.T, width, height int, col color.RGBA) image.Image {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, col)
		}
	}
	return img
}

// imageToDataURI converts an image to a data URI
func imageToDataURI(t *testing.T, img image.Image, format string) string {
	t.Helper()
	buf := &bytes.Buffer{}
	
	switch format {
	case "png":
		require.NoError(t, png.Encode(buf, img))
	default:
		t.Fatalf("Unsupported format: %s", format)
	}
	
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:image/" + format + ";base64," + encoded
}

// imageToBase64 converts an image to base64 string without the data URI prefix
func imageToBase64(t *testing.T, img image.Image, format string) string {
	t.Helper()
	buf := &bytes.Buffer{}
	
	switch format {
	case "png":
		require.NoError(t, png.Encode(buf, img))
	default:
		t.Fatalf("Unsupported format: %s", format)
	}
	
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}