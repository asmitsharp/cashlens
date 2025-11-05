package services

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
