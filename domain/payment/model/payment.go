package model

import (
	"time"

	"github.com/google/uuid"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	// PaymentStatusPending indicates a payment is awaiting processing
	PaymentStatusPending PaymentStatus = "pending"
	// PaymentStatusCompleted indicates a payment has been processed
	PaymentStatusCompleted PaymentStatus = "completed"
	// PaymentStatusRefunded indicates a payment has been refunded
	PaymentStatusRefunded PaymentStatus = "refunded"
	// PaymentStatusFailed indicates a payment has failed
	PaymentStatusFailed PaymentStatus = "failed"
)

// PaymentMethod represents the payment method
type PaymentMethod string

const (
	// PaymentMethodCreditCard represents credit card payment
	PaymentMethodCreditCard PaymentMethod = "credit_card"
	// PaymentMethodDebitCard represents debit card payment
	PaymentMethodDebitCard PaymentMethod = "debit_card"
	// PaymentMethodPayPal represents PayPal payment
	PaymentMethodPayPal PaymentMethod = "paypal"
	// PaymentMethodBankTransfer represents bank transfer payment
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	// PaymentMethodCash represents cash payment
	PaymentMethodCash PaymentMethod = "cash"
)

// Payment represents a payment transaction
type Payment struct {
	CreatedAt     time.Time
	UpdatedAt     time.Time
	PaidAt        *time.Time
	RefundedAt    *time.Time
	Status        PaymentStatus
	PaymentMethod PaymentMethod
	TransactionID string
	Amount        float64
	ID            uuid.UUID
	BookingID     uuid.UUID
}
