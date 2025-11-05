package services

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateFilename_Valid tests valid filename validation
func TestValidateFilename_Valid(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	testCases := []struct {
		name     string
		filename string
	}{
		{"CSV file", "transactions.csv"},
		{"XLSX file", "report.xlsx"},
		{"XLS file", "data.xls"},
		{"PDF file", "invoice.pdf"},
		{"Filename with spaces", "my file.csv"},
		{"Filename with numbers", "report2024.xlsx"},
		{"Filename with underscore", "bank_statement.csv"},
		{"Filename with hyphen", "quarterly-report.pdf"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateFilename(tc.filename)
			assert.NoError(t, err, "Expected %s to be valid", tc.filename)
		})
	}
}

// TestValidateFilename_Empty tests empty filename validation
func TestValidateFilename_Empty(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	err := validator.ValidateFilename("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "filename cannot be empty")
}

// TestValidateFilename_PathTraversal tests path traversal attack prevention
func TestValidateFilename_PathTraversal(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	testCases := []struct {
		name     string
		filename string
	}{
		{"Double dots", "../../../etc/passwd.csv"},
		{"Double dots in middle", "dir/../../../file.csv"},
		{"Windows path traversal", "..\\..\\windows\\system32\\file.csv"},
		{"Multiple traversals", "../../../../file.csv"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateFilename(tc.filename)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "path traversal")
		})
	}
}

// TestValidateFilename_NullBytes tests null byte injection prevention
func TestValidateFilename_NullBytes(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	err := validator.ValidateFilename("file\x00.csv")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "null bytes")
}

// TestValidateFilename_AbsolutePath tests absolute path prevention
func TestValidateFilename_AbsolutePath(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	testCases := []struct {
		name     string
		filename string
	}{
		{"Unix absolute path", "/etc/passwd.csv"},
		{"Windows absolute path", "\\Windows\\System32\\file.csv"},
		{"Root path", "/file.csv"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateFilename(tc.filename)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "absolute path")
		})
	}
}

// TestValidateFilename_NoExtension tests missing extension validation
func TestValidateFilename_NoExtension(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	err := validator.ValidateFilename("filewithoutext")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must have an extension")
}

// TestValidateFilename_UnsupportedExtension tests unsupported file extension
func TestValidateFilename_UnsupportedExtension(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	testCases := []struct {
		name     string
		filename string
	}{
		{"Executable", "malicious.exe"},
		{"Script", "script.sh"},
		{"JavaScript", "code.js"},
		{"Image", "photo.jpg"},
		{"Zip", "archive.zip"},
		{"HTML", "page.html"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateFilename(tc.filename)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "unsupported file extension")
		})
	}
}

// TestValidateMimeType_Valid tests valid MIME type validation
func TestValidateMimeType_Valid(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	testCases := []struct {
		name        string
		contentType string
	}{
		{"CSV", "text/csv"},
		{"XLS", "application/vnd.ms-excel"},
		{"XLSX", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		{"PDF", "application/pdf"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateMimeType(tc.contentType)
			assert.NoError(t, err)
		})
	}
}

// TestValidateMimeType_Invalid tests invalid MIME type rejection
func TestValidateMimeType_Invalid(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	testCases := []struct {
		name        string
		contentType string
	}{
		{"Image", "image/jpeg"},
		{"HTML", "text/html"},
		{"JSON", "application/json"},
		{"XML", "application/xml"},
		{"ZIP", "application/zip"},
		{"Executable", "application/x-msdownload"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateMimeType(tc.contentType)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "unsupported MIME type")
		})
	}
}

