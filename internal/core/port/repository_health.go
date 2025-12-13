package port

import "context"

//go:generate mockgen -destination=./mocks/$GOFILE -package=mocks -source=$GOFILE

// HealthRepository defines the output port for Health checks.
type HealthRepository interface {
	CheckDatabase(ctx context.Context) (string, error)
}
