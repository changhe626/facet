package postgres

import (
	"context"
	"fmt"

	"facet-backend/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type segmentConditionRepository struct {
	pool *pgxpool.Pool
}

func NewSegmentConditionRepository(pool *pgxpool.Pool) domain.SegmentConditionRepository {
	return &segmentConditionRepository{pool: pool}
}

func (r *segmentConditionRepository) Create(ctx context.Context, cond *domain.SegmentCondition) error {
	query := `
		INSERT INTO segment_conditions (id, segment_id, attribute_key, operator, value)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query, cond.ID, cond.SegmentID, cond.AttributeKey, cond.Operator, cond.Value)
	if err != nil {
		return fmt.Errorf("failed to create segment condition: %w", err)
	}
	return nil
}

func (r *segmentConditionRepository) ListBySegmentID(ctx context.Context, segmentID string) ([]*domain.SegmentCondition, error) {
	query := `
		SELECT id, segment_id, attribute_key, operator, value
		FROM segment_conditions
		WHERE segment_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, segmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list segment conditions: %w", err)
	}
	defer rows.Close()

	var conditions []*domain.SegmentCondition
	for rows.Next() {
		var cond domain.SegmentCondition
		if err := rows.Scan(&cond.ID, &cond.SegmentID, &cond.AttributeKey, &cond.Operator, &cond.Value); err != nil {
			return nil, fmt.Errorf("failed to scan segment condition: %w", err)
		}
		conditions = append(conditions, &cond)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating segment conditions: %w", err)
	}
	return conditions, nil
}

func (r *segmentConditionRepository) DeleteBySegmentID(ctx context.Context, segmentID string) error {
	query := `DELETE FROM segment_conditions WHERE segment_id = $1`
	_, err := r.pool.Exec(ctx, query, segmentID)
	if err != nil {
		return fmt.Errorf("failed to delete segment conditions: %w", err)
	}
	return nil
}
