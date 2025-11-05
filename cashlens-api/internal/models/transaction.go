package models

import (
	"time"

	"github.com/google/uuid"
)

// Transaction represents a parsed bank transaction
type Transaction struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	TxnDate     time.Time  `json:"txn_date"`
	Description string     `json:"description"`
	Amount      float64    `json:"amount"` // Negative for debit, positive for credit
	TxnType     string     `json:"txn_type"` // "credit" or "debit"
	Category    *string    `json:"category,omitempty"` // Nullable, set after categorization
	IsReviewed  bool       `json:"is_reviewed"`
	RawData     *string    `json:"raw_data,omitempty"` // Original CSV row for debugging
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ParsedTransaction represents a transaction after CSV parsing but before DB insertion
type ParsedTransaction struct {
	TxnDate     time.Time `json:"txn_date"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"` // Negative for debit, positive for credit
	TxnType     string    `json:"txn_type"` // "credit" or "debit"
	RawData     string    `json:"raw_data"` // Original CSV row
}

// BankSchema defines the column structure for each bank's CSV format
type BankSchema struct {
	BankName          string
	DateColumn        string
	DescriptionColumn string
	DebitColumn       string  // For banks with separate debit/credit columns
	CreditColumn      string
	AmountColumn      string  // For banks with single amount column
	DrCrColumn        string  // For banks with Dr/Cr indicator
	HasSeparateAmounts bool   // true if debit/credit are separate columns
}
