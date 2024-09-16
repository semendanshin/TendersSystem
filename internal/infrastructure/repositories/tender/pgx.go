package tender

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"tenderSystem/internal/abstraction"
	"tenderSystem/internal/domain/models"
	"time"

	"github.com/jackc/pgx/v5"
)

var _ abstraction.TenderRepository = &PGXTenderRepository{}

type tender struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	Status         string
	CreatedAt      time.Time

	CurrentVersionID uuid.UUID
}

type tenderVersion struct {
	ID        uuid.UUID
	TenderID  uuid.UUID
	Version   int
	CreatedAt time.Time

	Name        string
	Description string
	ServiceType string
}

type PGXTenderRepository struct {
	conn *pgx.Conn
}

func NewPGXRepository(conn *pgx.Conn) *PGXTenderRepository {
	return &PGXTenderRepository{
		conn: conn,
	}
}

func (P *PGXTenderRepository) Create(ctx context.Context, data *models.Tender) (models.Tender, error) {
	const tenderQuery = `
		INSERT INTO tender (id, organization_id, status, created_at, current_version_id)
		VALUES ($1, $2, $3, $4, $5)
	`

	const tenderVersionQuery = `
		INSERT INTO tender_version (id, tender_id, version, created_at, name, description, service_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	tenderVersionEntity := tenderVersion{
		ID:          uuid.New(),
		TenderID:    uuid.UUID(data.ID),
		Version:     data.Version,
		CreatedAt:   data.CreatedAt,
		Name:        data.Name,
		Description: data.Description,
		ServiceType: string(data.ServiceType),
	}
	tenderEntity := tender{
		ID:               uuid.UUID(data.ID),
		OrganizationID:   uuid.UUID(data.OrganizationID),
		Status:           string(data.Status),
		CreatedAt:        data.CreatedAt,
		CurrentVersionID: tenderVersionEntity.ID,
	}

	transaction, err := P.conn.Begin(ctx)
	if err != nil {
		return models.Tender{}, err
	}

	_, err = transaction.Exec(ctx, tenderQuery, tenderEntity.ID, tenderEntity.OrganizationID, tenderEntity.Status, tenderEntity.CreatedAt, tenderEntity.CurrentVersionID)
	if err != nil {
		err := transaction.Rollback(ctx)
		if err != nil {
			return models.Tender{}, err
		}
		return models.Tender{}, err
	}

	_, err = transaction.Exec(ctx, tenderVersionQuery, tenderVersionEntity.ID, tenderVersionEntity.TenderID, tenderVersionEntity.Version, tenderVersionEntity.CreatedAt, tenderVersionEntity.Name, tenderVersionEntity.Description, tenderVersionEntity.ServiceType)
	if err != nil {
		err := transaction.Rollback(ctx)
		if err != nil {
			return models.Tender{}, err
		}
		return models.Tender{}, err
	}

	err = transaction.Commit(ctx)
	if err != nil {
		return models.Tender{}, err
	}

	return *data, nil
}

func (P *PGXTenderRepository) GetByID(ctx context.Context, id models.ID) (models.Tender, error) {
	const query = `
		SELECT t.id, t.organization_id, t.status, t.created_at, t.current_version_id, tv.version, tv.name, tv.description, tv.service_type
		FROM tender t
		JOIN tender_version tv ON t.current_version_id = tv.id
		WHERE t.id = $1
	`

	idUUID := uuid.UUID(id)

	row := P.conn.QueryRow(ctx, query, idUUID)

	var tenderEntity tender
	var tenderVersionEntity tenderVersion

	err := row.Scan(&tenderEntity.ID, &tenderEntity.OrganizationID, &tenderEntity.Status, &tenderEntity.CreatedAt, &tenderEntity.CurrentVersionID, &tenderVersionEntity.Version, &tenderVersionEntity.Name, &tenderVersionEntity.Description, &tenderVersionEntity.ServiceType)
	if err != nil {
		return models.Tender{}, err
	}

	tenderModel := models.Tender{
		ID:             models.ID(tenderEntity.ID),
		Name:           tenderVersionEntity.Name,
		Description:    tenderVersionEntity.Description,
		Status:         models.TenderStatus(tenderEntity.Status),
		ServiceType:    models.TenderType(tenderVersionEntity.ServiceType),
		OrganizationID: models.ID(tenderEntity.OrganizationID),
		Version:        tenderVersionEntity.Version,
		CreatedAt:      tenderEntity.CreatedAt,
	}

	return tenderModel, nil
}

func (P *PGXTenderRepository) GetAll(ctx context.Context, options ...abstraction.GetTendersOptFunc) ([]models.Tender, error) {
	const query = `
		SELECT t.id, t.organization_id, t.status, t.created_at, t.current_version_id, tv.version, tv.name, tv.description, tv.service_type
		FROM tender t
		JOIN tender_version tv ON t.current_version_id = tv.id
		WHERE tv.service_type = any($1) OR $1 = '{}'
		ORDER BY t.created_at DESC
		OFFSET $2
		LIMIT $3
	`

	getTenderOptions, err := abstraction.NewGetTendersOptions(options...)
	if err != nil {
		return nil, err
	}

	serviceTypes := make([]string, 0, len(getTenderOptions.ServiceTypes))
	for _, serviceType := range getTenderOptions.ServiceTypes {
		serviceTypes = append(serviceTypes, serviceType.String())
	}

	rows, err := P.conn.Query(ctx, query, serviceTypes, getTenderOptions.PaginationOptions.Offset, getTenderOptions.PaginationOptions.Limit)
	if err != nil {
		return nil, err
	}

	var tenders []models.Tender
	for rows.Next() {
		var tenderEntity tender
		var tenderVersionEntity tenderVersion

		err := rows.Scan(&tenderEntity.ID, &tenderEntity.OrganizationID, &tenderEntity.Status, &tenderEntity.CreatedAt, &tenderEntity.CurrentVersionID, &tenderVersionEntity.Version, &tenderVersionEntity.Name, &tenderVersionEntity.Description, &tenderVersionEntity.ServiceType)
		if err != nil {
			return nil, err
		}

		tenderModel := models.Tender{
			ID:             models.ID(tenderEntity.ID),
			Name:           tenderVersionEntity.Name,
			Description:    tenderVersionEntity.Description,
			Status:         models.TenderStatus(tenderEntity.Status),
			ServiceType:    models.TenderType(tenderVersionEntity.ServiceType),
			OrganizationID: models.ID(tenderEntity.OrganizationID),
			Version:        tenderVersionEntity.Version,
			CreatedAt:      tenderEntity.CreatedAt,
		}

		tenders = append(tenders, tenderModel)
	}

	return tenders, nil
}

func (P *PGXTenderRepository) GetByOrganizationID(ctx context.Context, authorID models.ID, options ...abstraction.PaginationOptFunc) ([]models.Tender, error) {
	const query = `
		SELECT t.id, t.organization_id, t.status, t.created_at, t.current_version_id, tv.version, tv.name, tv.description, tv.service_type
		FROM tender t
		JOIN tender_version tv ON t.current_version_id = tv.id
		WHERE t.organization_id = $1
		ORDER BY t.created_at DESC
		OFFSET $2
		LIMIT $3
	`

	authorIDUUID := uuid.UUID(authorID)

	paginationOptions, err := abstraction.NewPaginationOptions(options...)
	if err != nil {
		return nil, fmt.Errorf("error creating pagination options: %w", err)
	}

	rows, err := P.conn.Query(ctx, query, authorIDUUID, paginationOptions.Offset, paginationOptions.Limit)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}

	var tenders []models.Tender
	for rows.Next() {
		var tenderEntity tender
		var tenderVersionEntity tenderVersion

		err := rows.Scan(&tenderEntity.ID, &tenderEntity.OrganizationID, &tenderEntity.Status, &tenderEntity.CreatedAt, &tenderEntity.CurrentVersionID, &tenderVersionEntity.Version, &tenderVersionEntity.Name, &tenderVersionEntity.Description, &tenderVersionEntity.ServiceType)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		tenderModel := models.Tender{
			ID:             models.ID(tenderEntity.ID),
			Name:           tenderVersionEntity.Name,
			Description:    tenderVersionEntity.Description,
			Status:         models.TenderStatus(tenderEntity.Status),
			ServiceType:    models.TenderType(tenderVersionEntity.ServiceType),
			OrganizationID: models.ID(tenderEntity.OrganizationID),
			Version:        tenderVersionEntity.Version,
			CreatedAt:      tenderEntity.CreatedAt,
		}

		tenders = append(tenders, tenderModel)
	}

	return tenders, nil
}

func (P *PGXTenderRepository) GetLatestVersionNumber(ctx context.Context, id models.ID) (int, error) {
	const query = `
		SELECT tv.version
		FROM tender_version tv
		WHERE tv.tender_id = $1
		ORDER BY tv.version DESC
		LIMIT 1
	`

	idUUID := uuid.UUID(id)

	row := P.conn.QueryRow(ctx, query, idUUID)

	var version int

	err := row.Scan(&version)
	if err != nil {
		return 0, err
	}

	return version, nil
}

func (P *PGXTenderRepository) SetStatus(ctx context.Context, id models.ID, status models.TenderStatus) (models.Tender, error) {
	const query = `
		UPDATE tender
		SET status = $1
		WHERE id = $2
		RETURNING organization_id, created_at, current_version_id
	`

	idUUID := uuid.UUID(id)

	row := P.conn.QueryRow(ctx, query, status.String(), idUUID)

	var tenderEntity tender

	err := row.Scan(&tenderEntity.OrganizationID, &tenderEntity.CreatedAt, &tenderEntity.CurrentVersionID)
	if err != nil {
		return models.Tender{}, err
	}

	const queryVersion = `
		SELECT version, name, description, service_type
		FROM tender_version
		WHERE id = $1
	`

	row = P.conn.QueryRow(ctx, queryVersion, tenderEntity.CurrentVersionID)

	var tenderVersionEntity tenderVersion

	err = row.Scan(&tenderVersionEntity.Version, &tenderVersionEntity.Name, &tenderVersionEntity.Description, &tenderVersionEntity.ServiceType)
	if err != nil {
		return models.Tender{}, err
	}

	tenderModel := models.Tender{
		ID:             models.ID(idUUID),
		Name:           tenderVersionEntity.Name,
		Description:    tenderVersionEntity.Description,
		Status:         status,
		ServiceType:    models.TenderType(tenderVersionEntity.ServiceType),
		OrganizationID: models.ID(tenderEntity.OrganizationID),
		Version:        tenderVersionEntity.Version,
		CreatedAt:      tenderEntity.CreatedAt,
	}

	return tenderModel, nil
}

func (P *PGXTenderRepository) Update(ctx context.Context, id models.ID, data *models.Tender) (models.Tender, error) {
	const query = `
		INSERT INTO tender_version (id, tender_id, version, created_at, name, description, service_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	idUUID := uuid.UUID(id)

	tenderVersionEntity := tenderVersion{
		ID:          uuid.New(),
		TenderID:    idUUID,
		Version:     data.Version + 1,
		CreatedAt:   time.Now(),
		Name:        data.Name,
		Description: data.Description,
		ServiceType: string(data.ServiceType),
	}

	transaction, err := P.conn.Begin(ctx)
	if err != nil {
		return models.Tender{}, err
	}

	_, err = transaction.Exec(ctx, query, tenderVersionEntity.ID, tenderVersionEntity.TenderID, tenderVersionEntity.Version, tenderVersionEntity.CreatedAt, tenderVersionEntity.Name, tenderVersionEntity.Description, tenderVersionEntity.ServiceType)
	if err != nil {
		err := transaction.Rollback(ctx)
		if err != nil {
			return models.Tender{}, err
		}
		return models.Tender{}, err
	}

	const updateQuery = `
		UPDATE tender
		SET current_version_id = $1
		WHERE id = $2
	`

	_, err = transaction.Exec(ctx, updateQuery, tenderVersionEntity.ID, idUUID)
	if err != nil {
		err := transaction.Rollback(ctx)
		if err != nil {
			return models.Tender{}, err
		}
		return models.Tender{}, err
	}

	err = transaction.Commit(ctx)
	if err != nil {
		return models.Tender{}, err
	}

	tenderModel := models.Tender{
		ID:             models.ID(idUUID),
		Name:           tenderVersionEntity.Name,
		Description:    tenderVersionEntity.Description,
		Status:         data.Status,
		ServiceType:    data.ServiceType,
		OrganizationID: data.OrganizationID,
		Version:        tenderVersionEntity.Version,
		CreatedAt:      tenderVersionEntity.CreatedAt,
	}

	return tenderModel, nil
}

func (P *PGXTenderRepository) GetVersions(ctx context.Context, id models.ID, options ...abstraction.PaginationOptFunc) ([]models.Tender, error) {
	const query = `
		SELECT tv.id, tv.tender_id, tv.version, tv.created_at, tv.name, tv.description, tv.service_type
		FROM tender_version tv
		WHERE tv.tender_id = $1
		ORDER BY tv.version DESC
		OFFSET $2
		LIMIT $3
	`

	idUUID := uuid.UUID(id)

	paginationOptions, err := abstraction.NewPaginationOptions(options...)
	if err != nil {
		return nil, err
	}

	rows, err := P.conn.Query(ctx, query, idUUID, paginationOptions.Offset, paginationOptions.Limit)
	if err != nil {
		return nil, err
	}

	var tenders []models.Tender
	for rows.Next() {
		var tenderVersionEntity tenderVersion

		err := rows.Scan(&tenderVersionEntity.ID, &tenderVersionEntity.TenderID, &tenderVersionEntity.Version, &tenderVersionEntity.CreatedAt, &tenderVersionEntity.Name, &tenderVersionEntity.Description, &tenderVersionEntity.ServiceType)
		if err != nil {
			return nil, err
		}

		tenderModel := models.Tender{
			ID:             models.ID(tenderVersionEntity.TenderID),
			Name:           tenderVersionEntity.Name,
			Description:    tenderVersionEntity.Description,
			Status:         models.TenderStatusUnknown,
			ServiceType:    models.TenderType(tenderVersionEntity.ServiceType),
			OrganizationID: models.ID(tenderVersionEntity.TenderID),
			Version:        tenderVersionEntity.Version,
			CreatedAt:      tenderVersionEntity.CreatedAt,
		}

		tenders = append(tenders, tenderModel)
	}

	return tenders, nil
}

func (P *PGXTenderRepository) GetSpecificVersion(ctx context.Context, id models.ID, version int) (models.Tender, error) {
	const query = `
		SELECT tv.id, tv.tender_id, tv.version, tv.created_at, tv.name, tv.description, tv.service_type
		FROM tender_version tv
		WHERE tv.tender_id = $1 AND tv.version = $2
	`

	idUUID := uuid.UUID(id)

	row := P.conn.QueryRow(ctx, query, idUUID, version)

	var tenderVersionEntity tenderVersion

	err := row.Scan(&tenderVersionEntity.ID, &tenderVersionEntity.TenderID, &tenderVersionEntity.Version, &tenderVersionEntity.CreatedAt, &tenderVersionEntity.Name, &tenderVersionEntity.Description, &tenderVersionEntity.ServiceType)
	if err != nil {
		return models.Tender{}, err
	}

	tenderModel := models.Tender{
		ID:             models.ID(tenderVersionEntity.TenderID),
		Name:           tenderVersionEntity.Name,
		Description:    tenderVersionEntity.Description,
		Status:         models.TenderStatusUnknown,
		ServiceType:    models.TenderType(tenderVersionEntity.ServiceType),
		OrganizationID: models.ID(tenderVersionEntity.TenderID),
		Version:        tenderVersionEntity.Version,
		CreatedAt:      tenderVersionEntity.CreatedAt,
	}

	return tenderModel, nil
}

func (P *PGXTenderRepository) Rollback(ctx context.Context, id models.ID, version int) (models.Tender, error) {
	const query = `
		SELECT tv.id, tv.tender_id, tv.version, tv.created_at, tv.name, tv.description, tv.service_type
		FROM tender_version tv
		WHERE tv.tender_id = $1 AND tv.version = $2
	`

	idUUID := uuid.UUID(id)

	row := P.conn.QueryRow(ctx, query, idUUID, version)

	var tenderVersionEntity tenderVersion

	err := row.Scan(&tenderVersionEntity.ID, &tenderVersionEntity.TenderID, &tenderVersionEntity.Version, &tenderVersionEntity.CreatedAt, &tenderVersionEntity.Name, &tenderVersionEntity.Description, &tenderVersionEntity.ServiceType)
	if err != nil {
		return models.Tender{}, err
	}

	const updateQuery = `
		UPDATE tender
		SET current_version_id = $1
		WHERE id = $2
	`

	_, err = P.conn.Exec(ctx, updateQuery, tenderVersionEntity.ID, idUUID)
	if err != nil {
		return models.Tender{}, err
	}

	tenderModel := models.Tender{
		ID:             models.ID(tenderVersionEntity.TenderID),
		Name:           tenderVersionEntity.Name,
		Description:    tenderVersionEntity.Description,
		Status:         models.TenderStatusUnknown,
		ServiceType:    models.TenderType(tenderVersionEntity.ServiceType),
		OrganizationID: models.ID(tenderVersionEntity.TenderID),
		Version:        tenderVersionEntity.Version,
		CreatedAt:      tenderVersionEntity.CreatedAt,
	}

	return tenderModel, nil
}
