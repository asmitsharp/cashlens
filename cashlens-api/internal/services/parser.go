package services

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ashmitsharp/cashlens-api/internal/models"
	"github.com/xuri/excelize/v2"
)

// PDFParserResponse represents the response from the Python PDF parser microservice
type PDFParserResponse struct {
	Rows           [][]string `json:"rows"`
	PagesProcessed int        `json:"pages_processed"`
}

// Parser handles CSV/XLSX/PDF parsing for multiple bank formats
type Parser struct {
	bankSchemas   map[string]models.BankSchema
	pdfServiceURL string
	httpClient    *http.Client
}

// NewParser creates a new parser instance with predefined bank schemas
func NewParser() *Parser {
	return NewParserWithPDFClient("http://localhost:5000")
}

// NewParserWithPDFClient creates a parser with a custom PDF service URL (useful for testing)
func NewParserWithPDFClient(pdfServiceURL string) *Parser {
	return &Parser{
		bankSchemas: map[string]models.BankSchema{
			"HDFC": {
				BankName:           "HDFC",
				DateColumn:         "Date",
				DescriptionColumn:  "Narration",
				DebitColumn:        "Withdrawal Amt.",
				CreditColumn:       "Deposit Amt.",
				HasSeparateAmounts: true,
			},
			"ICICI": {
				BankName:           "ICICI",
				DateColumn:         "Transaction Date",
				DescriptionColumn:  "Transaction Remarks",
				DebitColumn:        "Withdrawal Amount (INR)",
				CreditColumn:       "Deposit Amount (INR)",
				HasSeparateAmounts: true,
			},
			"SBI": {
				BankName:           "SBI",
				DateColumn:         "Txn Date",
				DescriptionColumn:  "Description",
				DebitColumn:        "Debit",
				CreditColumn:       "Credit",
				HasSeparateAmounts: true,
			},
			"Axis": {
				BankName:           "Axis",
				DateColumn:         "Transaction Date",
				DescriptionColumn:  "Particulars",
				AmountColumn:       "Amount",
				DrCrColumn:         "Dr/Cr",
				HasSeparateAmounts: false,
			},
			"Kotak": {
				BankName:           "Kotak",
				DateColumn:         "Date",
				DescriptionColumn:  "Description",
				DebitColumn:        "Debit",
				CreditColumn:       "Credit",
				HasSeparateAmounts: true,
			},
		},
		pdfServiceURL: pdfServiceURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DetectBank detects the bank from CSV headers
func DetectBank(headers []string) string {
	headerSet := make(map[string]bool)
	for _, h := range headers {
		headerSet[strings.ToLower(strings.TrimSpace(h))] = true
	}

	// HDFC detection
	if headerSet["narration"] && headerSet["withdrawal amt."] {
		return "HDFC"
	}

	// ICICI detection
	if headerSet["transaction remarks"] && headerSet["withdrawal amount (inr)"] {
		return "ICICI"
	}

	// SBI detection
	if headerSet["txn date"] && headerSet["description"] {
		return "SBI"
	}

	// Axis detection
	if headerSet["particulars"] && headerSet["dr/cr"] {
		return "Axis"
	}

	// Kotak detection (most generic, check last)
	if headerSet["date"] && headerSet["debit"] && headerSet["credit"] && headerSet["description"] {
		return "Kotak"
	}

	return "UNKNOWN"
}

// ParseDate parses date strings in multiple formats
func ParseDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)

	dateFormats := []string{
		"02/01/2006",   // DD/MM/YYYY (HDFC, ICICI, Kotak)
		"2006-01-02",   // YYYY-MM-DD (ISO)
		"02-Jan-2006",  // DD-MMM-YYYY (SBI)
		"02-01-2006",   // DD-MM-YYYY
		"02/01/06",     // DD/MM/YY
		"Jan 02, 2006", // MMM DD, YYYY
	}

	for _, format := range dateFormats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// ParseAmount parses amount strings, handling currency symbols and commas
func ParseAmount(amountStr string) (float64, error) {
	// Remove currency symbols and commas
	cleaned := strings.ReplaceAll(amountStr, "₹", "")
	cleaned = strings.ReplaceAll(cleaned, "Rs.", "")
	cleaned = strings.ReplaceAll(cleaned, "Rs", "")
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	cleaned = strings.TrimSpace(cleaned)

	// Handle empty amounts
	if cleaned == "" || cleaned == "-" {
		return 0, nil
	}

	// Parse float
	amount, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount: %s", amountStr)
	}

	return amount, nil
}

// ParseCSV parses a CSV file and returns a list of transactions
func (p *Parser) ParseCSV(file io.Reader) ([]models.ParsedTransaction, error) {
	reader := csv.NewReader(file)

	// Read header row
	headers, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("empty file")
		}
		return nil, fmt.Errorf("failed to read headers: %w", err)
	}

	// Read all data rows
	var dataRows [][]string
	rowNum := 1 // Start from 1 since we already read headers

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading row %d: %w", rowNum, err)
		}
		rowNum++
		dataRows = append(dataRows, row)
	}

	// Use common parsing logic
	return p.parseRows(headers, dataRows)
}

