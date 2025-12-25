package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zercle/zercle-go-template/domain/service"
	"github.com/zercle/zercle-go-template/domain/service/model"
	"github.com/zercle/zercle-go-template/infrastructure/logger"
	"github.com/zercle/zercle-go-template/infrastructure/sqlc/db"
)

// ErrServiceNotFound is returned when a service cannot be found
var ErrServiceNotFound = errors.New("service not found")

type serviceRepository struct {
	sqlc *db.Queries
	log  *logger.Logger
}

// NewServiceRepository creates a new service repository
func NewServiceRepository(sqlc *db.Queries, log *logger.Logger) service.Repository {
	return &serviceRepository{
		sqlc: sqlc,
		log:  log,
	}
}

// Helper functions for pgtype conversions
func toUUID(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: u, Valid: true}
}

func fromUUID(u pgtype.UUID) uuid.UUID {
	return u.Bytes
}

func toText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

func fromText(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func toTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func fromTimestamptz(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

// fromNumeric converts pgtype.Numeric to float64
// pgtype.Numeric stores values as Int * 10^Exp where Exp is negative for decimal places
// For example, DECIMAL(10,2) stores 100.50 as Int=10050, Exp=-2
func fromNumeric(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	if n.Int == nil {
		return 0
	}
	str := n.Int.String()
	exp := int(n.Exp)

	if exp < 0 {
		exp = -exp
		if len(str) <= exp {
			for len(str) < exp {
				str = "0" + str
			}
			str = "0." + str
		} else {
			pos := len(str) - exp
			str = str[:pos] + "." + str[pos:]
		}
	} else if exp > 0 {
		str += strings.Repeat("0", exp)
	}

	var f float64
	_, _ = fmt.Sscanf(str, "%f", &f)
	return f
}

// toNumeric converts float64 to pgtype.Numeric
func toNumeric(f float64) pgtype.Numeric {
	// Use the string-based numeric scan
	n := pgtype.Numeric{}
	_ = n.Scan(fmt.Sprintf("%.2f", f))
	return n
}

func toInt4(i int) pgtype.Int4 {
	if i < math.MinInt32 || i > math.MaxInt32 {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: int32(i), Valid: true}
}

// toInt32Safe safely converts int to int32 with overflow check.
// Panics if value is outside int32 range (should not happen with validated input).
func toInt32Safe(i int) int32 {
	if i < math.MinInt32 || i > math.MaxInt32 {
		panic(fmt.Sprintf("value %d overflows int32", i))
	}
	return int32(i)
}

func toBool(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}

func (r *serviceRepository) Create(ctx context.Context, svc *model.Service) (*model.Service, error) {
	now := time.Now()
	params := db.CreateServiceParams{
		Name:            svc.Name,
		Description:     toText(svc.Description),
		DurationMinutes: toInt32Safe(svc.DurationMinutes),
		Price:           toNumeric(svc.Price),
		MaxCapacity:     toInt32Safe(svc.MaxCapacity),
		IsActive:        svc.IsActive,
		CreatedAt:       toTimestamptz(now),
		UpdatedAt:       toTimestamptz(now),
	}

	row, err := r.sqlc.CreateService(ctx, params)
	if err != nil {
		r.log.Error("Failed to create service", "error", err, "name", svc.Name)
		return nil, err
	}

	return &model.Service{
		ID:              fromUUID(row.ID),
		Name:            row.Name,
		Description:     fromText(row.Description),
		DurationMinutes: int(row.DurationMinutes),
		Price:           fromNumeric(row.Price),
		MaxCapacity:     int(row.MaxCapacity),
		IsActive:        row.IsActive,
		CreatedAt:       fromTimestamptz(row.CreatedAt),
		UpdatedAt:       fromTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *serviceRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Service, error) {
	row, err := r.sqlc.GetService(ctx, toUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrServiceNotFound
		}
		r.log.Error("Failed to get service by ID", "error", err, "service_id", id)
		return nil, err
	}

	return &model.Service{
		ID:              fromUUID(row.ID),
		Name:            row.Name,
		Description:     fromText(row.Description),
		DurationMinutes: int(row.DurationMinutes),
		Price:           fromNumeric(row.Price),
		MaxCapacity:     int(row.MaxCapacity),
		IsActive:        row.IsActive,
		CreatedAt:       fromTimestamptz(row.CreatedAt),
		UpdatedAt:       fromTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *serviceRepository) Update(ctx context.Context, svc *model.Service) (*model.Service, error) {
	params := db.UpdateServiceParams{
		ID:              toUUID(svc.ID),
		UpdatedAt:       toTimestamptz(time.Now()),
		Name:            toText(svc.Name),
		Description:     toText(svc.Description),
		DurationMinutes: toInt4(svc.DurationMinutes),
		Price:           toNumeric(svc.Price),
		MaxCapacity:     toInt4(svc.MaxCapacity),
		IsActive:        toBool(svc.IsActive),
	}

	row, err := r.sqlc.UpdateService(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrServiceNotFound
		}
		r.log.Error("Failed to update service", "error", err, "service_id", svc.ID)
		return nil, err
	}

	return &model.Service{
		ID:              fromUUID(row.ID),
		Name:            row.Name,
		Description:     fromText(row.Description),
		DurationMinutes: int(row.DurationMinutes),
		Price:           fromNumeric(row.Price),
		MaxCapacity:     int(row.MaxCapacity),
		IsActive:        row.IsActive,
		CreatedAt:       fromTimestamptz(row.CreatedAt),
		UpdatedAt:       fromTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *serviceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.sqlc.DeleteService(ctx, toUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrServiceNotFound
		}
		r.log.Error("Failed to delete service", "error", err, "service_id", id)
		return err
	}
	return nil
}

func (r *serviceRepository) List(ctx context.Context, isActive bool, limit, offset int) ([]*model.Service, error) {
	params := db.ListServicesParams{
		IsActive: isActive,
		Limit:    toInt32Safe(limit),
		Offset:   toInt32Safe(offset),
	}

	rows, err := r.sqlc.ListServices(ctx, params)
	if err != nil {
		r.log.Error("Failed to list services", "error", err)
		return nil, err
	}

	services := make([]*model.Service, len(rows))
	for i, row := range rows {
		services[i] = &model.Service{
			ID:              fromUUID(row.ID),
			Name:            row.Name,
			Description:     fromText(row.Description),
			DurationMinutes: int(row.DurationMinutes),
			Price:           fromNumeric(row.Price),
			MaxCapacity:     int(row.MaxCapacity),
			IsActive:        row.IsActive,
			CreatedAt:       fromTimestamptz(row.CreatedAt),
			UpdatedAt:       fromTimestamptz(row.UpdatedAt),
		}
	}

	return services, nil
}

func (r *serviceRepository) SearchByName(ctx context.Context, name string, isActive bool, limit int) ([]*model.Service, error) {
	params := db.SearchServicesByNameParams{
		Name:     "%" + name + "%", // Add ILIKE wildcards
		IsActive: isActive,
		Limit:    toInt32Safe(limit),
	}

	rows, err := r.sqlc.SearchServicesByName(ctx, params)
	if err != nil {
		r.log.Error("Failed to search services", "error", err, "name", name)
		return nil, err
	}

	services := make([]*model.Service, len(rows))
	for i, row := range rows {
		services[i] = &model.Service{
			ID:              fromUUID(row.ID),
			Name:            row.Name,
			Description:     fromText(row.Description),
			DurationMinutes: int(row.DurationMinutes),
			Price:           fromNumeric(row.Price),
			MaxCapacity:     int(row.MaxCapacity),
			IsActive:        row.IsActive,
			CreatedAt:       fromTimestamptz(row.CreatedAt),
			UpdatedAt:       fromTimestamptz(row.UpdatedAt),
		}
	}

	return services, nil
}

func (r *serviceRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.Service, error) {
	if len(ids) == 0 {
		return []*model.Service{}, nil
	}

	pgIDs := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		pgIDs[i] = toUUID(id)
	}

	rows, err := r.sqlc.GetServicesByIds(ctx, pgIDs)
	if err != nil {
		r.log.Error("Failed to get services by IDs", "error", err, "count", len(ids))
		return nil, err
	}

	services := make([]*model.Service, len(rows))
	for i, row := range rows {
		services[i] = &model.Service{
			ID:              fromUUID(row.ID),
			Name:            row.Name,
			Description:     fromText(row.Description),
			DurationMinutes: int(row.DurationMinutes),
			Price:           fromNumeric(row.Price),
			MaxCapacity:     int(row.MaxCapacity),
			IsActive:        row.IsActive,
			CreatedAt:       fromTimestamptz(row.CreatedAt),
			UpdatedAt:       fromTimestamptz(row.UpdatedAt),
		}
	}

	return services, nil
}
