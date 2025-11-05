package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/xuri/excelize/v2"
)

func main() {
	// Change to project root
	os.Chdir("/Users/asmitsingh/Desktop/side/cashlens/cashlens-api")

	generateHDFCFixture()
	generateICICIFixture()
	generateSBIFixture()
	generateAxisFixture()
	generateKotakFixture()
	fmt.Println("\n✅ All XLSX fixtures generated successfully!")
}

func generateHDFCFixture() {
	f := excelize.NewFile()
	sheet := "Sheet1"

	// Headers
	headers := []string{"Date", "Narration", "Chq./Ref.No.", "Value Dt", "Withdrawal Amt.", "Deposit Amt.", "Closing Balance"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Data rows (matching CSV data)
	data := [][]interface{}{
		{"15/01/2024", "AWS SERVICES", "UPI/123456", "15/01/2024", 3500.00, "", 450000.00},
		{"16/01/2024", "SALARY CREDIT - ACME CORP", "NEFT/789012", "16/01/2024", "", 50000.00, 500000.00},
		{"17/01/2024", "RAZORPAY PAYMENT GATEWAY", "UPI/234567", "17/01/2024", 2500.00, "", 497500.00},
		{"18/01/2024", "GOOGLE ADS MARKETING", "UPI/345678", "18/01/2024", 15000.00, "", 482500.00},
		{"19/01/2024", "SWIGGY TEAM LUNCH", "UPI/456789", "19/01/2024", 850.00, "", 481650.00},
		{"20/01/2024", "OFFICE SUPPLIES - AMAZON", "UPI/567890", "20/01/2024", 1200.00, "", 480450.00},
		{"22/01/2024", "DOMAIN RENEWAL - GODADDY", "CC/678901", "22/01/2024", 999.00, "", 479451.00},
		{"23/01/2024", "STRIPE PAYOUT", "NEFT/789123", "23/01/2024", "", 25000.00, 504451.00},
		{"24/01/2024", "CA FEES - TAX FILING", "UPI/890234", "24/01/2024", 5000.00, "", 499451.00},
		{"25/01/2024", "UBER FOR BUSINESS", "UPI/901345", "25/01/2024", 450.00, "", 499001.00},
	}

	for rowIdx, row := range data {
		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheet, cell, val)
		}
	}

	path := filepath.Join("testdata", "hdfc_sample.xlsx")
	if err := f.SaveAs(path); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Generated", path)
}

func generateICICIFixture() {
	f := excelize.NewFile()
	sheet := "Sheet1"

	// Headers
	headers := []string{"Value Date", "Transaction Date", "Cheque Number", "Transaction Remarks", "Withdrawal Amount (INR)", "Deposit Amount (INR)", "Balance (INR)"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Data rows
	data := [][]interface{}{
		{"15/01/2024", "15/01/2024", "UPI/123456", "PAYMENT TO AWS SERVICES", 3500.00, "", 450000.00},
		{"16/01/2024", "16/01/2024", "NEFT/789012", "SALARY CREDIT FROM ACME CORP", "", 50000.00, 500000.00},
		{"17/01/2024", "17/01/2024", "UPI/234567", "PAYMENT TO RAZORPAY", 2500.00, "", 497500.00},
		{"18/01/2024", "18/01/2024", "UPI/345678", "PAYMENT TO GOOGLE ADS", 15000.00, "", 482500.00},
		{"19/01/2024", "19/01/2024", "UPI/456789", "PAYMENT TO SWIGGY", 850.00, "", 481650.00},
		{"20/01/2024", "20/01/2024", "UPI/567890", "PAYMENT TO AMAZON", 1200.00, "", 480450.00},
		{"22/01/2024", "22/01/2024", "CC/678901", "PAYMENT TO GODADDY", 999.00, "", 479451.00},
		{"23/01/2024", "23/01/2024", "NEFT/789123", "PAYMENT FROM STRIPE", "", 25000.00, 504451.00},
		{"24/01/2024", "24/01/2024", "UPI/890234", "PAYMENT TO CA", 5000.00, "", 499451.00},
		{"25/01/2024", "25/01/2024", "UPI/901345", "PAYMENT TO UBER", 450.00, "", 499001.00},
	}

	for rowIdx, row := range data {
		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheet, cell, val)
		}
	}

	path := filepath.Join("testdata", "icici_sample.xlsx")
	if err := f.SaveAs(path); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Generated", path)
}

