package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/domain/booking"
	"github.com/zercle/zercle-go-template/domain/booking/model"
	bookingRepository "github.com/zercle/zercle-go-template/domain/booking/repository"
	"github.com/zercle/zercle-go-template/domain/booking/request"
	bookingResponse "github.com/zercle/zercle-go-template/domain/booking/response"
	"github.com/zercle/zercle-go-template/domain/service"
	serviceRepository "github.com/zercle/zercle-go-template/domain/service/repository"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
)

var (
	// ErrBookingNotFound is returned when a booking cannot be found
	ErrBookingNotFound = errors.New("booking not found")
	// ErrServiceNotFound is returned when a service cannot be found
	ErrServiceNotFound = errors.New("service not found")
	// ErrBookingTimeInPast is returned when booking time is in the past
	ErrBookingTimeInPast = errors.New("booking time must be in future")
	// ErrBookingConflict is returned when booking conflicts with an existing booking
	ErrBookingConflict = errors.New("booking time conflicts with existing booking")
	// ErrInvalidStatus is returned when booking status is invalid
	ErrInvalidStatus = errors.New("invalid booking status")
	// ErrUnauthorizedAccess is returned when user tries to access another user's booking
	ErrUnauthorizedAccess = errors.New("unauthorized access to booking")
	// ErrCannotCancelComplete is returned when trying to cancel a completed booking
	ErrCannotCancelComplete = errors.New("cannot cancel a completed booking")
)

type bookingUseCase struct {
	repo        booking.Repository
	serviceRepo service.Repository
	log         *logger.Logger
}

// NewBookingUseCase creates a new booking use case with dependencies
func NewBookingUseCase(repo booking.Repository, serviceRepo service.Repository, log *logger.Logger) booking.Usecase {
	return &bookingUseCase{
		repo:        repo,
		serviceRepo: serviceRepo,
		log:         log,
	}
}

func (uc *bookingUseCase) CreateBooking(ctx context.Context, userID uuid.UUID, req request.CreateBooking) (*bookingResponse.BookingResponse, error) {
	// Validate booking time is in the future
	if req.StartTime.Before(time.Now()) {
		return nil, ErrBookingTimeInPast
	}

	// Get service details
	svc, err := uc.serviceRepo.GetByID(ctx, req.ServiceID)
	if err != nil {
		if errors.Is(err, serviceRepository.ErrServiceNotFound) {
			return nil, ErrServiceNotFound
		}
		return nil, err
	}

	// Check if service is active
	if !svc.IsActive {
		return nil, errors.New("service is not available for booking")
	}

	// Calculate end time based on service duration
	endTime := req.StartTime.Add(time.Duration(svc.DurationMinutes) * time.Minute)

	// Check for booking conflicts
	conflicts, err := uc.repo.CheckConflict(ctx, req.ServiceID, req.StartTime, endTime)
	if err != nil {
		uc.log.Error("Failed to check booking conflicts", "error", err, "service_id", req.ServiceID)
		return nil, err
	}
	if len(conflicts) > 0 {
		return nil, ErrBookingConflict
	}

	// Calculate total price based on service price
	totalPrice := svc.Price

	// Create booking
	bookingModel := &model.Booking{
		UserID:     userID,
		ServiceID:  req.ServiceID,
		StartTime:  req.StartTime,
		EndTime:    endTime,
		Status:     model.BookingStatusPending,
		TotalPrice: totalPrice,
		Notes:      req.Notes,
	}

	created, err := uc.repo.Create(ctx, bookingModel)
	if err != nil {
		uc.log.Error("Failed to create booking", "error", err, "user_id", userID)
		return nil, err
	}

	return toBookingResponse(created), nil
}

func (uc *bookingUseCase) GetBooking(ctx context.Context, id uuid.UUID) (*bookingResponse.BookingResponse, error) {
	bookingModel, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, bookingRepository.ErrBookingNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}

	return toBookingResponse(bookingModel), nil
}

func (uc *bookingUseCase) CancelBooking(ctx context.Context, id, userID uuid.UUID) (*bookingResponse.BookingResponse, error) {
	// Get booking first to check ownership
	bookingModel, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, bookingRepository.ErrBookingNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}

	// Check ownership
	if bookingModel.UserID != userID {
		return nil, ErrUnauthorizedAccess
	}

	// Check if can be canceled
	if bookingModel.Status == model.BookingStatusCompleted {
		return nil, ErrCannotCancelComplete
	}

	if bookingModel.Status == model.BookingStatusCancelled {
		return toBookingResponse(bookingModel), nil
	}

	// Cancel booking
	canceled, err := uc.repo.Cancel(ctx, id)
	if err != nil {
		uc.log.Error("Failed to cancel booking", "error", err, "booking_id", id)
		return nil, err
	}

	return toBookingResponse(canceled), nil
}

