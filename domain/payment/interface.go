//go:generate go run go.uber.org/mock/mockgen@latest -source=$GOFILE -destination=mock/$GOFILE -package=mock

package payment

import (
	"context"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/zercle/zercle-go-template/domain/payment/model"
	"github.com/zercle/zercle-go-template/domain/payment/request"
	"github.com/zercle/zercle-go-template/domain/payment/response"
)

// Repository defines the data access interface for payments
type Repository interface {
	Create(ctx context.Context, payment *model.Payment) (*model.Payment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Payment, error)
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*model.Payment, error)
	GetByTransactionID(ctx context.Context, transactionID string) (*model.Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.PaymentStatus, transactionID string) (*model.Payment, error)
	Confirm(ctx context.Context, id uuid.UUID) (*model.Payment, error)
	Refund(ctx context.Context, id uuid.UUID) (*model.Payment, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*model.Payment, error)
	ListByBooking(ctx context.Context, bookingID uuid.UUID) ([]*model.Payment, error)
}

// Usecase defines the business logic interface for payments
type Usecase interface {
	CreatePayment(ctx context.Context, req request.CreatePayment) (*response.PaymentResponse, error)
	GetPayment(ctx context.Context, id uuid.UUID) (*response.PaymentResponse, error)
	GetPaymentByBooking(ctx context.Context, bookingID uuid.UUID) (*response.PaymentResponse, error)
	ConfirmPayment(ctx context.Context, id uuid.UUID) (*response.PaymentResponse, error)
	RefundPayment(ctx context.Context, id uuid.UUID, req request.RefundPayment) (*response.PaymentResponse, error)
	UpdatePaymentStatus(ctx context.Context, id uuid.UUID, req request.UpdatePaymentStatus) (*response.PaymentResponse, error)
	ListPayments(ctx context.Context, req request.ListPayments) (*response.ListPaymentsResponse, error)
}

// Handler defines the HTTP handler interface for payments
type Handler interface {
	CreatePayment(c echo.Context) error
	GetPayment(c echo.Context) error
	GetPaymentByBooking(c echo.Context) error
	ConfirmPayment(c echo.Context) error
	RefundPayment(c echo.Context) error
	ListPayments(c echo.Context) error
}
