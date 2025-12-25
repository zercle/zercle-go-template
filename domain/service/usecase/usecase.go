package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/domain/service"
	"github.com/zercle/zercle-go-template/domain/service/model"
	"github.com/zercle/zercle-go-template/domain/service/repository"
	"github.com/zercle/zercle-go-template/domain/service/request"
	serviceResponse "github.com/zercle/zercle-go-template/domain/service/response"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
)

var (
	// ErrServiceNotFound is returned when a service cannot be found
	ErrServiceNotFound = errors.New("service not found")
	// ErrInvalidServicePrice is returned when service price is invalid
	ErrInvalidServicePrice = errors.New("price must be greater than 0")
	// ErrInvalidDuration is returned when service duration is invalid
	ErrInvalidDuration = errors.New("duration must be between 1 and 480 minutes")
	// ErrInvalidCapacity is returned when service capacity is invalid
	ErrInvalidCapacity = errors.New("capacity must be between 1 and 50")
	// ErrServiceAlreadyExists is returned when service name already exists
	ErrServiceAlreadyExists = errors.New("service with this name already exists")
)

type serviceUseCase struct {
	repo service.Repository
	log  *logger.Logger
}

// NewServiceUseCase creates a new service use case
func NewServiceUseCase(repo service.Repository, log *logger.Logger) service.Usecase {
	return &serviceUseCase{
		repo: repo,
		log:  log,
	}
}

func (uc *serviceUseCase) CreateService(ctx context.Context, req request.CreateService) (*serviceResponse.ServiceResponse, error) {
	// Validate business rules
	if req.Price <= 0 {
		return nil, ErrInvalidServicePrice
	}
	if req.DurationMinutes < 1 || req.DurationMinutes > 480 {
		return nil, ErrInvalidDuration
	}
	if req.MaxCapacity < 1 || req.MaxCapacity > 50 {
		return nil, ErrInvalidCapacity
	}

	// Create service model
	svc := &model.Service{
		Name:            req.Name,
		Description:     req.Description,
		DurationMinutes: req.DurationMinutes,
		Price:           req.Price,
		MaxCapacity:     req.MaxCapacity,
		IsActive:        req.IsActive,
	}

	created, err := uc.repo.Create(ctx, svc)
	if err != nil {
		uc.log.Error("Failed to create service", "error", err, "name", req.Name)
		return nil, err
	}

	return toServiceResponse(created), nil
}

func (uc *serviceUseCase) GetService(ctx context.Context, id uuid.UUID) (*serviceResponse.ServiceResponse, error) {
	svc, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrServiceNotFound) {
			return nil, ErrServiceNotFound
		}
		uc.log.Error("Failed to get service", "error", err, "service_id", id)
		return nil, err
	}

	return toServiceResponse(svc), nil
}

func (uc *serviceUseCase) UpdateService(ctx context.Context, id uuid.UUID, req request.UpdateService) (*serviceResponse.ServiceResponse, error) {
	// Get existing service
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrServiceNotFound) {
			return nil, ErrServiceNotFound
		}
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.DurationMinutes != nil {
		if *req.DurationMinutes < 1 || *req.DurationMinutes > 480 {
			return nil, ErrInvalidDuration
		}
		existing.DurationMinutes = *req.DurationMinutes
	}
	if req.Price != nil {
		if *req.Price <= 0 {
			return nil, ErrInvalidServicePrice
		}
		existing.Price = *req.Price
	}
	if req.MaxCapacity != nil {
		if *req.MaxCapacity < 1 || *req.MaxCapacity > 50 {
			return nil, ErrInvalidCapacity
		}
		existing.MaxCapacity = *req.MaxCapacity
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	existing.UpdatedAt = time.Now()

	// Save
	updated, err := uc.repo.Update(ctx, existing)
	if err != nil {
		uc.log.Error("Failed to update service", "error", err, "service_id", id)
		return nil, err
	}

	return toServiceResponse(updated), nil
}

func (uc *serviceUseCase) DeleteService(ctx context.Context, id uuid.UUID) error {
	err := uc.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrServiceNotFound) {
			return ErrServiceNotFound
		}
		uc.log.Error("Failed to delete service", "error", err, "service_id", id)
		return err
	}
	return nil
}

func (uc *serviceUseCase) ListServices(ctx context.Context, req request.ListServices) (*serviceResponse.ListServicesResponse, error) {
	// Apply defaults if not provided
	limit := int(req.Limit)
	offset := int(req.Offset)
	if limit == 0 {
		limit = 20 // default limit
	}

	services, err := uc.repo.List(ctx, req.IsActive, limit, offset)
	if err != nil {
		uc.log.Error("Failed to list services", "error", err)
		return nil, err
	}

	serviceResponses := make([]serviceResponse.ServiceResponse, len(services))
	for i, svc := range services {
		serviceResponses[i] = *toServiceResponse(svc)
	}

	return &serviceResponse.ListServicesResponse{
		Services: serviceResponses,
		Total:    len(serviceResponses),
		Limit:    limit,
		Offset:   offset,
		IsActive: req.IsActive,
	}, nil
}

func (uc *serviceUseCase) SearchServices(ctx context.Context, name string, isActive bool, limit int) ([]serviceResponse.ServiceResponse, error) {
	if limit == 0 {
		limit = 20 // default limit
	}

	services, err := uc.repo.SearchByName(ctx, name, isActive, limit)
	if err != nil {
		uc.log.Error("Failed to search services", "error", err, "name", name)
		return nil, err
	}

	responses := make([]serviceResponse.ServiceResponse, len(services))
	for i, svc := range services {
		responses[i] = *toServiceResponse(svc)
	}

	return responses, nil
}

// toServiceResponse converts a service model to response DTO
func toServiceResponse(svc *model.Service) *serviceResponse.ServiceResponse {
	return &serviceResponse.ServiceResponse{
		ID:              svc.ID,
		Name:            svc.Name,
		Description:     svc.Description,
		DurationMinutes: svc.DurationMinutes,
		Price:           svc.Price,
		MaxCapacity:     svc.MaxCapacity,
		IsActive:        svc.IsActive,
		CreatedAt:       svc.CreatedAt,
		UpdatedAt:       svc.UpdatedAt,
	}
}
