package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JCzapla/dep-dashboard/internal/domain"
	"github.com/JCzapla/dep-dashboard/internal/port/outbound"
)

const workerLimit = 10

type DependencyService struct {
	repo outbound.Repository
	client outbound.DepsDevClient
}

func NewDependencyService(repo outbound.Repository, client outbound.DepsDevClient) *DependencyService {
	return &DependencyService{repo: repo, client: client}
}

func (s *DependencyService) StoreDependencies(ctx context.Context, name string) (*domain.Package, error) {
	oldPackage, err := s.repo.GetCurrent(ctx, nil)
	if err != nil && err != domain.ErrNotFound {
		return nil, err
	}
	

	defaultVersion, err := s.client.FetchDefaultVersion(ctx, name)
	if err != nil {
		return nil, err
	} 
	
	ref := domain.PackageRef{
		Name: name,
		Version: defaultVersion,
	}

	nodes, err := s.client.FetchDependencies(ctx, ref)
	if err != nil {
		return nil, err
	} 

	s.enrichWithScores(ctx, nodes)

	pkg := &domain.Package{
		PackageRef: ref,
		Dependencies: nodes,
		LastUpdatedAt: time.Now().UTC(),
	}

	if err := s.repo.Save(ctx, pkg); err != nil {
		return nil, fmt.Errorf("Saving package error: %w", err)
	}
	if oldPackage != nil && pkg.PackageRef.Name != oldPackage.PackageRef.Name {
		err = s.DeleteDependenciesByName(ctx, oldPackage.PackageRef.Name)
		if err != nil {
			return nil, err
		} 
	}
	return pkg, nil
}

func (s *DependencyService) GetDependencies(ctx context.Context, filters []domain.Filter) (*domain.Package, error) {
	return s.repo.GetCurrent(ctx, filters)

}

func (s *DependencyService) DeleteDependenciesByName(ctx context.Context, name string) (error) {
	return s.repo.DeleteByName(ctx, name)
}

func (s *DependencyService) enrichWithScores(ctx context.Context, nodes []domain.DependencyNode) {
	guard := make(chan struct{}, workerLimit)
	var wg sync.WaitGroup
	wg.Add(len(nodes))

	for i := range nodes {
		guard <- struct{}{}
		go func(i int) {
			defer func() {
				wg.Done()
				<-guard
			}()

			ref := domain.PackageRef{
				Name: nodes[i].Name,
				Version: nodes[i].Version,
			}
			projectKey, err := s.client.FetchProjectKey(ctx, ref)
			if err != nil || projectKey == "" {
				return
			}

			score, err := s.client.FetchScore(ctx, projectKey)
			if err != nil {
				return
			}

			nodes[i].Score = &score
		}(i)
	}
	wg.Wait()
}