package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestDetectBank_HDFC(t *testing.T) {
	headers := []string{"Date", "Narration", "Chq./Ref.No.", "Value Dt", "Withdrawal Amt.", "Deposit Amt.", "Closing Balance"}
	bank := DetectBank(headers)
	assert.Equal(t, "HDFC", bank)
}

func TestDetectBank_ICICI(t *testing.T) {
	headers := []string{"Value Date", "Transaction Date", "Cheque Number", "Transaction Remarks", "Withdrawal Amount (INR)", "Deposit Amount (INR)", "Balance (INR)"}
	bank := DetectBank(headers)
	assert.Equal(t, "ICICI", bank)
}

func TestDetectBank_SBI(t *testing.T) {
	headers := []string{"Txn Date", "Description", "Ref No./Cheque No.", "Value Date", "Debit", "Credit", "Balance"}
	bank := DetectBank(headers)
	assert.Equal(t, "SBI", bank)
}

func TestDetectBank_Axis(t *testing.T) {
	headers := []string{"Transaction Date", "Particulars", "Cheque No.", "Dr/Cr", "Amount", "Balance"}
	bank := DetectBank(headers)
	assert.Equal(t, "Axis", bank)
}

func TestDetectBank_Kotak(t *testing.T) {
	headers := []string{"Date", "Description", "Ref No.", "Debit", "Credit", "Balance"}
	bank := DetectBank(headers)
	assert.Equal(t, "Kotak", bank)
}

func TestDetectBank_Unknown(t *testing.T) {
	headers := []string{"Random", "Headers", "That", "Dont", "Match"}
	bank := DetectBank(headers)
	assert.Equal(t, "UNKNOWN", bank)
}

func TestParseDate_DDMMYYYY(t *testing.T) {
	date, err := ParseDate("15/01/2024")
	require.NoError(t, err)
	assert.Equal(t, 2024, date.Year())
	assert.Equal(t, time.January, date.Month())
	assert.Equal(t, 15, date.Day())
}

func TestParseDate_DDMonYYYY(t *testing.T) {
	date, err := ParseDate("15-Jan-2024")
	require.NoError(t, err)
	assert.Equal(t, 2024, date.Year())
	assert.Equal(t, time.January, date.Month())
	assert.Equal(t, 15, date.Day())
}

func TestParseDate_YYYYMMDD(t *testing.T) {
	date, err := ParseDate("2024-01-15")
	require.NoError(t, err)
	assert.Equal(t, 2024, date.Year())
	assert.Equal(t, time.January, date.Month())
	assert.Equal(t, 15, date.Day())
}

func TestParseDate_Invalid(t *testing.T) {
	_, err := ParseDate("invalid-date")
	assert.Error(t, err)
}

func TestParseAmount_Simple(t *testing.T) {
	amount, err := ParseAmount("3500.00")
	require.NoError(t, err)
	assert.Equal(t, 3500.0, amount)
}

func TestParseAmount_WithCommas(t *testing.T) {
	amount, err := ParseAmount("1,50,000.00")
	require.NoError(t, err)
	assert.Equal(t, 150000.0, amount)
}

func TestParseAmount_WithRupeeSymbol(t *testing.T) {
	amount, err := ParseAmount("₹3500.00")
	require.NoError(t, err)
	assert.Equal(t, 3500.0, amount)
}

func TestParseAmount_WithRs(t *testing.T) {
	amount, err := ParseAmount("Rs. 3500.00")
	require.NoError(t, err)
	assert.Equal(t, 3500.0, amount)
}

func TestParseAmount_Empty(t *testing.T) {
	amount, err := ParseAmount("")
	require.NoError(t, err)
	assert.Equal(t, 0.0, amount)
}

func TestParseAmount_Invalid(t *testing.T) {
	_, err := ParseAmount("not-a-number")
	assert.Error(t, err)
}

