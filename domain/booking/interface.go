//go:generate go run go.uber.org/mock/mockgen@latest -source=$GOFILE -destination=mock/$GOFILE -package=mock

package booking

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/zercle/zercle-go-template/domain/booking/model"
	"github.com/zercle/zercle-go-template/domain/booking/request"
	"github.com/zercle/zercle-go-template/domain/booking/response"
)

// Repository defines the data access interface for bookings
type Repository interface {
	Create(ctx context.Context, booking *model.Booking) (*model.Booking, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Booking, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.BookingStatus) (*model.Booking, error)
	Cancel(ctx context.Context, id uuid.UUID) (*model.Booking, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Booking, error)
	ListByService(ctx context.Context, serviceID uuid.UUID, limit, offset int) ([]*model.Booking, error)
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*model.Booking, error)
	ListByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*model.Booking, error)
	CheckConflict(ctx context.Context, serviceID uuid.UUID, startTime, endTime time.Time) ([]*model.Booking, error)
}

// Usecase defines the business logic interface for bookings
type Usecase interface {
	CreateBooking(ctx context.Context, userID uuid.UUID, req request.CreateBooking) (*response.BookingResponse, error)
	GetBooking(ctx context.Context, id uuid.UUID) (*response.BookingResponse, error)
	CancelBooking(ctx context.Context, id, userID uuid.UUID) (*response.BookingResponse, error)
	UpdateBookingStatus(ctx context.Context, id uuid.UUID, req request.UpdateBookingStatus) (*response.BookingResponse, error)
	ListBookingsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) (*response.ListBookingsResponse, error)
	ListBookingsByService(ctx context.Context, serviceID uuid.UUID, limit, offset int) ([]response.BookingResponse, error)
	ListBookingsByDateRange(ctx context.Context, req request.ListBookingsByDateRange) ([]response.BookingResponse, error)
}

// Handler defines the HTTP handler interface for bookings
type Handler interface {
	CreateBooking(c echo.Context) error
	GetBooking(c echo.Context) error
	CancelBooking(c echo.Context) error
	UpdateBookingStatus(c echo.Context) error
	ListBookingsByUser(c echo.Context) error
	ListBookingsByService(c echo.Context) error
	ListBookingsByDateRange(c echo.Context) error
}
