package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ashmitsharp/cashlens-api/internal/services"
)

// Example usage of the unified parser interface
func main() {
	parser := services.NewParser()

	// Example 1: Parse CSV file
	fmt.Println("Example 1: Parsing CSV file...")
	csvFile, err := os.Open("../testdata/hdfc_sample.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()

	transactions, err := parser.ParseFile(csvFile, "hdfc_sample.csv")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Parsed %d transactions from CSV\n", len(transactions))
	fmt.Printf("First transaction: %s - %.2f (%s)\n\n",
		transactions[0].Description, transactions[0].Amount, transactions[0].TxnType)

	// Example 2: Parse XLSX file
	fmt.Println("Example 2: Parsing XLSX file...")
	xlsxFile, err := os.Open("../testdata/icici_sample.xlsx")
	if err != nil {
		log.Fatal(err)
	}
	defer xlsxFile.Close()

	transactions, err = parser.ParseFile(xlsxFile, "icici_sample.xlsx")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Parsed %d transactions from XLSX\n", len(transactions))
	fmt.Printf("First transaction: %s - %.2f (%s)\n\n",
		transactions[0].Description, transactions[0].Amount, transactions[0].TxnType)

	// Example 3: Parse PDF file (requires Python microservice running)
	// Uncomment the following lines when PDF service is available:
	/*
		fmt.Println("Example 3: Parsing PDF file...")
		pdfFile, err := os.Open("../testdata/statement.pdf")
		if err != nil {
			log.Fatal(err)
		}
		defer pdfFile.Close()

		transactions, err = parser.ParseFile(pdfFile, "statement.pdf")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Parsed %d transactions from PDF\n", len(transactions))
	*/

	// Example 4: Unsupported file type
	fmt.Println("Example 4: Handling unsupported file type...")
	csvFile2, _ := os.Open("../testdata/hdfc_sample.csv")
	defer csvFile2.Close()

	_, err = parser.ParseFile(csvFile2, "document.docx")
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	}
}
