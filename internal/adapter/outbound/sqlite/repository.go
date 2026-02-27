package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JCzapla/dep-dashboard/internal/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) (*Repository, error) {
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("Applying schema: %w", err)
	}
	return &Repository{db: db}, nil
}

func (r *Repository) Save(ctx context.Context, pkg *domain.Package) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Begin Transaction error: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO packages (name, version, last_updated_at)
		 VALUES (?, ?, ?)
		 ON CONFLICT (name)
		 DO UPDATE SET last_updated_at = excluded.last_updated_at`,
		 pkg.PackageRef.Name,
		 pkg.PackageRef.Version,
		 pkg.LastUpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("Upsert error: %w", err)
	}

	packageId, err := res.LastInsertId()
	if err != nil || packageId == 0 {
		err = tx.QueryRowContext(ctx,
			`SELECT id FROM packages WHERE name = ? AND version = ?`,
			pkg.PackageRef.Name,
		 	pkg.PackageRef.Version,
		).Scan(&packageId)
		if err != nil {
			return fmt.Errorf("Fetch error: %w", err)
		}
	}
	pkg.ID = packageId

	if _, err = tx.ExecContext(ctx,
		`DELETE FROM dependency_nodes WHERE package_id = ?`, packageId,
	); err != nil {
		return fmt.Errorf("Delete old nodes error: %w", err)
	}

	for _, node := range pkg.Dependencies {
		if _, err = tx.ExecContext(ctx,
		`INSERT INTO dependency_nodes (package_id, name, version, relation, score)
		 VALUES (?, ?, ?, ?, ?)`,
		 packageId,
		 node.Name,
		 node.Version,
		 node.Relation,
		 node.Score,
		); err != nil {
			return fmt.Errorf("Insert node error: %w", err)
		}
	}

	return tx.Commit()
}

func (r *Repository) scanPackageWithDeps(ctx context.Context, row *sql.Row, filters []domain.Filter) (*domain.Package, error) {
	pkg := &domain.Package{}
	var lastUpdatedAtStr string

	err := row.Scan(
		&pkg.ID,
		&pkg.PackageRef.Name,
		&pkg.PackageRef.Version,
		&lastUpdatedAtStr,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("Query error: %w", err)
	}

	pkg.LastUpdatedAt, err = time.Parse(time.RFC3339, lastUpdatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("Update time parse error: %w", err)
	}

	filterClause, filterArgs, err := buildFilters(filters)
	if err != nil {
		return nil, fmt.Errorf("Invalid filter error: %w", err)
	}
	nodesQuery := `SELECT name, version, relation, score
		 FROM dependency_nodes
		 WHERE package_id = ?` + filterClause
	nodesArgs := append([]any{pkg.ID}, filterArgs...)
	rows, err := r.db.QueryContext(ctx, nodesQuery, nodesArgs...)
	if err != nil {
		return nil, fmt.Errorf("Query nodes error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var node domain.DependencyNode
		if err := rows.Scan(&node.Name, &node.Version, &node.Relation, &node.Score); err != nil {
			return nil, fmt.Errorf("Node scan error: %w", err)
		}
		pkg.Dependencies = append(pkg.Dependencies, node)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Node iteration error: %w", err)
	}

	return pkg, nil
}

func (r *Repository) GetCurrent(ctx context.Context, filters []domain.Filter) (*domain.Package, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name,version, last_updated_at
		 FROM packages
		 LIMIT 1`,
	)
	return r.scanPackageWithDeps(ctx, row, filters)
}

func (r *Repository) DeleteByName(ctx context.Context, name string) error {
	query := "DELETE FROM packages WHERE name = ?"
	res, err := r.db.ExecContext(ctx, query, name)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return domain.ErrNotFound
	}
	return nil
}

var allowedColumns = map[string]string {
	"name": "name",
	"score": "score",
}

func buildFilters(filters []domain.Filter) (string, []any, error) {
	if len(filters) == 0 {
		return "", nil, nil
	}

	clauses := make([]string, 0, len(filters))
	args := make([]any, 0, len(filters))

	for _, f := range filters {
		col, ok := allowedColumns[f.Column]
		if !ok {
			return "", nil, fmt.Errorf("Unknown filter error: %q", f.Column)
		}
		switch f.Operator {
		case domain.FilterEq:
			clauses = append(clauses, fmt.Sprintf("%s = ?", col))
			args = append(args, f.Value)
		case domain.FilterGte:
			val, err := strconv.ParseFloat(f.Value, 64)
			if err != nil {
				return "", nil, fmt.Errorf("Minimum filter value must be numeric: %q", f.Value)
			}
			clauses = append(clauses, fmt.Sprintf("%s >= ?", col))
			args = append(args, val)
		default:
			return "", nil, fmt.Errorf("Unknown operator error: %q", f.Operator)
		}
	}
	return " AND " + strings.Join(clauses, " AND "), args, nil
}