// TestValidateMimeType_Empty tests empty MIME type
func TestValidateMimeType_Empty(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	err := validator.ValidateMimeType("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

// TestValidateMagicBytes_CSV tests CSV magic byte detection
func TestValidateMagicBytes_CSV(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	// CSV has no magic bytes, it's text-based
	csvData := []byte("Date,Description,Amount\n01/01/2024,Test,100.00")

	detectedType, err := validator.ValidateMagicBytes(csvData)
	assert.NoError(t, err)
	assert.Equal(t, "CSV", detectedType)
}

// TestValidateMagicBytes_XLSX tests XLSX magic byte detection
func TestValidateMagicBytes_XLSX(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	// XLSX files start with ZIP signature: PK\x03\x04
	xlsxData := []byte{0x50, 0x4B, 0x03, 0x04, 0x00, 0x00}

	detectedType, err := validator.ValidateMagicBytes(xlsxData)
	assert.NoError(t, err)
	assert.Equal(t, "XLSX", detectedType)
}

// TestValidateMagicBytes_PDF tests PDF magic byte detection
func TestValidateMagicBytes_PDF(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	// PDF files start with %PDF
	pdfData := []byte{0x25, 0x50, 0x44, 0x46, 0x2D, 0x31, 0x2E, 0x34}

	detectedType, err := validator.ValidateMagicBytes(pdfData)
	assert.NoError(t, err)
	assert.Equal(t, "PDF", detectedType)
}

// TestValidateMagicBytes_Invalid tests invalid magic bytes
func TestValidateMagicBytes_Invalid(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	// Random binary data that doesn't match any signature
	invalidData := []byte{0xFF, 0xD8, 0xFF, 0xE0} // JPEG signature

	_, err := validator.ValidateMagicBytes(invalidData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported file type")
}

// TestValidateMagicBytes_Empty tests empty data
func TestValidateMagicBytes_Empty(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	_, err := validator.ValidateMagicBytes([]byte{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty file")
}

// TestValidateFileSize_Valid tests valid file size
func TestValidateFileSize_Valid(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024) // 10MB

	testCases := []struct {
		name string
		size int64
	}{
		{"Small file", 1024},          // 1KB
		{"Medium file", 5 * 1024 * 1024}, // 5MB
		{"Max size", 10 * 1024 * 1024},   // 10MB
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateFileSize(tc.size)
			assert.NoError(t, err)
		})
	}
}

// TestValidateFileSize_TooLarge tests file size exceeding limit
func TestValidateFileSize_TooLarge(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024) // 10MB

	err := validator.ValidateFileSize(11 * 1024 * 1024) // 11MB
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum")
}

// TestValidateFileSize_Zero tests zero byte file
func TestValidateFileSize_Zero(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	err := validator.ValidateFileSize(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty file")
}

// TestValidateFileSize_Negative tests negative file size
func TestValidateFileSize_Negative(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	err := validator.ValidateFileSize(-1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid file size")
}

// TestValidateFile_ValidCSV tests complete validation of valid CSV
func TestValidateFile_ValidCSV(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	csvContent := "Date,Description,Amount\n01/01/2024,Test,100.00\n"
	reader := strings.NewReader(csvContent)

	result, err := validator.ValidateFile(reader, "test.csv", "text/csv")

	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Equal(t, "CSV", result.DetectedType)
	assert.Equal(t, "text/csv", result.ContentType)
	assert.Equal(t, int64(len(csvContent)), result.Size)
	assert.Empty(t, result.Errors)
}

// TestValidateFile_ValidXLSX tests complete validation of valid XLSX
func TestValidateFile_ValidXLSX(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	// XLSX file starts with ZIP signature
	xlsxContent := []byte{0x50, 0x4B, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00}
	reader := bytes.NewReader(xlsxContent)

	result, err := validator.ValidateFile(reader, "report.xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Equal(t, "XLSX", result.DetectedType)
	assert.Empty(t, result.Errors)
}

// TestValidateFile_ValidPDF tests complete validation of valid PDF
func TestValidateFile_ValidPDF(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	// PDF file starts with %PDF
	pdfContent := []byte{0x25, 0x50, 0x44, 0x46, 0x2D, 0x31, 0x2E, 0x34}
	reader := bytes.NewReader(pdfContent)

	result, err := validator.ValidateFile(reader, "invoice.pdf", "application/pdf")

	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Equal(t, "PDF", result.DetectedType)
	assert.Empty(t, result.Errors)
}

// TestValidateFile_InvalidFilename tests validation with invalid filename
func TestValidateFile_InvalidFilename(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	csvContent := "Date,Description,Amount\n"
	reader := strings.NewReader(csvContent)

	result, err := validator.ValidateFile(reader, "../../../etc/passwd.csv", "text/csv")

	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0], "path traversal")
}

// TestValidateFile_InvalidMimeType tests validation with invalid MIME type
func TestValidateFile_InvalidMimeType(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	csvContent := "Date,Description,Amount\n"
	reader := strings.NewReader(csvContent)

	result, err := validator.ValidateFile(reader, "test.csv", "image/jpeg")

	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0], "unsupported MIME type")
}

