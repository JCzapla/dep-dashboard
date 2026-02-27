package sqlite

import (
	"testing"

	"github.com/JCzapla/dep-dashboard/internal/domain"
)

func TestBuildFilters(t *testing.T) {
	tests := []struct {
		name string
		filters []domain.Filter
		expectedClause string
		expectedArgs []any
		Error string
	}{
		{
			name: "eq operator",
			filters: []domain.Filter{
				{Column: "name", Operator: domain.FilterEq, Value: "express"},
			},
			expectedClause: " AND name = ?",
			expectedArgs: []any{"express"},
		},
		{
			name: "gt operator",
			filters: []domain.Filter{
				{Column: "score", Operator: domain.FilterGte, Value: "5.5"},
			},
			expectedClause: " AND score >= ?",
			expectedArgs: []any{5.5},
		},
		{
			name: "injection",
			filters: []domain.Filter{
				{Column: "injected", Operator: domain.FilterEq, Value: "x"},
			},
			Error: "Unknown filter error: \"injected\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clause, args, err := buildFilters(tt.filters)
			if err != nil && err.Error() != tt.Error {
				t.Fatalf("Got error %s, expected %s", err.Error(), tt.Error)
			}
			if tt.Error != "" {
				return 
			}
			if clause != tt.expectedClause {
				t.Errorf("Got clause %s, expected %s", clause, tt.expectedClause)
			}
			if len(args) != len(tt.expectedArgs) {
				t.Fatalf("Got args len %d, expected %d", len(args), len(tt.expectedArgs))
			}
			for i, arg := range args {
				if arg != tt.expectedArgs[i] {
					t.Errorf("Got args[%d] %s, expected %v", i, arg, tt.expectedArgs[i])
				}
			}
		})
	}
}