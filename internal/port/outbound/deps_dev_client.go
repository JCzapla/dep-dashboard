package outbound

import (
	"context"

	"github.com/JCzapla/dep-dashboard/internal/domain"
)

type DepsDevClient interface {
	FetchDefaultVersion(ctx context.Context, name string) (string, error)
	FetchDependencies(ctx context.Context, ref domain.PackageRef) ([]domain.DependencyNode, error)
	FetchProjectKey(ctx context.Context, ref domain.PackageRef) (string, error)
	FetchScore(ctx context.Context, projectKey string) (float64, error)
}