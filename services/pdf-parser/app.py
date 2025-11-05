"""
PDF Parser Microservice for Bank Statements

This Flask service extracts transaction tables from PDF bank statements
using pdfplumber and camelot libraries. It returns structured JSON data
that can be parsed by the main Go API.

Supported banks: HDFC, ICICI, SBI, Axis, Kotak
"""

import os
import json
import tempfile
from flask import Flask, request, jsonify
import pdfplumber
import camelot
import pandas as pd
from datetime import datetime

app = Flask(__name__)

# Configure max file size (10MB)
app.config['MAX_CONTENT_LENGTH'] = 10 * 1024 * 1024


@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({
        'status': 'ok',
        'service': 'pdf-parser',
        'version': '1.0.0'
    })


@app.route('/parse', methods=['POST'])
def parse_pdf():
    """
    Parse PDF bank statement and extract transactions

    Expected request:
    - Content-Type: multipart/form-data
    - Field: 'file' (PDF file)

    Returns:
    {
        "rows": [
            ["Date", "Description", "Debit", "Credit"],
            ["01/01/2024", "AWS SERVICES", "3500.00", ""],
            ...
        ],
        "pages_processed": 3,
        "method_used": "pdfplumber" | "camelot",
        "total_rows": 50
    }
    """
    # Validate file upload
    if 'file' not in request.files:
        return jsonify({'error': 'No file uploaded'}), 400

    file = request.files['file']

    if file.filename == '':
        return jsonify({'error': 'Empty filename'}), 400

    if not file.filename.lower().endswith('.pdf'):
        return jsonify({'error': 'File must be a PDF'}), 400

    # Save uploaded file temporarily
    with tempfile.NamedTemporaryFile(delete=False, suffix='.pdf') as tmp_file:
        file.save(tmp_file.name)
        tmp_path = tmp_file.name

    try:
        # Try pdfplumber first (better for simple tables)
        result = extract_with_pdfplumber(tmp_path)

        # If pdfplumber fails or returns too few rows, try camelot
        if not result or len(result.get('rows', [])) < 2:
            result = extract_with_camelot(tmp_path)

        return jsonify(result)

    except Exception as e:
        app.logger.error(f"PDF parsing error: {str(e)}")
        return jsonify({'error': f'Failed to parse PDF: {str(e)}'}), 500

    finally:
        # Clean up temp file
        if os.path.exists(tmp_path):
            os.unlink(tmp_path)


def extract_with_pdfplumber(pdf_path):
    """
    Extract tables using pdfplumber (better for text-based PDFs)
    """
    all_rows = []
    pages_processed = 0

    with pdfplumber.open(pdf_path) as pdf:
        for page_num, page in enumerate(pdf.pages, start=1):
            # Extract tables from page
            tables = page.extract_tables()

            if not tables:
                continue

            pages_processed += 1

            # Process each table (usually 1 per page)
            for table in tables:
                # Filter out empty rows
                filtered_rows = [
                    row for row in table
                    if row and any(cell and str(cell).strip() for cell in row)
                ]

                # Skip if table is too small (not a transaction table)
                if len(filtered_rows) < 2:
                    continue

                # If this is the first table, include headers
                if not all_rows:
                    all_rows.extend(filtered_rows)
                else:
                    # For subsequent pages, skip headers if they match
                    if is_header_row(filtered_rows[0], all_rows[0]):
                        all_rows.extend(filtered_rows[1:])
                    else:
                        all_rows.extend(filtered_rows)

    return {
        'rows': clean_rows(all_rows),
        'pages_processed': pages_processed,
        'method_used': 'pdfplumber',
        'total_rows': len(all_rows)
    }


def extract_with_camelot(pdf_path):
    """
    Extract tables using camelot (better for complex/scanned PDFs)
    """
    # Try lattice mode first (for PDFs with grid lines)
    try:
        tables = camelot.read_pdf(pdf_path, pages='all', flavor='lattice')
    except:
        # Fall back to stream mode (for PDFs without grid lines)
        tables = camelot.read_pdf(pdf_path, pages='all', flavor='stream')

    if not tables:
        return {
            'rows': [],
            'pages_processed': 0,
            'method_used': 'camelot',
            'total_rows': 0
        }

    all_rows = []

    for table_num, table in enumerate(tables):
        # Convert to pandas DataFrame
        df = table.df

        # Convert to list of lists
        rows = df.values.tolist()

        # Filter out empty rows
        filtered_rows = [
            row for row in rows
            if any(cell and str(cell).strip() for cell in row)
        ]

        # Skip small tables
        if len(filtered_rows) < 2:
            continue

        # If first table, include headers
        if not all_rows:
            all_rows.extend(filtered_rows)
        else:
            # Skip duplicate headers
            if is_header_row(filtered_rows[0], all_rows[0]):
                all_rows.extend(filtered_rows[1:])
            else:
                all_rows.extend(filtered_rows)

    return {
        'rows': clean_rows(all_rows),
        'pages_processed': len(tables),
        'method_used': 'camelot',
        'total_rows': len(all_rows)
    }


def is_header_row(row, header_row):
    """
    Check if row is a duplicate header by comparing with known header
    """
    if not row or not header_row:
        return False

    # Compare first 3 cells
    for i in range(min(3, len(row), len(header_row))):
        if str(row[i]).strip().lower() == str(header_row[i]).strip().lower():
            return True

    return False


def clean_rows(rows):
    """
    Clean extracted rows by removing None values and extra whitespace
    """
    cleaned = []

    for row in rows:
        cleaned_row = [
            str(cell).strip() if cell is not None else ''
            for cell in row
        ]
        # Only include rows with at least one non-empty cell
        if any(cell for cell in cleaned_row):
            cleaned.append(cleaned_row)

    return cleaned


@app.errorhandler(413)
def too_large(e):
    """Handle file too large error"""
    return jsonify({'error': 'File too large (max 10MB)'}), 413


@app.errorhandler(500)
def internal_error(e):
    """Handle internal server errors"""
    return jsonify({'error': 'Internal server error'}), 500


if __name__ == '__main__':
    # Development server
    port = int(os.environ.get('PORT', 5000))
    app.run(host='0.0.0.0', port=port, debug=True)
