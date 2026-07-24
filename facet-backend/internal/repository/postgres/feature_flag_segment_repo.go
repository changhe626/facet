package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type featureFlagSegmentRepository struct {
	pool *pgxpool.Pool
}

func NewFeatureFlagSegmentRepository(pool *pgxpool.Pool) *featureFlagSegmentRepository {
	return &featureFlagSegmentRepository{pool: pool}
}

func (r *featureFlagSegmentRepository) SetSegments(ctx context.Context, flagID string, segmentIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM feature_flag_segments WHERE feature_flag_id = $1`, flagID); err != nil {
		return fmt.Errorf("failed to clear segments: %w", err)
	}

	for _, sid := range segmentIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO feature_flag_segments (feature_flag_id, segment_id) VALUES ($1, $2)`, flagID, sid); err != nil {
			return fmt.Errorf("failed to insert segment: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *featureFlagSegmentRepository) GetSegmentIDs(ctx context.Context, flagID string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT segment_id FROM feature_flag_segments WHERE feature_flag_id = $1`, flagID)
	if err != nil {
		return nil, fmt.Errorf("failed to get segment IDs: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan segment ID: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating segment IDs: %w", err)
	}
	return ids, nil
}
