package outbound

import (
	"context"

	"github.com/JCzapla/dep-dashboard/internal/domain"
)


type Repository interface {
	Save(ctx context.Context, pkg *domain.Package) error
	GetCurrent(ctx context.Context, filters []domain.Filter) (*domain.Package, error)
	DeleteByName(ctx context.Context, name string) (error)
}