func TestParseCSV_HDFC(t *testing.T) {
	file, err := os.Open("../../testdata/hdfc_sample.csv")
	require.NoError(t, err)
	defer file.Close()

	parser := NewParser()
	transactions, err := parser.ParseCSV(file)

	require.NoError(t, err)
	assert.Len(t, transactions, 10)

	// Validate first transaction (debit)
	assert.Equal(t, "AWS SERVICES", transactions[0].Description)
	assert.Equal(t, -3500.0, transactions[0].Amount)
	assert.Equal(t, "debit", transactions[0].TxnType)
	assert.Equal(t, 2024, transactions[0].TxnDate.Year())
	assert.Equal(t, time.January, transactions[0].TxnDate.Month())
	assert.Equal(t, 15, transactions[0].TxnDate.Day())

	// Validate second transaction (credit)
	assert.Equal(t, "SALARY CREDIT - ACME CORP", transactions[1].Description)
	assert.Equal(t, 50000.0, transactions[1].Amount)
	assert.Equal(t, "credit", transactions[1].TxnType)
}

func TestParseCSV_ICICI(t *testing.T) {
	file, err := os.Open("../../testdata/icici_sample.csv")
	require.NoError(t, err)
	defer file.Close()

	parser := NewParser()
	transactions, err := parser.ParseCSV(file)

	require.NoError(t, err)
	assert.Len(t, transactions, 10)

	// Validate first transaction
	assert.Equal(t, "PAYMENT TO AWS SERVICES", transactions[0].Description)
	assert.Equal(t, -3500.0, transactions[0].Amount)
	assert.Equal(t, "debit", transactions[0].TxnType)

	// Validate credit transaction
	assert.Equal(t, "SALARY CREDIT FROM ACME CORP", transactions[1].Description)
	assert.Equal(t, 50000.0, transactions[1].Amount)
	assert.Equal(t, "credit", transactions[1].TxnType)
}

func TestParseCSV_SBI(t *testing.T) {
	file, err := os.Open("../../testdata/sbi_sample.csv")
	require.NoError(t, err)
	defer file.Close()

	parser := NewParser()
	transactions, err := parser.ParseCSV(file)

	require.NoError(t, err)
	assert.Len(t, transactions, 10)

	// Validate first transaction
	assert.Equal(t, "PAYMENT TO AWS SERVICES", transactions[0].Description)
	assert.Equal(t, -3500.0, transactions[0].Amount)
	assert.Equal(t, "debit", transactions[0].TxnType)
}

func TestParseCSV_Axis(t *testing.T) {
	file, err := os.Open("../../testdata/axis_sample.csv")
	require.NoError(t, err)
	defer file.Close()

	parser := NewParser()
	transactions, err := parser.ParseCSV(file)

	require.NoError(t, err)
	assert.Len(t, transactions, 10)

	// Validate first transaction
	assert.Equal(t, "PAYMENT TO AWS SERVICES", transactions[0].Description)
	assert.Equal(t, -3500.0, transactions[0].Amount)
	assert.Equal(t, "debit", transactions[0].TxnType)

	// Validate credit transaction
	assert.Equal(t, "SALARY FROM ACME CORP", transactions[1].Description)
	assert.Equal(t, 50000.0, transactions[1].Amount)
	assert.Equal(t, "credit", transactions[1].TxnType)
}

func TestParseCSV_Kotak(t *testing.T) {
	file, err := os.Open("../../testdata/kotak_sample.csv")
	require.NoError(t, err)
	defer file.Close()

	parser := NewParser()
	transactions, err := parser.ParseCSV(file)

	require.NoError(t, err)
	assert.Len(t, transactions, 10)

	// Validate first transaction
	assert.Equal(t, "PAYMENT TO AWS SERVICES", transactions[0].Description)
	assert.Equal(t, -3500.0, transactions[0].Amount)
	assert.Equal(t, "debit", transactions[0].TxnType)
}

