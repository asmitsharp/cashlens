package services

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// ValidationResult contains the results of file validation
type ValidationResult struct {
	Valid        bool
	DetectedType string // "CSV", "XLSX", "PDF"
	ContentType  string
	Size         int64
	Errors       []string
	Warnings     []string
}

// FileValidator validates uploaded files for security and format compliance
type FileValidator struct {
	maxSizeBytes int64
	allowedTypes map[string]bool
	magicBytes   map[string][]byte
}

// File magic bytes signatures
var fileMagicBytes = map[string][]byte{
	"CSV":  []byte(""),               // CSV has no magic bytes, text-based
	"XLSX": {0x50, 0x4B, 0x03, 0x04}, // ZIP signature (XLSX is a ZIP)
	"PDF":  {0x25, 0x50, 0x44, 0x46}, // %PDF
}

// Allowed MIME types for file uploads
var allowedMimeTypes = map[string]bool{
	"text/csv":                 true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
	"application/pdf": true,
}

// Allowed file extensions
var allowedExtensions = map[string]bool{
	".csv":  true,
	".xlsx": true,
	".xls":  true,
	".pdf":  true,
}

// NewFileValidator creates a new file validator with the specified maximum file size
func NewFileValidator(maxSizeBytes int64) *FileValidator {
	return &FileValidator{
		maxSizeBytes: maxSizeBytes,
		allowedTypes: allowedMimeTypes,
		magicBytes:   fileMagicBytes,
	}
}

// ValidateFile performs comprehensive validation on an uploaded file
func (v *FileValidator) ValidateFile(reader io.Reader, filename, contentType string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:       true,
		ContentType: contentType,
		Errors:      []string{},
		Warnings:    []string{},
	}

	// 1. Validate filename
	if err := v.ValidateFilename(filename); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
	}

	// 2. Validate MIME type
	if err := v.ValidateMimeType(contentType); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
	}

	// 3. Read entire file content to validate size and magic bytes
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 4. Validate file size
	result.Size = int64(len(data))
	if err := v.ValidateFileSize(result.Size); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
	}

	// 5. Detect file type from magic bytes
	detectedType, err := v.ValidateMagicBytes(data)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
	} else {
		result.DetectedType = detectedType

		// 6. Check MIME type matches detected type
		if !v.isContentTypeMatch(contentType, detectedType) {
			result.Valid = false
			result.Errors = append(result.Errors, "MIME type does not match file content")
		}
	}

	return result, nil
}

// ValidateFilename validates the filename for security issues
func (v *FileValidator) ValidateFilename(filename string) error {
	// Check for empty filename
	if filename == "" {
		return errors.New("filename cannot be empty")
	}

	// Check for path traversal attempts
	if strings.Contains(filename, "..") {
		return errors.New("filename contains path traversal")
	}

	// Check for null bytes
	if strings.Contains(filename, "\x00") {
		return errors.New("filename contains null bytes")
	}

	// Check for absolute paths
	if strings.HasPrefix(filename, "/") || strings.HasPrefix(filename, "\\") {
		return errors.New("filename cannot be absolute path")
	}

	// Check extension
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return errors.New("filename must have an extension")
	}

	if !allowedExtensions[ext] {
		return fmt.Errorf("unsupported file extension: %s", ext)
	}

	return nil
}

// ValidateMimeType validates the MIME type is allowed
func (v *FileValidator) ValidateMimeType(contentType string) error {
	if contentType == "" {
		return errors.New("MIME type cannot be empty")
	}

	if !v.allowedTypes[contentType] {
		return fmt.Errorf("unsupported MIME type: %s", contentType)
	}

	return nil
}

// ValidateMagicBytes detects and validates file type based on magic bytes
func (v *FileValidator) ValidateMagicBytes(data []byte) (string, error) {
	if len(data) == 0 {
		return "", errors.New("empty file")
	}

	// Check for PDF signature
	if bytes.HasPrefix(data, v.magicBytes["PDF"]) {
		return "PDF", nil
	}

	// Check for XLSX/ZIP signature
	if bytes.HasPrefix(data, v.magicBytes["XLSX"]) {
		return "XLSX", nil
	}

	// CSV detection: text-based file without binary magic bytes
	// Check if content appears to be text (no null bytes, printable characters)
	if v.isTextContent(data) {
		return "CSV", nil
	}

	return "", errors.New("unsupported file type based on content")
}

// ValidateFileSize validates the file size is within limits
func (v *FileValidator) ValidateFileSize(size int64) error {
	if size < 0 {
		return errors.New("invalid file size")
	}

	if size == 0 {
		return errors.New("empty file")
	}

	if size > v.maxSizeBytes {
		return fmt.Errorf("file size (%d bytes) exceeds maximum allowed size (%d bytes)", size, v.maxSizeBytes)
	}

	return nil
}

// isContentTypeMatch checks if the MIME type matches the detected file type
func (v *FileValidator) isContentTypeMatch(contentType, detectedType string) bool {
	switch detectedType {
	case "CSV":
		return contentType == "text/csv"
	case "XLSX":
		// Both XLSX and XLS use similar structures
		return contentType == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" ||
			contentType == "application/vnd.ms-excel"
	case "PDF":
		return contentType == "application/pdf"
	default:
		return false
	}
}

// isTextContent checks if the data appears to be text (for CSV detection)
func (v *FileValidator) isTextContent(data []byte) bool {
	// Check first 512 bytes (or less if file is smaller)
	checkLen := len(data)
	if checkLen > 512 {
		checkLen = 512
	}

	sample := data[:checkLen]

	// Text files shouldn't have null bytes
	if bytes.Contains(sample, []byte{0x00}) {
		return false
	}

	// Count printable characters
	printable := 0
	for _, b := range sample {
		// Printable ASCII + common whitespace (tab, newline, carriage return)
		if (b >= 0x20 && b <= 0x7E) || b == 0x09 || b == 0x0A || b == 0x0D {
			printable++
		}
	}

	// If more than 95% of characters are printable, consider it text
	return float64(printable)/float64(len(sample)) > 0.95
}