// TestValidateFile_MismatchedMimeAndContent tests MIME type not matching content
func TestValidateFile_MismatchedMimeAndContent(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	// Send PDF content but claim it's CSV
	pdfContent := []byte{0x25, 0x50, 0x44, 0x46, 0x2D, 0x31, 0x2E, 0x34}
	reader := bytes.NewReader(pdfContent)

	result, err := validator.ValidateFile(reader, "fake.csv", "text/csv")

	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[len(result.Errors)-1], "MIME type does not match")
}

// TestValidateFile_EmptyFile tests validation of empty file
func TestValidateFile_EmptyFile(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	reader := strings.NewReader("")

	result, err := validator.ValidateFile(reader, "empty.csv", "text/csv")

	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0], "empty file")
}

// TestValidateFile_TooLarge tests validation of oversized file
func TestValidateFile_TooLarge(t *testing.T) {
	validator := NewFileValidator(10) // 10 bytes max

	content := strings.Repeat("a", 100) // 100 bytes
	reader := strings.NewReader(content)

	result, err := validator.ValidateFile(reader, "large.csv", "text/csv")

	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0], "exceeds maximum")
}

// TestValidateFile_MultipleErrors tests accumulation of multiple errors
func TestValidateFile_MultipleErrors(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	// Invalid filename AND invalid MIME type
	csvContent := "Date,Description,Amount\n"
	reader := strings.NewReader(csvContent)

	result, err := validator.ValidateFile(reader, "../malicious.exe", "application/x-msdownload")

	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.GreaterOrEqual(t, len(result.Errors), 2, "Should have multiple errors")
}

// TestNewFileValidator tests constructor
func TestNewFileValidator(t *testing.T) {
	maxSize := int64(5 * 1024 * 1024) // 5MB
	validator := NewFileValidator(maxSize)

	assert.NotNil(t, validator)
	assert.Equal(t, maxSize, validator.maxSizeBytes)
}

// TestValidateFile_ReadError tests handling of read errors
func TestValidateFile_ReadError(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	// Create a reader that will fail
	errorReader := &errorReader{err: io.ErrUnexpectedEOF}

	_, err := validator.ValidateFile(errorReader, "test.csv", "text/csv")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

// errorReader is a helper for testing read errors
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

// TestValidateFile_CSVWithHeaders tests CSV validation with typical headers
func TestValidateFile_CSVWithHeaders(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	csvContent := `Date,Narration,Withdrawal Amt,Deposit Amt,Balance
01/01/2024,AWS SERVICES,500.00,,10000.00
02/01/2024,SALARY CREDIT,,50000.00,60000.00`

	reader := strings.NewReader(csvContent)

	result, err := validator.ValidateFile(reader, "hdfc_statement.csv", "text/csv")

	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Equal(t, "CSV", result.DetectedType)
	assert.Empty(t, result.Errors)
}

// TestValidateFile_XLSAlternativeMime tests XLS with alternative MIME type
func TestValidateFile_XLSAlternativeMime(t *testing.T) {
	validator := NewFileValidator(10 * 1024 * 1024)

	// XLS files also use ZIP-like structure
	xlsContent := []byte{0x50, 0x4B, 0x03, 0x04, 0x00, 0x00}
	reader := bytes.NewReader(xlsContent)

	result, err := validator.ValidateFile(reader, "data.xls", "application/vnd.ms-excel")

	require.NoError(t, err)
	assert.True(t, result.Valid)
	// XLS should be detected as XLSX (both are ZIP-based)
	assert.Equal(t, "XLSX", result.DetectedType)
}
