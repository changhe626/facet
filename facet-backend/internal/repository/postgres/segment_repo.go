package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"facet-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type segmentRepository struct {
	pool *pgxpool.Pool
}

func NewSegmentRepository(pool *pgxpool.Pool) domain.SegmentRepository {
	return &segmentRepository{pool: pool}
}

func (r *segmentRepository) ListAndCount(ctx context.Context, params domain.ListSegmentsParams) ([]*domain.Segment, int64, error) {
	var total int64
	var whereClauses []string
	var args []interface{}
	argCount := 1

	whereClauses = append(whereClauses, fmt.Sprintf("tenant_id = $%d", argCount))
	args = append(args, params.TenantID)
	argCount++

	if params.Keyword != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argCount))
		args = append(args, "%"+params.Keyword+"%")
		argCount++
	}

	whereClauses = append(whereClauses, "deleted_at IS NULL")
	whereStr := "WHERE " + strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM segments %s", whereStr)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count segments: %w", err)
	}
	if total == 0 {
		return []*domain.Segment{}, 0, nil
	}

	listQuery := fmt.Sprintf(`
		SELECT id, tenant_id, name, description, match_type, created_by, updated_by, created_at, updated_at
		FROM segments %s
		ORDER BY updated_at DESC
		LIMIT $%d OFFSET $%d
	`, whereStr, argCount, argCount+1)

	listArgs := append(args, params.Limit, params.Offset)
	rows, err := r.pool.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list segments: %w", err)
	}
	defer rows.Close()

	var segments []*domain.Segment
	for rows.Next() {
		var seg domain.Segment
		if err := rows.Scan(
			&seg.ID, &seg.TenantID, &seg.Name, &seg.Description,
			&seg.MatchType, &seg.CreatedBy, &seg.UpdatedBy, &seg.CreatedAt, &seg.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan segment row: %w", err)
		}
		segments = append(segments, &seg)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating segment rows: %w", err)
	}
	return segments, total, nil
}

func (r *segmentRepository) Create(ctx context.Context, seg *domain.Segment) error {
	query := `
		INSERT INTO segments (id, tenant_id, name, description, match_type, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		seg.ID, seg.TenantID, seg.Name, seg.Description, seg.MatchType,
		seg.CreatedBy, seg.UpdatedBy,
	).Scan(&seg.CreatedAt, &seg.UpdatedAt)
}

func (r *segmentRepository) GetByID(ctx context.Context, id string, tenantID string) (*domain.Segment, error) {
	query := `
		SELECT id, tenant_id, name, description, match_type, created_by, updated_by, created_at, updated_at
		FROM segments WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`
	var seg domain.Segment
	err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
		&seg.ID, &seg.TenantID, &seg.Name, &seg.Description,
		&seg.MatchType, &seg.CreatedBy, &seg.UpdatedBy, &seg.CreatedAt, &seg.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSegmentNotFound
		}
		return nil, fmt.Errorf("failed to scan segment: %w", err)
	}
	return &seg, nil
}

func (r *segmentRepository) Update(ctx context.Context, seg *domain.Segment) error {
	query := `
		UPDATE segments
		SET name = $1, description = $2, match_type = $3, updated_by = $4, updated_at = NOW()
		WHERE id = $5 AND tenant_id = $6 AND deleted_at IS NULL
	`
	cmdTag, err := r.pool.Exec(ctx, query, seg.Name, seg.Description, seg.MatchType, seg.UpdatedBy, seg.ID, seg.TenantID)
	if err != nil {
		return fmt.Errorf("failed to update segment: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrSegmentNotFound
	}
	return nil
}

func (r *segmentRepository) Delete(ctx context.Context, id string, tenantID string) error {
	query := `UPDATE segments SET deleted_at = NOW() WHERE id = $1 AND tenant_id = $2`
	cmdTag, err := r.pool.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to soft delete segment: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrSegmentNotFound
	}
	return nil
}

func (r *segmentRepository) CheckNameExists(ctx context.Context, name, tenantID string) (string, error) {
	query := `SELECT id FROM segments WHERE name = $1 AND tenant_id = $2 AND deleted_at IS NULL`
	var id string
	err := r.pool.QueryRow(ctx, query, name, tenantID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("failed to check segment name: %w", err)
	}
	return id, nil
}
