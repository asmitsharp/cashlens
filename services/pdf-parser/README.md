# PDF Parser Microservice

Python Flask service for extracting transaction tables from PDF bank statements.

## Features

- Supports multiple extraction methods (pdfplumber + camelot)
- Handles both text-based and scanned PDFs
- Automatic fallback between extraction methods
- Multi-page PDF support
- Header deduplication across pages
- 10MB file size limit

## API Endpoints

### Health Check
```bash
GET /health
```

Response:
```json
{
  "status": "ok",
  "service": "pdf-parser",
  "version": "1.0.0"
}
```

### Parse PDF
```bash
POST /parse
Content-Type: multipart/form-data
```

Request:
- Field: `file` (PDF file, max 10MB)

Response:
```json
{
  "rows": [
    ["Date", "Description", "Debit", "Credit"],
    ["01/01/2024", "AWS SERVICES", "3500.00", ""],
    ["02/01/2024", "SALARY CREDIT", "", "50000.00"]
  ],
  "pages_processed": 3,
  "method_used": "pdfplumber",
  "total_rows": 50
}
```

## Local Development

### Setup Virtual Environment
```bash
cd services/pdf-parser

# Create virtual environment
python3 -m venv venv

# Activate
source venv/bin/activate  # Linux/Mac
# or
venv\Scripts\activate     # Windows

# Install dependencies
pip install -r requirements.txt
```

### Run Development Server
```bash
python app.py
```

Server will start on http://localhost:5000

### Test the Service
```bash
# Health check
curl http://localhost:5000/health

# Parse a PDF
curl -X POST http://localhost:5000/parse \
  -F "file=@/path/to/bank_statement.pdf"
```

## Docker Deployment

### Build Image
```bash
docker build -t cashlens-pdf-parser:latest .
```

### Run Container
```bash
docker run -d \
  --name pdf-parser \
  -p 5000:5000 \
  cashlens-pdf-parser:latest
```

### Docker Compose (with main stack)
```yaml
# Add to docker-compose.yml
services:
  pdf-parser:
    build: ./services/pdf-parser
    ports:
      - "5000:5000"
    environment:
      - PORT=5000
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:5000/health"]
      interval: 30s
      timeout: 3s
      retries: 3
```

## How It Works

1. **pdfplumber First**: Tries text-based extraction (faster, works for most PDFs)
2. **Camelot Fallback**: If pdfplumber fails, tries lattice/stream mode (for scanned PDFs)
3. **Multi-Page Handling**: Merges tables across pages, deduplicates headers
4. **Cleaning**: Removes empty rows, trims whitespace, filters out summary rows

## Supported Bank Formats

The service extracts raw table data. Format-specific parsing is handled by the Go API using the existing CSV/XLSX parser logic.

- HDFC: Date, Narration, Withdrawal Amt, Deposit Amt
- ICICI: Transaction Date, Transaction Remarks, Withdrawal Amount, Deposit Amount
- SBI: Txn Date, Description, Debit, Credit
- Axis: Transaction Date, Particulars, Amount, Dr/Cr
- Kotak: Date, Description, Debit, Credit

## Error Handling

- 400: No file uploaded or invalid file type
- 413: File too large (>10MB)
- 500: PDF parsing error (corrupted file, password-protected, etc.)

## Performance

- Average parsing time: 2-5 seconds per page
- Memory usage: ~100-200MB per request
- Recommended: 2 gunicorn workers for production

## Security

- Non-root user in Docker container
- File size limits enforced
- Temporary files cleaned up after processing
- No persistent storage of uploaded files

## Dependencies

- **Flask**: Web framework
- **pdfplumber**: Text-based PDF extraction
- **camelot-py**: Table extraction with grid detection
- **pandas**: Data manipulation
- **opencv-python**: Image processing for camelot
- **gunicorn**: Production WSGI server