func TestParseCSV_EmptyFile(t *testing.T) {
	// Create temporary empty file
	tmpFile, err := os.CreateTemp("", "empty-*.csv")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	parser := NewParser()
	_, err = parser.ParseCSV(tmpFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty file")
}

func TestParseCSV_InvalidFormat(t *testing.T) {
	// Create temporary file with invalid headers
	tmpFile, err := os.CreateTemp("", "invalid-*.csv")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	tmpFile.WriteString("Invalid,Headers,That,Dont,Match\n")
	tmpFile.WriteString("some,random,data,here,ok\n")
	tmpFile.Seek(0, 0)

	parser := NewParser()
	_, err = parser.ParseCSV(tmpFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown bank format")
}

// XLSX Parser Tests

func TestParseXLSX_HDFC(t *testing.T) {
	file, err := os.Open("../../testdata/hdfc_sample.xlsx")
	require.NoError(t, err)
	defer file.Close()

	parser := NewParser()
	transactions, err := parser.ParseXLSX(file)

	require.NoError(t, err)
	assert.Len(t, transactions, 10)

	// Validate first transaction (debit)
	assert.Equal(t, "AWS SERVICES", transactions[0].Description)
	assert.Equal(t, -3500.0, transactions[0].Amount)
	assert.Equal(t, "debit", transactions[0].TxnType)
	assert.Equal(t, 2024, transactions[0].TxnDate.Year())
	assert.Equal(t, time.January, transactions[0].TxnDate.Month())
	assert.Equal(t, 15, transactions[0].TxnDate.Day())

	// Validate second transaction (credit)
	assert.Equal(t, "SALARY CREDIT - ACME CORP", transactions[1].Description)
	assert.Equal(t, 50000.0, transactions[1].Amount)
	assert.Equal(t, "credit", transactions[1].TxnType)
}

func TestParseXLSX_ICICI(t *testing.T) {
	file, err := os.Open("../../testdata/icici_sample.xlsx")
	require.NoError(t, err)
	defer file.Close()

	parser := NewParser()
	transactions, err := parser.ParseXLSX(file)

	require.NoError(t, err)
	assert.Len(t, transactions, 10)

	// Validate first transaction
	assert.Equal(t, "PAYMENT TO AWS SERVICES", transactions[0].Description)
	assert.Equal(t, -3500.0, transactions[0].Amount)
	assert.Equal(t, "debit", transactions[0].TxnType)

	// Validate credit transaction
	assert.Equal(t, "SALARY CREDIT FROM ACME CORP", transactions[1].Description)
	assert.Equal(t, 50000.0, transactions[1].Amount)
	assert.Equal(t, "credit", transactions[1].TxnType)
}

func TestParseXLSX_SBI(t *testing.T) {
	file, err := os.Open("../../testdata/sbi_sample.xlsx")
	require.NoError(t, err)
	defer file.Close()

	parser := NewParser()
	transactions, err := parser.ParseXLSX(file)

	require.NoError(t, err)
	assert.Len(t, transactions, 10)

	// Validate first transaction
	assert.Equal(t, "PAYMENT TO AWS SERVICES", transactions[0].Description)
	assert.Equal(t, -3500.0, transactions[0].Amount)
	assert.Equal(t, "debit", transactions[0].TxnType)
}

func TestParseXLSX_Axis(t *testing.T) {
	file, err := os.Open("../../testdata/axis_sample.xlsx")
	require.NoError(t, err)
	defer file.Close()

	parser := NewParser()
	transactions, err := parser.ParseXLSX(file)

	require.NoError(t, err)
	assert.Len(t, transactions, 10)

	// Validate first transaction
	assert.Equal(t, "PAYMENT TO AWS SERVICES", transactions[0].Description)
	assert.Equal(t, -3500.0, transactions[0].Amount)
	assert.Equal(t, "debit", transactions[0].TxnType)

	// Validate credit transaction
	assert.Equal(t, "SALARY FROM ACME CORP", transactions[1].Description)
	assert.Equal(t, 50000.0, transactions[1].Amount)
	assert.Equal(t, "credit", transactions[1].TxnType)
}

func TestParseXLSX_Kotak(t *testing.T) {
	file, err := os.Open("../../testdata/kotak_sample.xlsx")
	require.NoError(t, err)
	defer file.Close()

	parser := NewParser()
	transactions, err := parser.ParseXLSX(file)

	require.NoError(t, err)
	assert.Len(t, transactions, 10)

	// Validate first transaction
	assert.Equal(t, "PAYMENT TO AWS SERVICES", transactions[0].Description)
	assert.Equal(t, -3500.0, transactions[0].Amount)
	assert.Equal(t, "debit", transactions[0].TxnType)
}

func TestParseXLSX_EmptyFile(t *testing.T) {
	// Create temporary empty XLSX file
	f := excelize.NewFile()
	tmpFile, err := os.CreateTemp("", "empty-*.xlsx")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Save empty file with just Sheet1 (no data)
	buf := bytes.NewBuffer(nil)
	err = f.Write(buf)
	require.NoError(t, err)
	tmpFile.Write(buf.Bytes())
	tmpFile.Seek(0, 0)

	parser := NewParser()
	_, err = parser.ParseXLSX(tmpFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty file")
}

func TestParseXLSX_InvalidFormat(t *testing.T) {
	// Create temporary XLSX file with invalid headers
	f := excelize.NewFile()
	sheet := "Sheet1"

	// Add invalid headers
	headers := []string{"Invalid", "Headers", "That", "Dont", "Match"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Add a data row
	row := []string{"some", "random", "data", "here", "ok"}
	for i, val := range row {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(sheet, cell, val)
	}

	// Save to temp file
	tmpFile, err := os.CreateTemp("", "invalid-*.xlsx")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	buf := bytes.NewBuffer(nil)
	err = f.Write(buf)
	require.NoError(t, err)
	tmpFile.Write(buf.Bytes())
	tmpFile.Seek(0, 0)

	parser := NewParser()
	_, err = parser.ParseXLSX(tmpFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown bank format")
}

func TestParseXLSX_WithEmptyRowsAndSummary(t *testing.T) {
	// Create XLSX file with empty rows and summary rows
	f := excelize.NewFile()
	sheet := "Sheet1"

	// HDFC headers
	headers := []string{"Date", "Narration", "Chq./Ref.No.", "Value Dt", "Withdrawal Amt.", "Deposit Amt.", "Closing Balance"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Add valid transaction
	row1 := []interface{}{"15/01/2024", "AWS SERVICES", "UPI/123456", "15/01/2024", 3500.00, "", 450000.00}
	for i, val := range row1 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(sheet, cell, val)
	}

	// Add empty row (row 3 - all empty cells)
	// (excelize handles empty rows naturally)

	// Add summary row
	summaryRow := []interface{}{"Total", "", "", "", 3500.00, 0, ""}
	for i, val := range summaryRow {
		cell, _ := excelize.CoordinatesToCellName(i+1, 4)
		f.SetCellValue(sheet, cell, val)
	}

	// Add another valid transaction
	row2 := []interface{}{"16/01/2024", "SALARY CREDIT - ACME CORP", "NEFT/789012", "16/01/2024", "", 50000.00, 500000.00}
	for i, val := range row2 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 5)
		f.SetCellValue(sheet, cell, val)
	}

	// Save to temp file
	tmpFile, err := os.CreateTemp("", "edges-*.xlsx")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	buf := bytes.NewBuffer(nil)
	err = f.Write(buf)
	require.NoError(t, err)
	tmpFile.Write(buf.Bytes())
	tmpFile.Seek(0, 0)

	parser := NewParser()
	transactions, err := parser.ParseXLSX(tmpFile)

	require.NoError(t, err)
	// Should only have 2 transactions (skipped empty row and summary row)
	assert.Len(t, transactions, 2)
	assert.Equal(t, "AWS SERVICES", transactions[0].Description)
	assert.Equal(t, "SALARY CREDIT - ACME CORP", transactions[1].Description)
}

// ============================
// ParseFile Tests (Unified Interface)
// ============================

func TestParseFile_CSV_RoutesToParseCSV(t *testing.T) {
	file, err := os.Open("../../testdata/hdfc_sample.csv")
	require.NoError(t, err)
	defer file.Close()

	parser := NewParser()
	transactions, err := parser.ParseFile(file, "hdfc_sample.csv")

	require.NoError(t, err)
	assert.Len(t, transactions, 10)
	assert.Equal(t, "AWS SERVICES", transactions[0].Description)
}

func TestParseFile_XLSX_RoutesToParseXLSX(t *testing.T) {
	file, err := os.Open("../../testdata/hdfc_sample.xlsx")
	require.NoError(t, err)
	defer file.Close()

	parser := NewParser()
	transactions, err := parser.ParseFile(file, "hdfc_sample.xlsx")

	require.NoError(t, err)
	assert.Len(t, transactions, 10)
	assert.Equal(t, "AWS SERVICES", transactions[0].Description)
}

func TestParseFile_XLS_RoutesToParseXLSX(t *testing.T) {
	// .xls extension should also route to XLSX parser (excelize supports both)
	file, err := os.Open("../../testdata/hdfc_sample.xlsx")
	require.NoError(t, err)
	defer file.Close()

	parser := NewParser()
	transactions, err := parser.ParseFile(file, "hdfc_sample.xls")

	require.NoError(t, err)
	assert.Len(t, transactions, 10)
}

func TestParseFile_UnsupportedExtension_ReturnsError(t *testing.T) {
	file, err := os.Open("../../testdata/hdfc_sample.csv")
	require.NoError(t, err)
	defer file.Close()

	parser := NewParser()
	_, err = parser.ParseFile(file, "document.docx")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported file type")
	assert.Contains(t, err.Error(), ".docx")
}

func TestParseFile_NoExtension_ReturnsError(t *testing.T) {
	file, err := os.Open("../../testdata/hdfc_sample.csv")
	require.NoError(t, err)
	defer file.Close()

	parser := NewParser()
	_, err = parser.ParseFile(file, "document")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported file type")
}

func TestParseFile_CaseInsensitiveExtension(t *testing.T) {
	file, err := os.Open("../../testdata/hdfc_sample.csv")
	require.NoError(t, err)
	defer file.Close()

	parser := NewParser()
	transactions, err := parser.ParseFile(file, "hdfc_sample.CSV")

	require.NoError(t, err)
	assert.Len(t, transactions, 10)
}

// ============================
// ParsePDF Tests (HTTP Client Integration)
// ============================

func TestParsePDF_Success_HDFCFormat(t *testing.T) {
	// Mock Python microservice response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/parse", r.URL.Path)

		// Return mock HDFC transactions
		response := PDFParserResponse{
			Rows: [][]string{
				{"Date", "Narration", "Chq./Ref.No.", "Value Dt", "Withdrawal Amt.", "Deposit Amt.", "Closing Balance"},
				{"15/01/2024", "AWS SERVICES", "UPI/123456", "15/01/2024", "3500.00", "", "450000.00"},
				{"16/01/2024", "SALARY CREDIT - ACME CORP", "NEFT/789012", "16/01/2024", "", "50000.00", "500000.00"},
			},
			PagesProcessed: 1,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// Create parser with mock server URL
	parser := NewParserWithPDFClient(mockServer.URL)

	// Test PDF parsing
	pdfContent := strings.NewReader("mock pdf content")
	transactions, err := parser.ParsePDF(pdfContent)

	require.NoError(t, err)
	assert.Len(t, transactions, 2)
	assert.Equal(t, "AWS SERVICES", transactions[0].Description)
	assert.Equal(t, -3500.0, transactions[0].Amount)
	assert.Equal(t, "debit", transactions[0].TxnType)
	assert.Equal(t, "SALARY CREDIT - ACME CORP", transactions[1].Description)
	assert.Equal(t, 50000.0, transactions[1].Amount)
	assert.Equal(t, "credit", transactions[1].TxnType)
}

func TestParsePDF_Success_ICICIFormat(t *testing.T) {
	// Mock Python microservice response with ICICI format
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := PDFParserResponse{
			Rows: [][]string{
				{"Value Date", "Transaction Date", "Cheque Number", "Transaction Remarks", "Withdrawal Amount (INR)", "Deposit Amount (INR)", "Balance (INR)"},
				{"15/01/2024", "15/01/2024", "", "PAYMENT TO AWS SERVICES", "3500.00", "", "450000.00"},
				{"16/01/2024", "16/01/2024", "", "SALARY CREDIT FROM ACME CORP", "", "50000.00", "500000.00"},
			},
			PagesProcessed: 1,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	parser := NewParserWithPDFClient(mockServer.URL)
	pdfContent := strings.NewReader("mock pdf content")
	transactions, err := parser.ParsePDF(pdfContent)

	require.NoError(t, err)
	assert.Len(t, transactions, 2)
	assert.Equal(t, "PAYMENT TO AWS SERVICES", transactions[0].Description)
	assert.Equal(t, -3500.0, transactions[0].Amount)
}

func TestParsePDF_EmptyResponse(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := PDFParserResponse{
			Rows:           [][]string{},
			PagesProcessed: 0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	parser := NewParserWithPDFClient(mockServer.URL)
	pdfContent := strings.NewReader("mock pdf content")
	_, err := parser.ParsePDF(pdfContent)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no rows returned")
}

func TestParsePDF_ServiceUnavailable(t *testing.T) {
	// Use invalid URL to simulate service unavailable
	parser := NewParserWithPDFClient("http://localhost:9999")
	pdfContent := strings.NewReader("mock pdf content")
	_, err := parser.ParsePDF(pdfContent)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to call PDF parser service")
}

func TestParsePDF_InvalidJSON(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer mockServer.Close()

	parser := NewParserWithPDFClient(mockServer.URL)
	pdfContent := strings.NewReader("mock pdf content")
	_, err := parser.ParsePDF(pdfContent)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode response")
}

func TestParsePDF_HTTPError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer mockServer.Close()

	parser := NewParserWithPDFClient(mockServer.URL)
	pdfContent := strings.NewReader("mock pdf content")
	_, err := parser.ParsePDF(pdfContent)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PDF parser service returned error")
}

func TestParsePDF_UnknownBankFormat(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := PDFParserResponse{
			Rows: [][]string{
				{"Unknown", "Headers", "That", "Dont", "Match"},
				{"some", "random", "data", "here", "ok"},
			},
			PagesProcessed: 1,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	parser := NewParserWithPDFClient(mockServer.URL)
	pdfContent := strings.NewReader("mock pdf content")
	_, err := parser.ParsePDF(pdfContent)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown bank format")
}

func TestParseFile_PDF_RoutesToParsePDF(t *testing.T) {
	// Mock Python microservice
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := PDFParserResponse{
			Rows: [][]string{
				{"Date", "Narration", "Chq./Ref.No.", "Value Dt", "Withdrawal Amt.", "Deposit Amt.", "Closing Balance"},
				{"15/01/2024", "AWS SERVICES", "UPI/123456", "15/01/2024", "3500.00", "", "450000.00"},
			},
			PagesProcessed: 1,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	parser := NewParserWithPDFClient(mockServer.URL)
	pdfContent := strings.NewReader("mock pdf content")
	transactions, err := parser.ParseFile(pdfContent, "statement.pdf")

	require.NoError(t, err)
	assert.Len(t, transactions, 1)
	assert.Equal(t, "AWS SERVICES", transactions[0].Description)
}
