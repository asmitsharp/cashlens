# Bank Statement Format Research - Indian Banks

**Research Date:** 2025-11-05
**Purpose:** Multi-format parser implementation (Day 2+)
**Scope:** PDF, CSV, XLSX formats for 5 major Indian banks

---

## Executive Summary

Indian banks provide statements in **three primary formats**:

1. **PDF** (most common, password-protected, tabular data)
2. **CSV** (available via export, direct download from some banks)
3. **Excel/XLSX** (available via net banking download)

**Key Findings:**

- Most banks provide **PDF as the default format**
- CSV/Excel exports are available through net banking portals
- Each bank has **unique column structures** and naming conventions
- PDF parsing requires OCR/table extraction libraries (Camelot, pdfplumber, Tabula)
- CSV/XLSX formats are more reliable for automated parsing

---

## Format Availability by Bank

| Bank  | PDF | CSV | XLSX | Direct Download | Notes                              |
| ----- | --- | --- | ---- | --------------- | ---------------------------------- |
| HDFC  | ✅  | ✅  | ✅   | Via net banking | CSV available via export           |
| ICICI | ✅  | ✅  | ✅   | Via net banking | CSV from statement download        |
| SBI   | ✅  | ❌  | ✅   | Via net banking | Primarily PDF/Excel                |
| Axis  | ✅  | ✅  | ✅   | Via net banking | CSV export available               |
| Kotak | ✅  | ✅  | ✅   | Via net banking | Multiple format support            |

---

## Column Structures by Bank

### 1. HDFC Bank

**CSV/Excel Columns:**

```
Date | Narration | Chq./Ref.No. | Value Dt | Withdrawal Amt. | Deposit Amt. | Closing Balance
```

**Alternative Format:**

```
Date | Narration | Category | Withdrawal_Amount | Deposit_Amount
```

**Key Characteristics:**

- Date format: DD/MM/YYYY
- Separate columns for withdrawals and deposits
- Narration contains transaction description
- Category field (when available) has pre-categorization
- Closing balance after each transaction

**Sample Row:**

```csv
15/01/2024,AWS SERVICES,UPI/123456,15/01/2024,3500.00,,450000.00
```

### 2. ICICI Bank

**CSV Columns:**

```
Value Date | Transaction Date | Cheque Number | Transaction Remarks | Withdrawal Amount (INR) | Deposit Amount (INR) | Balance (INR)
```

**Key Characteristics:**

- Two date columns: Value Date and Transaction Date
- Transaction Remarks = description
- Amount columns explicitly labeled with currency (INR)
- Balance shown after each transaction

**Sample Row:**

```csv
15/01/2024,15/01/2024,,PAYMENT TO AWS SERVICES,3500.00,,450000.00
```

### 3. SBI (State Bank of India)

**Excel Columns:**

```
Txn Date | Description | Ref No./Cheque No. | Value Date | Debit | Credit | Balance
```

**Alternative Format:**

```
Txn Date | Description | Debit | Credit | Balance
```

**Key Characteristics:**

- "Txn Date" instead of "Date"
- Uses "Debit/Credit" instead of "Withdrawal/Deposit"
- Description field contains transaction details
- Date format: DD/MM/YYYY or DD-MMM-YYYY

**Sample Row:**

```csv
15-Jan-2024,PAYMENT TO AWS SERVICES,,15-Jan-2024,3500.00,,450000.00
```

### 4. Axis Bank

**CSV Columns:**

```
Transaction Date | Particulars | Cheque No. | Dr/Cr | Amount | Balance
```

**Key Characteristics:**

- "Particulars" = transaction description
- Single "Amount" column with "Dr/Cr" indicator
- Compact format compared to other banks

**Sample Row:**

```csv
15/01/2024,PAYMENT TO AWS SERVICES,,Dr,3500.00,450000.00
```

### 5. Kotak Mahindra Bank

**CSV Columns:**

```
Date | Description | Ref No. | Debit | Credit | Balance
```

**Key Characteristics:**

- Simple column structure
- Uses Debit/Credit like SBI
- Description field for transaction details

**Sample Row:**

```csv
15/01/2024,PAYMENT TO AWS SERVICES,,3500.00,,450000.00
```

---

## PDF Format Analysis

### Common PDF Structure

Indian bank statements in PDF typically follow this structure:

