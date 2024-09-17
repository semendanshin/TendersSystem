package bid

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"tenderSystem/internal/abstraction"
	"tenderSystem/internal/domain"
	"tenderSystem/internal/domain/models"
	"time"
)

var _ abstraction.BidRepository = &PGXRepository{}

type bid struct {
	ID       uuid.UUID
	TenderID uuid.UUID

	Status string

	AuthorType string
	AuthorID   uuid.UUID

	CurrentVersionID uuid.UUID
}

type bidVersion struct {
	ID        uuid.UUID
	BidID     uuid.UUID
	Version   int
	CreatedAt time.Time

	Name        string
	Description string
}

// PGXRepository is a repository for working with bids using pgx driver
type PGXRepository struct {
	conn *pgx.Conn
}

// NewPGXRepository creates a new instance of PGXRepository
func NewPGXRepository(conn *pgx.Conn) *PGXRepository {
	return &PGXRepository{conn: conn}
}

func (P *PGXRepository) Create(ctx context.Context, data *models.Bid) (models.Bid, error) {
	const bidInsertQuery = `
		INSERT INTO bid (id, tender_id, status, author_type, author_id, current_version_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	const bidVersionInsertQuery = `
		INSERT INTO bid_version (id, bid_id, version, name, description)
		VALUES ($1, $2, $3, $4, $5)
	`

	tx, err := P.conn.Begin(ctx)
	if err != nil {
		return models.Bid{}, err
	}

	bidVersionEntity := bidVersion{
		ID:          uuid.New(),
		BidID:       uuid.UUID(data.ID),
		Version:     data.Version,
		CreatedAt:   data.CreatedAt,
		Name:        data.Name,
		Description: data.Description,
	}

	bidEntity := bid{
		ID:               uuid.UUID(data.ID),
		TenderID:         uuid.UUID(data.TenderID),
		Status:           string(data.Status),
		AuthorType:       string(data.AuthorType),
		AuthorID:         uuid.UUID(data.AuthorID),
		CurrentVersionID: bidVersionEntity.ID,
	}

	_, err = tx.Exec(ctx, bidInsertQuery, bidEntity.ID, bidEntity.TenderID, bidEntity.Status, bidEntity.AuthorType, bidEntity.AuthorID, bidEntity.CurrentVersionID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return models.Bid{}, err
	}

	_, err = tx.Exec(ctx, bidVersionInsertQuery, bidVersionEntity.ID, bidVersionEntity.BidID, bidVersionEntity.Version, bidVersionEntity.Name, bidVersionEntity.Description)
	if err != nil {
		_ = tx.Rollback(ctx)
		return models.Bid{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return models.Bid{}, err
	}

	return *data, nil
}

func (P *PGXRepository) GetByID(ctx context.Context, id models.ID) (models.Bid, error) {
	const bidSelectQuery = `
		SELECT b.id, b.tender_id, b.status, b.author_type, b.author_id, bv.name, bv.description, bv.version, b.created_at
		FROM bid b
		JOIN bid_version bv ON b.current_version_id = bv.id
		WHERE b.id = $1
	`

	row := P.conn.QueryRow(ctx, bidSelectQuery, id)

	var bid models.Bid

	err := row.Scan(&bid.ID, &bid.TenderID, &bid.Status, &bid.AuthorType, &bid.AuthorID, &bid.Name, &bid.Description, &bid.Version, &bid.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Bid{}, fmt.Errorf("bid with ID %s not found: %w", id, domain.ErrNotFound)
		}
		return models.Bid{}, err
	}

	return bid, nil
}

func (P *PGXRepository) GetAll(ctx context.Context, options ...abstraction.PaginationOptFunc) ([]models.Bid, error) {
	const bidSelectQuery = `
		SELECT b.id, b.tender_id, b.status, b.author_type, b.author_id, bv.name, bv.description, bv.version, b.created_at
		FROM bid b
		JOIN bid_version bv ON b.current_version_id = bv.id
		WHERE b.status = 'published'
		ORDER BY b.created_at DESC
		LIMIT $1 OFFSET $2
	`

	paginationOpts, err := abstraction.NewPaginationOptions(options...)
	if err != nil {
		return nil, err
	}

	rows, err := P.conn.Query(ctx, bidSelectQuery, paginationOpts.Limit, paginationOpts.Offset)
	if err != nil {
		return nil, err
	}

	var bids []models.Bid

	for rows.Next() {
		var bid models.Bid

		err := rows.Scan(&bid.ID, &bid.TenderID, &bid.Status, &bid.AuthorType, &bid.AuthorID, &bid.Name, &bid.Description, &bid.Version, &bid.CreatedAt)
		if err != nil {
			return nil, err
		}

		bids = append(bids, bid)
	}

	return bids, nil
}

func (P *PGXRepository) GetByAuthorID(ctx context.Context, authorID models.ID, options ...abstraction.PaginationOptFunc) ([]models.Bid, error) {
	const bidSelectQuery = `
		SELECT b.id, b.tender_id, b.status, b.author_type, b.author_id, bv.name, bv.description, bv.version, b.created_at
		FROM bid b
		JOIN bid_version bv ON b.current_version_id = bv.id
		WHERE b.author_id = $1	
		ORDER BY b.created_at DESC
		LIMIT $2 OFFSET $3
	`

	paginationOpts, err := abstraction.NewPaginationOptions(options...)
	if err != nil {
		return nil, err
	}

	rows, err := P.conn.Query(ctx, bidSelectQuery, authorID, paginationOpts.Limit, paginationOpts.Offset)
	if err != nil {
		return nil, err
	}

	var bids []models.Bid

	for rows.Next() {
		var bid models.Bid

		err := rows.Scan(&bid.ID, &bid.TenderID, &bid.Status, &bid.AuthorType, &bid.AuthorID, &bid.Name, &bid.Description, &bid.Version, &bid.CreatedAt)
		if err != nil {
			return nil, err
		}

		bids = append(bids, bid)
	}

	return bids, nil
}

func (P *PGXRepository) GetByTenderID(ctx context.Context, tenderID models.ID, options ...abstraction.PaginationOptFunc) ([]models.Bid, error) {
	const bidSelectQuery = `
		SELECT b.id, b.tender_id, b.status, b.author_type, b.author_id, bv.name, bv.description, bv.version, b.created_at
		FROM bid b
		JOIN bid_version bv ON b.current_version_id = bv.id
		WHERE b.tender_id = $1
		ORDER BY b.created_at DESC
		LIMIT $2 OFFSET $3
	`

	paginationOpts, err := abstraction.NewPaginationOptions(options...)
	if err != nil {
		return nil, err
	}

	rows, err := P.conn.Query(ctx, bidSelectQuery, tenderID, paginationOpts.Limit, paginationOpts.Offset)
	if err != nil {
		return nil, err
	}

	var bids []models.Bid

	for rows.Next() {
		var bid models.Bid

		err := rows.Scan(&bid.ID, &bid.TenderID, &bid.Status, &bid.AuthorType, &bid.AuthorID, &bid.Name, &bid.Description, &bid.Version, &bid.CreatedAt)
		if err != nil {
			return nil, err
		}

		bids = append(bids, bid)
	}

	return bids, nil
}

func (P *PGXRepository) SetStatus(ctx context.Context, id models.ID, status models.BidStatus) (models.Bid, error) {
	const bidUpdateQuery = `
		UPDATE bid
		SET status = $1	
		WHERE id = $2	
		RETURNING id, tender_id, status, author_type, author_id, current_version_id
	`

	row := P.conn.QueryRow(ctx, bidUpdateQuery, status, id)

	var bid models.Bid

	err := row.Scan(&bid.ID, &bid.TenderID, &bid.Status, &bid.AuthorType, &bid.AuthorID, &bid.Version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Bid{}, fmt.Errorf("bid with ID %s not found: %w", id, domain.ErrNotFound)
		}
		return models.Bid{}, err
	}

	return bid, nil
}

func (P *PGXRepository) Update(ctx context.Context, id models.ID, data *models.Bid) (models.Bid, error) {
	const bidUpdateQuery = `
		UPDATE bid
		SET current_version_id = $1
		WHERE id = $2
	`

	const bidVersionInsertQuery = `
		INSERT INTO bid_version (id, bid_id, version, name, description)
		VALUES ($1, $2, $3, $4, $5)		
	`

	tx, err := P.conn.Begin(ctx)
	if err != nil {
		return models.Bid{}, err
	}

	bidVersionEntity := bidVersion{
		ID:          uuid.New(),
		BidID:       uuid.UUID(data.ID),
		Version:     data.Version,
		CreatedAt:   time.Now(),
		Name:        data.Name,
		Description: data.Description,
	}

	_, err = tx.Exec(ctx, bidUpdateQuery, bidVersionEntity.ID, id)
	if err != nil {
		_ = tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Bid{}, fmt.Errorf("bid with ID %s not found: %w", id, domain.ErrNotFound)
		}
		return models.Bid{}, err
	}

	_, err = tx.Exec(ctx, bidVersionInsertQuery, bidVersionEntity.ID, bidVersionEntity.BidID, bidVersionEntity.Version, bidVersionEntity.Name, bidVersionEntity.Description)
	if err != nil {
		_ = tx.Rollback(ctx)
		return models.Bid{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return models.Bid{}, err
	}

	return *data, nil
}

func (P *PGXRepository) Rollback(ctx context.Context, id models.ID, version int) (models.Bid, error) {
	const bidVersionSelectQuery = `
		SELECT id, bid_id, version, name, description, created_at
		FROM bid_version
		WHERE bid_id = $1 AND version = $2
	`

	const bidUpdateQuery = `
		UPDATE bid
		SET current_version_id = $1
		WHERE id = $2
	`

	row := P.conn.QueryRow(ctx, bidVersionSelectQuery, id, version)

	var bidVersion bidVersion

	err := row.Scan(&bidVersion.ID, &bidVersion.BidID, &bidVersion.Version, &bidVersion.Name, &bidVersion.Description, &bidVersion.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Bid{}, fmt.Errorf("bid version with ID %s and version %d not found: %w", id, version, domain.ErrNotFound)
		}
		return models.Bid{}, err
	}

	tx, err := P.conn.Begin(ctx)
	if err != nil {
		return models.Bid{}, err
	}

	_, err = tx.Exec(ctx, bidUpdateQuery, bidVersion.ID, id)
	if err != nil {
		_ = tx.Rollback(ctx)
		return models.Bid{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return models.Bid{}, err
	}

	return models.Bid{
		ID:          models.ID(bidVersion.BidID),
		Name:        bidVersion.Name,
		Description: bidVersion.Description,
		Version:     bidVersion.Version,
		CreatedAt:   bidVersion.CreatedAt,
	}, nil
}

func (P *PGXRepository) GetLatestVersionNumber(ctx context.Context, id models.ID) (int, error) {
	const bidVersionSelectQuery = `
		SELECT version
		FROM bid_version
		WHERE bid_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := P.conn.QueryRow(ctx, bidVersionSelectQuery, id)

	var version int

	err := row.Scan(&version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("bid version with ID %s not found: %w", id, domain.ErrNotFound)
		}
		return 0, err
	}

	return version, nil
}
