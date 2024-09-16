package feedback

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"tenderSystem/internal/abstraction"
	"tenderSystem/internal/domain/models"
	"time"
)

var _ abstraction.BidFeedbackRepository = &PGXRepository{}

type bidFeedback struct {
	ID          uuid.UUID
	BidID       uuid.UUID
	Description string
	AuthorID    uuid.UUID
	CreatedAt   time.Time
}

// PGXRepository is a repository for working with bid feedback using pgx driver
type PGXRepository struct {
	conn *pgx.Conn
}

// NewPGXRepository creates a new instance of PGXRepository
func NewPGXRepository(conn *pgx.Conn) *PGXRepository {
	return &PGXRepository{conn: conn}
}

func (P *PGXRepository) Create(ctx context.Context, data *models.BidFeedback) (models.BidFeedback, error) {
	const query = `
		INSERT INTO bid_feedbacks (id, bid_id, description, author_id)
		VALUES ($1, $2, $3, $4)
	`

	_, err := P.conn.Exec(ctx, query, data.ID, data.BidID, data.Description, data.AuthorID)
	if err != nil {
		return models.BidFeedback{}, err
	}

	return *data, nil
}

func (P *PGXRepository) GetByAuthorID(ctx context.Context, authorID models.ID, options ...abstraction.PaginationOptFunc) ([]models.BidFeedback, error) {
	const query = `
		SELECT id, bid_id, description, author_id, created_at
		FROM bid_feedbacks
		WHERE author_id = $1
	`

	rows, err := P.conn.Query(ctx, query, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedbacks []models.BidFeedback
	for rows.Next() {
		var feedback bidFeedback
		err := rows.Scan(&feedback.ID, &feedback.BidID, &feedback.Description, &feedback.AuthorID, &feedback.CreatedAt)
		if err != nil {
			return nil, err
		}

		feedbacks = append(feedbacks, models.BidFeedback{
			ID:          models.ID(feedback.ID),
			BidID:       models.ID(feedback.BidID),
			Description: feedback.Description,
			AuthorID:    models.ID(feedback.AuthorID),
			CreatedAt:   feedback.CreatedAt,
		})
	}

	return feedbacks, nil
}