```
┌─────────────────────────────────────────────┐
│ BANK LOGO                                   │
│                                             │
│ Account Statement                           │
│                                             │
│ Account Holder: John Doe                    │
│ Account Number: XXXX-XXXX-1234             │
│ Statement Period: 01-Jan-2024 to 31-Jan-24 │
│                                             │
│ ┌─────────────────────────────────────────┐ │
│ │ Date │ Description │ Debit │ Credit │Bal│ │
│ ├─────────────────────────────────────────┤ │
│ │15/01 │AWS SERVICES │3500.00│      │450K│ │
│ │16/01 │SALARY       │       │50000│500K│ │
│ │...   │...          │...    │...   │... │ │
│ └─────────────────────────────────────────┘ │
│                                             │
│ Summary: Total Debit: ₹45,000             │
│          Total Credit: ₹1,20,000           │
└─────────────────────────────────────────────┘
```

### PDF Password Protection

- **SBI**: Password format typically: `DDMMYYYY` (account holder's DOB)
- **HDFC**: Password format: Last 4 digits of account + DOB
- **ICICI**: Password format: Account holder's DOB or custom
- **Axis**: Password format: DOB or mobile number
- **Kotak**: Password format: DOB or custom

### PDF Table Extraction Challenges

1. **Multi-line descriptions**: Transactions may span multiple rows
2. **Page breaks**: Tables split across pages
3. **Formatting variations**: Different PDF generators create different structures
4. **Scanned PDFs**: OCR required for image-based PDFs
5. **Currency symbols**: ₹, Rs., INR need to be stripped

---

## Python Libraries for PDF Parsing

### 1. **Camelot-py** (Recommended for Indian Banks)

```python
import camelot

# Extract tables from PDF
tables = camelot.read_pdf(
    'hdfc_statement.pdf',
    flavor='stream',  # Use 'stream' for tables without borders
    pages='all',
    password='password123'
)

# Convert to pandas DataFrame
df = tables[0].df
```

**Pros:**

- Excellent for tabular data extraction
- Supports both 'lattice' (bordered tables) and 'stream' (space-separated) modes
- Direct pandas DataFrame output
- HDFC bank statement parsing confirmed working

**Cons:**

- Requires dependencies: ghostscript, tkinter
- May struggle with complex layouts

### 2. **pdfplumber**

```python
import pdfplumber

with pdfplumber.open('statement.pdf', password='password') as pdf:
    for page in pdf.pages:
        tables = page.extract_tables()
        for table in tables:
            # Process table rows
            for row in table:
                print(row)
```

**Pros:**

- Robust for complex table structures
- Better handling of multi-line cells
- Detailed control over extraction

**Cons:**

- More manual processing required
- Slower than Camelot for simple tables

### 3. **Tabula-py**

```python
import tabula

# Extract all tables from PDF
df = tabula.read_pdf(
    'statement.pdf',
    pages='all',
    password='password123'
)
```

**Pros:**

- Built on Java-based Tabula
- Good for standardized table formats
- Fast extraction

**Cons:**

- Requires Java runtime
- Less accurate than pdfplumber for complex layouts

### 4. **PyPDF2** (Text Extraction Only)

```python
from PyPDF2 import PdfReader

reader = PdfReader('statement.pdf')
for page in reader.pages:
    text = page.extract_text()
    # Requires manual parsing with regex
```

**Pros:**

- Lightweight
- Fast text extraction

**Cons:**

- No table structure preservation
- Requires extensive regex parsing

---

## Go Libraries for Multi-Format Parsing

### CSV Parsing (Go)

```go
import (
    "encoding/csv"
    "os"
)

func ParseCSV(filePath string) ([][]string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        return nil, err
    }

    return records, nil
}
```

### XLSX Parsing (Go)

**Recommended Library: `excelize`**

```go
import "github.com/xuri/excelize/v2"

func ParseXLSX(filePath string) ([][]string, error) {
    f, err := excelize.OpenFile(filePath)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    // Get all rows from first sheet
    rows, err := f.GetRows("Sheet1")
    if err != nil {
        return nil, err
    }

    return rows, nil
}
```

### PDF Parsing (Go)

**Option 1: Use Python microservice** (Recommended)

```go
// Call Python service via HTTP
func ParsePDF(filePath string) ([]Transaction, error) {
    // Upload PDF to Python parser service
    resp, err := http.Post(
        "http://pdf-parser:5000/parse",
        "application/pdf",
        file,
    )
    // Parse JSON response
}
```

**Option 2: Use `unidoc/unipdf`** (Go-native, limited table extraction)

```go
import "github.com/unidoc/unipdf/v3/extractor"

func ExtractPDFText(filePath string) (string, error) {
    f, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer f.Close()

    pdfReader, err := pdf.NewPdfReader(f)
    if err != nil {
        return "", err
    }

    text, err := extractor.ExtractText(pdfReader.GetPage(1))
    return text, err
}
```

**Note:** For production-grade PDF table extraction, a Python microservice with Camelot is recommended.

---

## Implementation Strategy

### Phase 1: CSV/XLSX Only (Day 2 - Current Plan)

**Focus:** Implement CSV parser for 5 banks

```go
// internal/services/parser.go
type BankSchema struct {
    DateColumn        string
    DescriptionColumn string
    DebitColumn       string
    CreditColumn      string
    AmountColumn      string
    DrCrColumn        string
}

var BankSchemas = map[string]BankSchema{
    "HDFC": {
        DateColumn:        "Date",
        DescriptionColumn: "Narration",
        DebitColumn:       "Withdrawal Amt.",
        CreditColumn:      "Deposit Amt.",
    },
    "ICICI": {
        DateColumn:        "Transaction Date",
        DescriptionColumn: "Transaction Remarks",
        DebitColumn:       "Withdrawal Amount (INR)",
        CreditColumn:      "Deposit Amount (INR)",
    },
    "SBI": {
        DateColumn:        "Txn Date",
        DescriptionColumn: "Description",
        DebitColumn:       "Debit",
        CreditColumn:      "Credit",
    },
    // ... Axis, Kotak
}
```

### Phase 2: XLSX Support (Day 3 Enhancement)

**Add:** Excel file parsing using `excelize`

```go
func DetectFileType(filename string) string {
    ext := filepath.Ext(filename)
    switch strings.ToLower(ext) {
    case ".csv":
        return "csv"
    case ".xlsx", ".xls":
        return "excel"
    case ".pdf":
        return "pdf"
    default:
        return "unknown"
    }
}
```

### Phase 3: PDF Support (Post-MVP)

**Options:**

1. **Python Microservice** (Recommended)
   - Dockerized Flask/FastAPI service
   - Uses Camelot/pdfplumber
   - Exposes `/parse` endpoint
   - Returns JSON
2. **Third-party API** (Alternative)
   - Docparser.com
   - PDFTables.com
   - FormX.ai (Indian market focus)

---

## Date Format Parsing Strategy

Indian bank statements use multiple date formats:

```go
var DateFormats = []string{
    "02/01/2006",      // DD/MM/YYYY (HDFC, ICICI, Kotak)
    "2006-01-02",      // YYYY-MM-DD (ISO standard)
    "02-Jan-2006",     // DD-MMM-YYYY (SBI)
    "02-01-06",        // DD-MM-YY
    "Jan 02, 2006",    // MMM DD, YYYY
}

func ParseDate(dateStr string) (time.Time, error) {
    for _, format := range DateFormats {
        t, err := time.Parse(format, dateStr)
        if err == nil {
            return t, nil
        }
    }
    return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}
```

---

## Amount Parsing Strategy

Handle various amount representations:

```go
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
```

---

## Schema Detection Algorithm

Auto-detect bank based on column headers:

```go
func DetectBank(headers []string) string {
    headerSet := make(map[string]bool)
    for _, h := range headers {
        headerSet[strings.ToLower(h)] = true
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

    // Kotak detection (simple format)
    if headerSet["date"] && headerSet["debit"] && headerSet["credit"] {
        return "Kotak"
    }

    return "UNKNOWN"
}
```

---

## Edge Cases to Handle

### 1. Multi-line Descriptions

Some banks split long descriptions across multiple rows:

```csv
15/01/2024,PAYMENT TO AWS
,SERVICES INVOICE #12345,3500.00,,
```

**Solution:** Concatenate rows without dates

### 2. Summary Rows

Statements often include summary rows:

```csv
Total Debit,,,45000.00,,
Total Credit,,,,120000,
```

**Solution:** Skip rows with "Total", "Summary", "Opening Balance"

### 3. Empty Rows

```csv
15/01/2024,AWS SERVICES,3500.00,,

16/01/2024,SALARY,,50000,
```

**Solution:** Filter rows with all empty values

### 4. Currency Conversion

International transactions may show currency:

```csv
15/01/2024,STRIPE USD 50.00 @ 83.50,4175.00,,
```

**Solution:** Extract final INR amount only

### 5. Reversed Transactions

Refunds or reversals may appear:

```csv
15/01/2024,AWS SERVICES,3500.00,,
16/01/2024,REVERSAL - AWS SERVICES,,3500.00
```

**Solution:** Mark as separate transactions, let categorizer handle

---

## Testing Strategy

### Test Data Requirements

For each bank, create:

1. **Minimal CSV** (5-10 rows) - Schema validation
2. **Realistic CSV** (100+ rows) - Performance testing
3. **Edge case CSV** - Multi-line, empty rows, summary rows
4. **Real anonymized CSV** - Accuracy testing

### Test Fixtures Location

```
cashlens-api/
└── testdata/
    ├── hdfc/
    │   ├── minimal.csv
    │   ├── realistic.csv
    │   └── edge_cases.csv
    ├── icici/
    │   ├── minimal.csv
    │   └── ...
    ├── sbi/
    ├── axis/
    └── kotak/
```

### Unit Test Template

```go
func TestParseCSV_HDFC_Minimal(t *testing.T) {
    file, err := os.Open("testdata/hdfc/minimal.csv")
    require.NoError(t, err)
    defer file.Close()

    parser := NewParser()
    txns, err := parser.ParseCSV(file)

    assert.NoError(t, err)
    assert.Len(t, txns, 5)

    // Validate first transaction
    assert.Equal(t, "AWS SERVICES", txns[0].Description)
    assert.Equal(t, -3500.0, txns[0].Amount)
    assert.Equal(t, "debit", txns[0].TxnType)
}
```

---

## Recommendations for Day 2 Implementation

### Immediate Focus (CSV Only)

1. ✅ Implement CSV parser with schema detection
2. ✅ Support all 5 banks (HDFC, ICICI, SBI, Axis, Kotak)
3. ✅ Handle date/amount parsing edge cases
4. ✅ Write comprehensive tests with real CSV samples
5. ✅ Achieve 100% test coverage for parser

### Day 3 Enhancement (XLSX)

1. Add `excelize` dependency
2. Extend parser to handle XLSX files
3. Test with Excel files from all 5 banks

### Post-MVP (PDF Support)

1. **Option A:** Build Python microservice
   - Docker container with Camelot
   - Flask API for PDF parsing
   - Deploy alongside Go API
2. **Option B:** Integrate third-party API
   - FormX.ai or Docparser
   - Add API key to environment config

---

## External Resources

### Python Libraries Documentation

- [Camelot Documentation](https://camelot-py.readthedocs.io/)
- [pdfplumber Documentation](https://github.com/jsvine/pdfplumber)
- [Tabula-py Documentation](https://tabula-py.readthedocs.io/)

### Go Libraries Documentation

- [excelize (XLSX)](https://xuri.me/excelize/)
- [unipdf (PDF)](https://github.com/unidoc/unipdf)

### Third-party Services

- [FormX.ai](https://formx.ai/) - Indian bank statement OCR
- [Docparser](https://docparser.com/) - Generic PDF parsing
- [PDFTables](https://pdftables.com/) - Table extraction

### Sample Code Repositories

- [HDFC Statement Parser (Python)](https://dev.to/vishwaraja_pathivishwa/building-a-pdf-parser-for-hdfc-bank-statements-from-165-pages-to-csv-in-minutes-34c6)
- [Beancount India Importers](https://github.com/prabusw/beancount-importers-india)

---

## Next Steps

1. ✅ **Research complete** - This document
2. 🔄 **Day 2 Implementation** - CSV parser for 5 banks
3. 📋 **Create test fixtures** - Minimal CSV files for each bank
4. 🧪 **Write TDD tests** - Using `tdd-orchestrator` agent
5. 🚀 **Implement parser** - Following schema detection algorithm

---

**Document Status:** Complete
**Last Updated:** 2025-11-05
**Reviewed By:** Research Agent
**Next Review:** After Day 2 implementation
