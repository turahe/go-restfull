package utils

import "strings"

// IsSupportedMediaType checks if the given content type is supported for upload
// This function validates MIME types against a predefined list of supported file types
// including images, documents, audio, video, and archive files.
//
// Parameters:
//   - contentType: The MIME type string to validate (e.g., "image/jpeg", "application/pdf")
//
// Returns:
//   - bool: true if the content type is supported, false otherwise
//
// Usage examples:
//   if utils.IsSupportedMediaType("image/jpeg") {
//       // Process JPEG image
//   }
//   if utils.IsSupportedMediaType("application/zip") {
//       // Process ZIP archive
//   }
//
// Supported file types:
//   - Images: JPEG, PNG, GIF, WebP, SVG, BMP, TIFF
//   - Documents: PDF, Word, Excel, PowerPoint, Text, CSV
//   - Audio: MP3, WAV, OGG, AAC
//   - Video: MP4, AVI, MOV, WMV, FLV, WebM
//   - Archives: ZIP, RAR, 7Z, GZIP, TAR
func IsSupportedMediaType(contentType string) bool {
	supportedTypes := map[string]bool{
		// Images
		"image/jpeg":    true,
		"image/jpg":     true,
		"image/png":     true,
		"image/gif":     true,
		"image/webp":    true,
		"image/svg+xml": true,
		"image/bmp":     true,
		"image/tiff":    true,

		// Documents
		"application/pdf":    true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/vnd.ms-excel": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
		"application/vnd.ms-powerpoint":                                             true,
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
		"text/plain": true,
		"text/csv":   true,

		// Audio
		"audio/mpeg": true,
		"audio/mp3":  true,
		"audio/wav":  true,
		"audio/ogg":  true,
		"audio/aac":  true,

		// Video
		"video/mp4":  true,
		"video/avi":  true,
		"video/mov":  true,
		"video/wmv":  true,
		"video/flv":  true,
		"video/webm": true,

		// Archives
		"application/zip":              true,
		"application/x-rar-compressed": true,
		"application/x-7z-compressed":  true,
		"application/gzip":             true,
		"application/x-tar":            true,
	}

	return supportedTypes[contentType]
}

// GetSupportedMediaTypes returns a list of all supported MIME types
// This function provides access to the complete list of supported file types
// for use in documentation, API responses, or client-side validation.
//
// Returns:
//   - []string: A slice containing all supported MIME type strings
//
// Usage examples:
//   supportedTypes := utils.GetSupportedMediaTypes()
//   for _, mimeType := range supportedTypes {
//       fmt.Printf("Supports: %s\n", mimeType)
//   }
func GetSupportedMediaTypes() []string {
	return []string{
		// Images
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/svg+xml",
		"image/bmp",
		"image/tiff",

		// Documents
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"text/plain",
		"text/csv",

		// Audio
		"audio/mpeg",
		"audio/mp3",
		"audio/wav",
		"audio/ogg",
		"audio/aac",

		// Video
		"video/mp4",
		"video/avi",
		"video/mov",
		"video/wmv",
		"video/flv",
		"video/webm",

		// Archives
		"application/zip",
		"application/x-rar-compressed",
		"application/x-7z-compressed",
		"application/gzip",
		"application/x-tar",
	}
}

// IsImageType checks if the given content type represents an image file
// This function provides a quick way to determine if a file is an image
// without checking the full supported types list.
//
// Parameters:
//   - contentType: The MIME type string to check
//
// Returns:
//   - bool: true if the content type is an image, false otherwise
//
// Usage examples:
//   if utils.IsImageType("image/jpeg") {
//       // Process as image
//   }
func IsImageType(contentType string) bool {
	return strings.HasPrefix(contentType, "image/")
}

// IsDocumentType checks if the given content type represents a document file
// This function provides a quick way to determine if a file is a document
// including PDFs, Office documents, and text files.
//
// Parameters:
//   - contentType: The MIME type string to check
//
// Returns:
//   - bool: true if the content type is a document, false otherwise
//
// Usage examples:
//   if utils.IsDocumentType("application/pdf") {
//       // Process as document
//   }
func IsDocumentType(contentType string) bool {
	return strings.HasPrefix(contentType, "application/") || strings.HasPrefix(contentType, "text/")
}

// IsAudioType checks if the given content type represents an audio file
// This function provides a quick way to determine if a file is an audio file.
//
// Parameters:
//   - contentType: The MIME type string to check
//
// Returns:
//   - bool: true if the content type is audio, false otherwise
//
// Usage examples:
//   if utils.IsAudioType("audio/mp3") {
//       // Process as audio
//   }
func IsAudioType(contentType string) bool {
	return strings.HasPrefix(contentType, "audio/")
}

// IsVideoType checks if the given content type represents a video file
// This function provides a quick way to determine if a file is a video file.
//
// Parameters:
//   - contentType: The MIME type string to check
//
// Returns:
//   - bool: true if the content type is video, false otherwise
//
// Usage examples:
//   if utils.IsVideoType("video/mp4") {
//       // Process as video
//   }
func IsVideoType(contentType string) bool {
	return strings.HasPrefix(contentType, "video/")
}

// IsArchiveType checks if the given content type represents an archive file
// This function provides a quick way to determine if a file is an archive file.
//
// Parameters:
//   - contentType: The MIME type string to check
//
// Returns:
//   - bool: true if the content type is an archive, false otherwise
//
// Usage examples:
//   if utils.IsArchiveType("application/zip") {
//       // Process as archive
//   }
func IsArchiveType(contentType string) bool {
	return strings.HasPrefix(contentType, "application/") &&
		(strings.Contains(contentType, "zip") ||
			strings.Contains(contentType, "rar") ||
			strings.Contains(contentType, "7z") ||
			strings.Contains(contentType, "gzip") ||
			strings.Contains(contentType, "tar"))
}