func (uc *bookingUseCase) UpdateBookingStatus(ctx context.Context, id uuid.UUID, req request.UpdateBookingStatus) (*bookingResponse.BookingResponse, error) {
	// Validate status transition
	bookingModel, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, bookingRepository.ErrBookingNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}

	// Validate status transition
	if !isValidStatusTransition(bookingModel.Status, req.Status) {
		return nil, ErrInvalidStatus
	}

	// Update status
	updated, err := uc.repo.UpdateStatus(ctx, id, req.Status)
	if err != nil {
		uc.log.Error("Failed to update booking status", "error", err, "booking_id", id)
		return nil, err
	}

	return toBookingResponse(updated), nil
}

func (uc *bookingUseCase) ListBookingsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) (*bookingResponse.ListBookingsResponse, error) {
	// Apply defaults
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	bookings, err := uc.repo.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		uc.log.Error("Failed to list bookings by user", "error", err, "user_id", userID)
		return nil, err
	}

	bookingResponses := make([]bookingResponse.BookingResponse, len(bookings))
	for i, b := range bookings {
		bookingResponses[i] = *toBookingResponse(b)
	}

	return &bookingResponse.ListBookingsResponse{
		Bookings: bookingResponses,
		Total:    len(bookingResponses),
		Limit:    limit,
		Offset:   offset,
	}, nil
}

func (uc *bookingUseCase) ListBookingsByService(ctx context.Context, serviceID uuid.UUID, limit, offset int) ([]bookingResponse.BookingResponse, error) {
	// Apply defaults
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	bookings, err := uc.repo.ListByService(ctx, serviceID, limit, offset)
	if err != nil {
		uc.log.Error("Failed to list bookings by service", "error", err, "service_id", serviceID)
		return nil, err
	}

	responses := make([]bookingResponse.BookingResponse, len(bookings))
	for i, b := range bookings {
		responses[i] = *toBookingResponse(b)
	}

	return responses, nil
}

func (uc *bookingUseCase) ListBookingsByDateRange(ctx context.Context, req request.ListBookingsByDateRange) ([]bookingResponse.BookingResponse, error) {
	// Validate date range
	if req.EndDate.Before(req.StartDate) || req.EndDate.Sub(req.StartDate) > 90*24*time.Hour {
		return nil, errors.New("end date must be after start date and within 90 days")
	}

	// Apply defaults
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	bookings, err := uc.repo.ListByDateRange(ctx, req.StartDate, req.EndDate, int(req.Limit), int(req.Offset))
	if err != nil {
		uc.log.Error("Failed to list bookings by date range", "error", err)
		return nil, err
	}

	responses := make([]bookingResponse.BookingResponse, len(bookings))
	for i, b := range bookings {
		responses[i] = *toBookingResponse(b)
	}

	return responses, nil
}

// toBookingResponse converts a booking model to response DTO
func toBookingResponse(b *model.Booking) *bookingResponse.BookingResponse {
	return &bookingResponse.BookingResponse{
		ID:          b.ID,
		UserID:      b.UserID,
		ServiceID:   b.ServiceID,
		StartTime:   b.StartTime,
		EndTime:     b.EndTime,
		Status:      b.Status,
		TotalPrice:  b.TotalPrice,
		Notes:       b.Notes,
		CancelledAt: b.CancelledAt,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}

// isValidStatusTransition validates if a status transition is allowed
func isValidStatusTransition(current, newStatus model.BookingStatus) bool {
	// Define valid transitions
	validTransitions := map[model.BookingStatus][]model.BookingStatus{
		model.BookingStatusPending:   {model.BookingStatusConfirmed, model.BookingStatusCancelled},
		model.BookingStatusConfirmed: {model.BookingStatusCompleted, model.BookingStatusCancelled},
		model.BookingStatusCompleted: {}, // No transitions from completed
		model.BookingStatusCancelled: {}, // No transitions from canceled
	}

	allowed, exists := validTransitions[current]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == newStatus {
			return true
		}
	}

	return false
}
