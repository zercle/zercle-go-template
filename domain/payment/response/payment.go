package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/domain/payment/model"
)

// PaymentResponse represents a payment response
type PaymentResponse struct {
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
	PaidAt        *time.Time          `json:"paid_at,omitempty"`
	RefundedAt    *time.Time          `json:"refunded_at,omitempty"`
	Status        model.PaymentStatus `json:"status"`
	PaymentMethod model.PaymentMethod `json:"payment_method"`
	TransactionID string              `json:"transaction_id"`
	Amount        float64             `json:"amount"`
	ID            uuid.UUID           `json:"id"`
	BookingID     uuid.UUID           `json:"booking_id"`
}

// ListPaymentsResponse represents a paginated list of payments
type ListPaymentsResponse struct {
	Payments []PaymentResponse `json:"payments"`
	Total    int               `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}