// parseRow parses a single CSV row into a ParsedTransaction
func (p *Parser) parseRow(row []string, headerIndex map[string]int, schema models.BankSchema) (models.ParsedTransaction, error) {
	var txn models.ParsedTransaction

	// Parse date
	dateIdx, ok := headerIndex[schema.DateColumn]
	if !ok {
		return txn, fmt.Errorf("date column '%s' not found", schema.DateColumn)
	}
	date, err := ParseDate(row[dateIdx])
	if err != nil {
		return txn, fmt.Errorf("failed to parse date: %w", err)
	}
	txn.TxnDate = date

	// Parse description
	descIdx, ok := headerIndex[schema.DescriptionColumn]
	if !ok {
		return txn, fmt.Errorf("description column '%s' not found", schema.DescriptionColumn)
	}
	txn.Description = strings.TrimSpace(row[descIdx])

	// Parse amount based on schema type
	if schema.HasSeparateAmounts {
		// Banks with separate debit/credit columns (HDFC, ICICI, SBI, Kotak)
		debitIdx := headerIndex[schema.DebitColumn]
		creditIdx := headerIndex[schema.CreditColumn]

		debit, _ := ParseAmount(row[debitIdx])
		credit, _ := ParseAmount(row[creditIdx])

		if debit > 0 {
			txn.Amount = -debit // Negative for debit
			txn.TxnType = "debit"
		} else if credit > 0 {
			txn.Amount = credit // Positive for credit
			txn.TxnType = "credit"
		} else {
			return txn, fmt.Errorf("both debit and credit are zero")
		}
	} else {
		// Banks with single amount column and Dr/Cr indicator (Axis)
		amountIdx := headerIndex[schema.AmountColumn]
		drCrIdx := headerIndex[schema.DrCrColumn]

		amount, err := ParseAmount(row[amountIdx])
		if err != nil {
			return txn, fmt.Errorf("failed to parse amount: %w", err)
		}

		drCr := strings.ToLower(strings.TrimSpace(row[drCrIdx]))
		if drCr == "dr" {
			txn.Amount = -amount
			txn.TxnType = "debit"
		} else if drCr == "cr" {
			txn.Amount = amount
			txn.TxnType = "credit"
		} else {
			return txn, fmt.Errorf("invalid Dr/Cr indicator: %s", row[drCrIdx])
		}
	}

	// Store raw data for debugging
	txn.RawData = strings.Join(row, ",")

	return txn, nil
}

// isEmptyRow checks if all fields in a row are empty
func isEmptyRow(row []string) bool {
	for _, field := range row {
		if strings.TrimSpace(field) != "" {
			return false
		}
	}
	return true
}

// isSummaryRow checks if a row is a summary row
func isSummaryRow(row []string) bool {
	if len(row) == 0 {
		return false
	}

	firstField := strings.ToLower(strings.TrimSpace(row[0]))
	summaryKeywords := []string{"total", "summary", "opening balance", "closing balance"}

	for _, keyword := range summaryKeywords {
		if strings.Contains(firstField, keyword) {
			return true
		}
	}

	return false
}

// ParseXLSX parses an XLSX file and returns a list of transactions
func (p *Parser) ParseXLSX(file io.Reader) ([]models.ParsedTransaction, error) {
	// Read the XLSX file into memory
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Open the XLSX file
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to open XLSX file: %w", err)
	}
	defer f.Close()

	// Get the first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in XLSX file")
	}
	sheetName := sheets[0]

	// Read all rows from the first sheet
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("empty file")
	}

	// Extract headers and data rows
	headers := rows[0]
	dataRows := rows[1:]

	// Use common parsing logic
	return p.parseRows(headers, dataRows)
}

// parseRows is a common function that processes headers and data rows
func (p *Parser) parseRows(headers []string, dataRows [][]string) ([]models.ParsedTransaction, error) {
	// Detect bank
	bankName := DetectBank(headers)
	if bankName == "UNKNOWN" {
		return nil, fmt.Errorf("unknown bank format")
	}

	schema := p.bankSchemas[bankName]

	// Create header index map
	headerIndex := make(map[string]int)
	for i, h := range headers {
		headerIndex[strings.TrimSpace(h)] = i
	}

	// Parse data rows
	var transactions []models.ParsedTransaction
	for rowNum, row := range dataRows {
		// Skip empty rows
		if isEmptyRow(row) {
			continue
		}

		// Skip summary rows
		if isSummaryRow(row) {
			continue
		}

		// Parse transaction
		txn, err := p.parseRow(row, headerIndex, schema)
		if err != nil {
			// Log error but continue parsing
			fmt.Printf("Warning: skipping row %d: %v\n", rowNum+2, err)
			continue
		}

		transactions = append(transactions, txn)
	}

	return transactions, nil
}

// ParseFile is the unified entry point for parsing CSV, XLSX, or PDF files
func (p *Parser) ParseFile(file io.Reader, filename string) ([]models.ParsedTransaction, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".csv":
		return p.ParseCSV(file)
	case ".xlsx", ".xls":
		return p.ParseXLSX(file)
	case ".pdf":
		return p.ParsePDF(file)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

// ParsePDF calls the Python PDF parser microservice and returns parsed transactions
func (p *Parser) ParsePDF(file io.Reader) ([]models.ParsedTransaction, error) {
	// Read file content
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF file: %w", err)
	}

	// Create multipart form request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "statement.pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = part.Write(fileData)
	if err != nil {
		return nil, fmt.Errorf("failed to write file data: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Send POST request to PDF parser service
	url := p.pdfServiceURL + "/parse"
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call PDF parser service: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("PDF parser service returned error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Decode JSON response
	var pdfResponse PDFParserResponse
	if err := json.NewDecoder(resp.Body).Decode(&pdfResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Validate response
	if len(pdfResponse.Rows) == 0 {
		return nil, fmt.Errorf("no rows returned from PDF parser")
	}

	// Extract headers and data rows
	headers := pdfResponse.Rows[0]
	dataRows := pdfResponse.Rows[1:]

	// Use common parsing logic
	return p.parseRows(headers, dataRows)
}
