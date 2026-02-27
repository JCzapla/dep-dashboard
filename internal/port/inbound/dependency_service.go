package inbound

import (
	"context"

	"github.com/JCzapla/dep-dashboard/internal/domain"
)


type DependencyService interface {
	StoreDependencies(ctx context.Context, name string) (*domain.Package, error)
	GetDependencies(ctx context.Context, filters []domain.Filter) (*domain.Package, error)
	DeleteDependenciesByName(ctx context.Context, name string) (error)	
}