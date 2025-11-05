# CSV Parser - Code Walkthrough

**For:** New developers joining the Cashlens project
**Purpose:** Step-by-step explanation of the CSV parser implementation
**Difficulty:** Intermediate Go
**Time to Read:** 20 minutes

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Understanding the Models](#understanding-the-models)
3. [The Parser Struct](#the-parser-struct)
4. [Step 1: Creating a Parser](#step-1-creating-a-parser)
5. [Step 2: Reading the CSV](#step-2-reading-the-csv)
6. [Step 3: Detecting the Bank](#step-3-detecting-the-bank)
7. [Step 4: Parsing Rows](#step-4-parsing-rows)
8. [Step 5: Handling Amounts](#step-5-handling-amounts)
9. [Common Patterns](#common-patterns)
10. [Debugging Tips](#debugging-tips)

---

## Quick Start

Before diving into the code, here's how you use the parser:

```go
package main

import (
    "os"
    "fmt"
    "github.com/ashmitsharp/cashlens-api/internal/services"
)

func main() {
    // Open CSV file
    file, err := os.Open("bank_statement.csv")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    // Create parser
    parser := services.NewParser()

    // Parse the file
    transactions, err := parser.ParseCSV(file)
    if err != nil {
        panic(err)
    }

    // Use transactions
    for _, txn := range transactions {
        fmt.Printf("%s: %s - %.2f\n",
            txn.TxnDate.Format("2006-01-02"),
            txn.Description,
            txn.Amount)
    }
}
```

**Output:**
```
2024-01-15: AWS SERVICES - -3500.00
2024-01-16: SALARY CREDIT - ACME CORP - 50000.00
2024-01-17: RAZORPAY PAYMENT GATEWAY - -2500.00
...
```

Now let's understand how this works internally.

---

## Understanding the Models

Before we parse anything, we need to understand the data structures.

### ParsedTransaction

This is what we produce after parsing a row:

```go
type ParsedTransaction struct {
    TxnDate     time.Time  `json:"txn_date"`     // When did it happen?
    Description string     `json:"description"`  // What was it for?
    Amount      float64    `json:"amount"`       // How much? (negative = debit)
    TxnType     string     `json:"txn_type"`     // "credit" or "debit"
    RawData     string     `json:"raw_data"`     // Original CSV row (for debugging)
}
```

**Key Design Decision: Why is Amount negative for debits?**

```go
// Option A: Negative for debits (what we use)
debit:  Amount = -3500.00
credit: Amount = 50000.00

balance += amount  // Simple math!

// Option B: Separate fields (more complex)
debit:  DebitAmount = 3500.00, CreditAmount = 0
credit: DebitAmount = 0, CreditAmount = 50000.00

balance = balance - debitAmount + creditAmount  // More complex
```

We chose Option A because:
1. Simpler balance calculations
2. Easier filtering (`WHERE amount < 0` for expenses)
3. Standard accounting convention

### BankSchema

This tells us where to find data in a specific bank's CSV:

```go
type BankSchema struct {
    BankName          string  // "HDFC", "ICICI", etc.
    DateColumn        string  // "Date", "Txn Date", etc.
    DescriptionColumn string  // "Narration", "Description", etc.
    DebitColumn       string  // "Withdrawal Amt.", "Debit", etc.
    CreditColumn      string  // "Deposit Amt.", "Credit", etc.
    AmountColumn      string  // For banks with single amount column
    DrCrColumn        string  // "Dr/Cr" indicator
    HasSeparateAmounts bool   // true if debit/credit are separate
}
```

**Example - HDFC Schema:**

```go
models.BankSchema{
    BankName:           "HDFC",
    DateColumn:         "Date",
    DescriptionColumn:  "Narration",
    DebitColumn:        "Withdrawal Amt.",
    CreditColumn:       "Deposit Amt.",
    HasSeparateAmounts: true,  // HDFC has separate columns
}
```

**Example - Axis Schema:**

```go
models.BankSchema{
    BankName:          "Axis",
    DateColumn:        "Transaction Date",
    DescriptionColumn: "Particulars",
    AmountColumn:      "Amount",      // Single amount column
    DrCrColumn:        "Dr/Cr",       // Indicator column
    HasSeparateAmounts: false,        // Uses Dr/Cr system
}
```

---

## The Parser Struct

The parser is simple - it just holds bank configurations:

```go
type Parser struct {
    bankSchemas map[string]models.BankSchema
}
```

Think of `bankSchemas` as a dictionary:

```go
{
    "HDFC":  { /* HDFC configuration */ },
    "ICICI": { /* ICICI configuration */ },
    "SBI":   { /* SBI configuration */ },
    "Axis":  { /* Axis configuration */ },
    "Kotak": { /* Kotak configuration */ },
}
```

**Why a map?**
- Fast lookup: O(1) access by bank name
- Easy to add new banks: just add another key
- Clean separation of concerns

---

## Step 1: Creating a Parser

```go
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
            // ... 4 more banks
        },
    }
}
```

**What's happening:**
1. Create a new `Parser` instance
2. Initialize it with 5 bank configurations
3. Return a pointer (efficient for passing around)

**Usage:**
```go
parser := NewParser()
```

---

## Step 2: Reading the CSV

Let's break down `ParseCSV` step by step:

```go
func (p *Parser) ParseCSV(file io.Reader) ([]models.ParsedTransaction, error) {
    // Step 2.1: Create CSV reader
    reader := csv.NewReader(file)
```

**What is `io.Reader`?**
- An interface - can be a file, network stream, string, etc.
- Flexible: works with `os.File`, `bytes.Buffer`, `strings.Reader`
- Testable: easy to mock in tests

```go
    // Step 2.2: Read headers
    headers, err := reader.Read()
    if err != nil {
        if err == io.EOF {
            return nil, fmt.Errorf("empty file")
        }
        return nil, fmt.Errorf("failed to read headers: %w", err)
    }
```

**Why check for EOF separately?**
- `io.EOF` means empty file - specific error message
- Other errors might be corrupted file, permission issues, etc.
- `%w` wraps the original error (Go 1.13+)

```go
    // Step 2.3: Detect which bank this is
    bankName := DetectBank(headers)
    if bankName == "UNKNOWN" {
        return nil, fmt.Errorf("unknown bank format")
    }
```

We'll cover `DetectBank` in the next section!

```go
    // Step 2.4: Get the schema for this bank
    schema := p.bankSchemas[bankName]
```

Now we know HOW to parse this CSV!

```go
    // Step 2.5: Build header index for fast lookup
    headerIndex := make(map[string]int)
    for i, h := range headers {
        headerIndex[strings.TrimSpace(h)] = i
    }
```

**What's headerIndex?**

Instead of searching for a column by name each time:
```go
// Slow: O(n) for each row
for i, h := range headers {
    if h == "Date" {
        dateIdx = i
        break
    }
}
```

We build a map once:
```go
// Fast: O(1) lookup
headerIndex = {
    "Date": 0,
    "Narration": 1,
    "Withdrawal Amt.": 4,
    "Deposit Amt.": 5,
}

dateIdx := headerIndex["Date"]  // Instant!
```

```go
    // Step 2.6: Parse each row
    var transactions []models.ParsedTransaction
    rowNum := 1  // Track row number for error messages

    for {
        row, err := reader.Read()
        if err == io.EOF {
            break  // Done reading
        }
        if err != nil {
            return nil, fmt.Errorf("error reading row %d: %w", rowNum, err)
        }

        rowNum++

        // Skip invalid rows
        if isEmptyRow(row) || isSummaryRow(row) {
            continue
        }

        // Parse this row
        txn, err := p.parseRow(row, headerIndex, schema)
        if err != nil {
            // Log warning but continue parsing
            fmt.Printf("Warning: skipping row %d: %v\n", rowNum, err)
            continue
        }

        transactions = append(transactions, txn)
    }

    return transactions, nil
}
```

**Key Pattern: Error Recovery**

Notice we don't stop parsing if one row fails:
```go
if err != nil {
    fmt.Printf("Warning: skipping row %d: %v\n", rowNum, err)
    continue  // Keep going!
}
```

**Why?**
- One bad row shouldn't kill the whole import
- User still gets 99 good transactions out of 100
- Warnings help debug issues

---

## Step 3: Detecting the Bank

This is where the magic happens:

```go
func DetectBank(headers []string) string {
    // Step 3.1: Normalize headers
    headerSet := make(map[string]bool)
    for _, h := range headers {
        headerSet[strings.ToLower(strings.TrimSpace(h))] = true
    }
```

**Why normalize?**
- Banks might use "Date" or "date" or "  Date  "
- `ToLower()` handles case differences
- `TrimSpace()` handles extra spaces
- Map allows fast lookups

```go
    // Step 3.2: Check for HDFC
    if headerSet["narration"] && headerSet["withdrawal amt."] {
        return "HDFC"
    }
```

**Why this combination?**
- "narration" is unique to HDFC
- "withdrawal amt." confirms it's HDFC format
- Both keywords reduce false positives

```go
    // Step 3.3: Check for ICICI
    if headerSet["transaction remarks"] && headerSet["withdrawal amount (inr)"] {
        return "ICICI"
    }

    // Step 3.4: Check for SBI
    if headerSet["txn date"] && headerSet["description"] {
        return "SBI"
    }

    // Step 3.5: Check for Axis
    if headerSet["particulars"] && headerSet["dr/cr"] {
        return "Axis"
    }

    // Step 3.6: Check for Kotak (most generic - check last!)
    if headerSet["date"] && headerSet["debit"] && headerSet["credit"] && headerSet["description"] {
        return "Kotak"
    }

    // Step 3.7: Unknown bank
    return "UNKNOWN"
}
```

**Why check Kotak last?**
- Kotak uses generic column names ("Date", "Debit", "Credit")
- Other banks might also have these columns
- Check specific banks first to avoid false matches

**Visual Example:**

```
CSV Headers: ["Date", "Narration", "Withdrawal Amt.", ...]

Step 1: Normalize
{"date": true, "narration": true, "withdrawal amt.": true, ...}

Step 2: Check HDFC
Has "narration"? ✅
Has "withdrawal amt."? ✅
→ Return "HDFC"
```

---

## Step 4: Parsing Rows

Now let's parse a single row:

```go
func (p *Parser) parseRow(row []string, headerIndex map[string]int, schema models.BankSchema) (models.ParsedTransaction, error) {
    var txn models.ParsedTransaction

    // Step 4.1: Parse date
    dateIdx, ok := headerIndex[schema.DateColumn]
    if !ok {
        return txn, fmt.Errorf("date column '%s' not found", schema.DateColumn)
    }

    date, err := ParseDate(row[dateIdx])
    if err != nil {
        return txn, fmt.Errorf("failed to parse date: %w", err)
    }
    txn.TxnDate = date
```

**Breaking it down:**

1. `headerIndex[schema.DateColumn]` - Get column position
   - For HDFC: `headerIndex["Date"]` → `0`
2. `row[dateIdx]` - Get the actual date string
   - `row[0]` → `"15/01/2024"`
3. `ParseDate(...)` - Parse into `time.Time`
4. Store in `txn.TxnDate`

```go
    // Step 4.2: Parse description
    descIdx, ok := headerIndex[schema.DescriptionColumn]
    if !ok {
        return txn, fmt.Errorf("description column '%s' not found", schema.DescriptionColumn)
    }
    txn.Description = strings.TrimSpace(row[descIdx])
```

Simple string extraction with whitespace trimming.

```go
    // Step 4.3: Parse amounts (two strategies!)
    if schema.HasSeparateAmounts {
        // Strategy A: Separate debit/credit columns (HDFC, ICICI, SBI, Kotak)
        debitIdx := headerIndex[schema.DebitColumn]
        creditIdx := headerIndex[schema.CreditColumn]

        debit, _ := ParseAmount(row[debitIdx])
        credit, _ := ParseAmount(row[creditIdx])

        if debit > 0 {
            txn.Amount = -debit  // Negative for money going out
            txn.TxnType = "debit"
        } else if credit > 0 {
            txn.Amount = credit  // Positive for money coming in
            txn.TxnType = "credit"
        } else {
            return txn, fmt.Errorf("both debit and credit are zero")
        }
    }
```

**Example - HDFC Row:**
```csv
15/01/2024,AWS SERVICES,...,3500.00,,
                           ^^debit ^^credit (empty)

debit  = ParseAmount("3500.00") → 3500.0
credit = ParseAmount("")        → 0.0

→ txn.Amount = -3500.0
→ txn.TxnType = "debit"
```

```go
    else {
        // Strategy B: Single amount + Dr/Cr indicator (Axis)
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
```

**Example - Axis Row:**
```csv
15/01/2024,AWS SERVICES,,Dr,3500.00,
                         ^^drCr ^^amount

amount = ParseAmount("3500.00") → 3500.0
drCr   = "dr"

→ txn.Amount = -3500.0
→ txn.TxnType = "debit"
```

```go
    // Step 4.4: Store raw data for debugging
    txn.RawData = strings.Join(row, ",")

    return txn, nil
}
```

**Why save RawData?**
- Helps debug parsing issues
- Audit trail for compliance
- Can re-parse if logic changes

---

## Step 5: Handling Amounts

The amount parser handles various formats:

```go
func ParseAmount(amountStr string) (float64, error) {
    // Step 5.1: Remove currency symbols
    cleaned := strings.ReplaceAll(amountStr, "₹", "")
    cleaned = strings.ReplaceAll(cleaned, "Rs.", "")
    cleaned = strings.ReplaceAll(cleaned, "Rs", "")

    // Step 5.2: Remove commas
    cleaned = strings.ReplaceAll(cleaned, ",", "")

    // Step 5.3: Trim whitespace
    cleaned = strings.TrimSpace(cleaned)

    // Step 5.4: Handle empty amounts
    if cleaned == "" || cleaned == "-" {
        return 0, nil
    }

    // Step 5.5: Parse to float64
    amount, err := strconv.ParseFloat(cleaned, 64)
    if err != nil {
        return 0, fmt.Errorf("invalid amount: %s", amountStr)
    }

    return amount, nil
}
```

**Visual Transformation:**

```
Input:  "₹1,50,000.00"
Step 1: "1,50,000.00"    (remove ₹)
Step 2: "150000.00"       (remove commas)
Step 3: "150000.00"       (trim - no change)
Step 5: 150000.0          (parse to float)
```

**Why handle commas?**

Indian numbering:
```
1,00,000  = 1 lakh (100,000)
10,00,000 = 10 lakhs (1,000,000)
```

Western numbering:
```
100,000   = 100 thousand
1,000,000 = 1 million
```

Banks use both formats!

---

## Common Patterns

### Pattern 1: Early Return on Error

```go
if err != nil {
    return txn, fmt.Errorf("failed: %w", err)
}
```

Instead of nested if-else, we return early. Makes code linear and easier to read.

### Pattern 2: Error Wrapping

```go
return nil, fmt.Errorf("failed to parse date: %w", err)
```

The `%w` preserves the original error. You can unwrap it later:

```go
if errors.Is(err, io.EOF) {
    // Handle EOF specifically
}
```

### Pattern 3: Map for Fast Lookup

```go
headerIndex := make(map[string]int)
for i, h := range headers {
    headerIndex[h] = i
}

// Later: O(1) lookup instead of O(n) search
idx := headerIndex["Date"]
```

### Pattern 4: Normalize Before Compare

```go
normalized := strings.ToLower(strings.TrimSpace(input))
if normalized == "dr" {
    // Handle debit
}
```

Handles "Dr", "DR", " dr ", etc.

---

## Debugging Tips

### Tip 1: Use the RawData Field

```go
fmt.Printf("Failed to parse: %s\n", txn.RawData)
```

Shows you the exact CSV row that failed.

### Tip 2: Add Debug Logging

```go
fmt.Printf("Parsing row %d: bank=%s, date=%s\n", rowNum, bankName, row[dateIdx])
```

### Tip 3: Test with Small Files

Create a 3-row CSV to isolate issues:

```csv
Date,Narration,Withdrawal Amt.,Deposit Amt.
15/01/2024,TEST1,100.00,
16/01/2024,TEST2,,200.00
```

### Tip 4: Check Your Test Fixtures

Look at `testdata/hdfc_sample.csv` for working examples.

### Tip 5: Use Table-Driven Tests

```go
tests := []struct {
    input    string
    expected float64
}{
    {"₹100.00", 100.0},
    {"1,50,000", 150000.0},
    {"", 0.0},
}

for _, tt := range tests {
    result, _ := ParseAmount(tt.input)
    if result != tt.expected {
        t.Errorf("ParseAmount(%q) = %v, want %v", tt.input, result, tt.expected)
    }
}
```

---

## Conclusion

You now understand:

1. ✅ How the parser struct works (map of bank schemas)
2. ✅ How CSV reading works (streaming with csv.Reader)
3. ✅ How bank detection works (keyword matching)
4. ✅ How row parsing works (two strategies for amounts)
5. ✅ How amount parsing works (cleaning + float conversion)

**Next Steps:**

1. Read the actual code in `internal/services/parser.go`
2. Run the tests: `go test -v ./internal/services`
3. Try modifying a test and see what breaks
4. Add a new bank format (great learning exercise!)

**Key Takeaways:**

- Error handling is crucial - don't crash on bad data
- Normalization prevents bugs (lowercase, trim)
- Maps are your friend for fast lookups
- Test with real data, not just happy paths
- Document edge cases in code comments

Happy coding! 🚀