func generateSBIFixture() {
	f := excelize.NewFile()
	sheet := "Sheet1"

	// Headers
	headers := []string{"Txn Date", "Description", "Ref No./Cheque No.", "Value Date", "Debit", "Credit", "Balance"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Data rows
	data := [][]interface{}{
		{"15-Jan-2024", "PAYMENT TO AWS SERVICES", "UPI/123456", "15-Jan-2024", 3500.00, "", 450000.00},
		{"16-Jan-2024", "SALARY CREDIT FROM ACME", "NEFT/789012", "16-Jan-2024", "", 50000.00, 500000.00},
		{"17-Jan-2024", "PAYMENT TO RAZORPAY", "UPI/234567", "17-Jan-2024", 2500.00, "", 497500.00},
		{"18-Jan-2024", "PAYMENT TO GOOGLE", "UPI/345678", "18-Jan-2024", 15000.00, "", 482500.00},
		{"19-Jan-2024", "PAYMENT TO SWIGGY", "UPI/456789", "19-Jan-2024", 850.00, "", 481650.00},
		{"20-Jan-2024", "PAYMENT TO AMAZON", "UPI/567890", "20-Jan-2024", 1200.00, "", 480450.00},
		{"22-Jan-2024", "PAYMENT TO GODADDY", "CC/678901", "22-Jan-2024", 999.00, "", 479451.00},
		{"23-Jan-2024", "PAYMENT FROM STRIPE", "NEFT/789123", "23-Jan-2024", "", 25000.00, 504451.00},
		{"24-Jan-2024", "PAYMENT TO CA", "UPI/890234", "24-Jan-2024", 5000.00, "", 499451.00},
		{"25-Jan-2024", "PAYMENT TO UBER", "UPI/901345", "25-Jan-2024", 450.00, "", 499001.00},
	}

	for rowIdx, row := range data {
		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheet, cell, val)
		}
	}

	path := filepath.Join("testdata", "sbi_sample.xlsx")
	if err := f.SaveAs(path); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Generated", path)
}

func generateAxisFixture() {
	f := excelize.NewFile()
	sheet := "Sheet1"

	// Headers
	headers := []string{"Transaction Date", "Particulars", "Cheque No.", "Dr/Cr", "Amount", "Balance"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Data rows
	data := [][]interface{}{
		{"15/01/2024", "PAYMENT TO AWS SERVICES", "UPI/123456", "Dr", 3500.00, 450000.00},
		{"16/01/2024", "SALARY FROM ACME CORP", "NEFT/789012", "Cr", 50000.00, 500000.00},
		{"17/01/2024", "PAYMENT TO RAZORPAY", "UPI/234567", "Dr", 2500.00, 497500.00},
		{"18/01/2024", "PAYMENT TO GOOGLE", "UPI/345678", "Dr", 15000.00, 482500.00},
		{"19/01/2024", "PAYMENT TO SWIGGY", "UPI/456789", "Dr", 850.00, 481650.00},
		{"20/01/2024", "PAYMENT TO AMAZON", "UPI/567890", "Dr", 1200.00, 480450.00},
		{"22/01/2024", "PAYMENT TO GODADDY", "CC/678901", "Dr", 999.00, 479451.00},
		{"23/01/2024", "PAYMENT FROM STRIPE", "NEFT/789123", "Cr", 25000.00, 504451.00},
		{"24/01/2024", "PAYMENT TO CA", "UPI/890234", "Dr", 5000.00, 499451.00},
		{"25/01/2024", "PAYMENT TO UBER", "UPI/901345", "Dr", 450.00, 499001.00},
	}

	for rowIdx, row := range data {
		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheet, cell, val)
		}
	}

	path := filepath.Join("testdata", "axis_sample.xlsx")
	if err := f.SaveAs(path); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Generated", path)
}

func generateKotakFixture() {
	f := excelize.NewFile()
	sheet := "Sheet1"

	// Headers
	headers := []string{"Date", "Description", "Ref No.", "Debit", "Credit", "Balance"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Data rows
	data := [][]interface{}{
		{"15/01/2024", "PAYMENT TO AWS SERVICES", "UPI/123456", 3500.00, "", 450000.00},
		{"16/01/2024", "SALARY FROM ACME CORP", "NEFT/789012", "", 50000.00, 500000.00},
		{"17/01/2024", "PAYMENT TO RAZORPAY", "UPI/234567", 2500.00, "", 497500.00},
		{"18/01/2024", "PAYMENT TO GOOGLE", "UPI/345678", 15000.00, "", 482500.00},
		{"19/01/2024", "PAYMENT TO SWIGGY", "UPI/456789", 850.00, "", 481650.00},
		{"20/01/2024", "PAYMENT TO AMAZON", "UPI/567890", 1200.00, "", 480450.00},
		{"22/01/2024", "PAYMENT TO GODADDY", "CC/678901", 999.00, "", 479451.00},
		{"23/01/2024", "PAYMENT FROM STRIPE", "NEFT/789123", "", 25000.00, 504451.00},
		{"24/01/2024", "PAYMENT TO CA", "UPI/890234", 5000.00, "", 499451.00},
		{"25/01/2024", "PAYMENT TO UBER", "UPI/901345", 450.00, "", 499001.00},
	}

	for rowIdx, row := range data {
		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheet, cell, val)
		}
	}

	path := filepath.Join("testdata", "kotak_sample.xlsx")
	if err := f.SaveAs(path); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Generated", path)
}
