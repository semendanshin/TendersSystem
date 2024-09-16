package decision

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"tenderSystem/internal/abstraction"
	"tenderSystem/internal/domain/models"
	"time"
)

var _ abstraction.BidDecisionRepository = &PGXRepository{}

type bidDecision struct {
	ID         uuid.UUID
	BidID      uuid.UUID
	Decision   string
	EmployeeID uuid.UUID
	CreatedAt  time.Time
}

// PGXRepository is a repository for working with bid decisions using pgx driver
type PGXRepository struct {
	conn *pgx.Conn
}

// NewPGXRepository creates a new instance of PGXRepository
func NewPGXRepository(conn *pgx.Conn) *PGXRepository {
	return &PGXRepository{conn: conn}
}

func (P *PGXRepository) Create(ctx context.Context, data *models.BidDecision) (models.BidDecision, error) {
	const query = `
		INSERT INTO bid_decision (id, bid_id, decision, employee_id)
		VALUES ($1, $2, $3, $4)
	`

	_, err := P.conn.Exec(ctx, query, data.ID, data.BidID, data.Decision, data.EmployeeID)
	if err != nil {
		return models.BidDecision{}, err
	}

	return *data, nil
}

func (P *PGXRepository) GetByBidID(ctx context.Context, bidID models.ID) ([]models.BidDecision, error) {
	const query = `
		SELECT id, bid_id, decision, employee_id, created_at
		FROM bid_decision
		WHERE bid_id = $1
	`

	rows, err := P.conn.Query(ctx, query, bidID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []models.BidDecision
	for rows.Next() {
		var decision bidDecision
		if err := rows.Scan(&decision.ID, &decision.BidID, &decision.Decision, &decision.EmployeeID, &decision.CreatedAt); err != nil {
			return nil, err
		}

		decisions = append(decisions, models.BidDecision{
			ID:         models.ID(decision.ID),
			BidID:      models.ID(decision.BidID),
			Decision:   models.BidDecisionType(decision.Decision),
			EmployeeID: models.ID(decision.EmployeeID),
			CreatedAt:  decision.CreatedAt,
		})
	}

	return decisions, nil
}
