package request

import (
	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/domain/payment/model"
)

// CreatePayment represents a request to create a new payment
type CreatePayment struct {
	PaymentMethod model.PaymentMethod `json:"payment_method" validate:"required"`
	TransactionID string              `json:"transaction_id"`
	BookingID     uuid.UUID           `json:"booking_id" validate:"required"`
}

// ConfirmPayment represents a request to confirm a payment
type ConfirmPayment struct {
	TransactionID string `json:"transaction_id" validate:"required"`
}

// RefundPayment represents a request to refund a payment
type RefundPayment struct {
	Reason string `json:"reason" validate:"max=500"`
}

// UpdatePaymentStatus represents a request to update payment status
type UpdatePaymentStatus struct {
	Status        model.PaymentStatus `json:"status" validate:"required,oneof=pending completed refunded failed"`
	TransactionID string              `json:"transaction_id,omitempty"`
}

// ListPayments represents a request to list payments with pagination
type ListPayments struct {
	Status string `query:"status" validate:"omitempty,oneof=pending completed refunded failed"`
	Limit  int    `query:"limit" validate:"min=1,max=100"`
	Offset int    `query:"offset" validate:"min=0"`
}
