package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/ashmitsharp/cashlens-api/internal/models"
)

// Parser handles CSV parsing for multiple bank formats
type Parser struct {
	bankSchemas map[string]models.BankSchema
}

// NewParser creates a new parser instance with predefined bank schemas
func NewParser() *Parser {
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
				BankName:          "Axis",
				DateColumn:        "Transaction Date",
				DescriptionColumn: "Particulars",
				AmountColumn:      "Amount",
				DrCrColumn:        "Dr/Cr",
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
		"02/01/2006",    // DD/MM/YYYY (HDFC, ICICI, Kotak)
		"2006-01-02",    // YYYY-MM-DD (ISO)
		"02-Jan-2006",   // DD-MMM-YYYY (SBI)
		"02-01-2006",    // DD-MM-YYYY
		"02/01/06",      // DD/MM/YY
		"Jan 02, 2006",  // MMM DD, YYYY
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

	// Parse rows
	var transactions []models.ParsedTransaction
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
			fmt.Printf("Warning: skipping row %d: %v\n", rowNum, err)
			continue
		}

		transactions = append(transactions, txn)
	}

	return transactions, nil
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
