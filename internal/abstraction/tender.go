package abstraction

import (
	"context"
	"tenderSystem/internal/domain/dto"
	"tenderSystem/internal/domain/models"
)

type GetTendersOptions struct {
	PaginationOptions *PaginationOptions
	ServiceTypes      []models.TenderType
}

type GetTendersOptFunc func(*GetTendersOptions) error

func WithPaginationOptions(paginationOptions *PaginationOptions) GetTendersOptFunc {
	return func(o *GetTendersOptions) error {
		o.PaginationOptions = paginationOptions
		return nil
	}
}

func WithServiceType(serviceType models.TenderType) GetTendersOptFunc {
	return func(o *GetTendersOptions) error {
		o.ServiceTypes = append(o.ServiceTypes, serviceType)
		return nil
	}
}

func NewGetTendersOptions(options ...GetTendersOptFunc) (*GetTendersOptions, error) {
	paginationOpts, _ := NewPaginationOptions()
	opts := &GetTendersOptions{
		PaginationOptions: paginationOpts,
		ServiceTypes:      make([]models.TenderType, 0),
	}
	for _, opt := range options {
		if err := opt(opts); err != nil {
			return nil, err
		}
	}
	return opts, nil
}

type TenderUseCaseInterface interface {
	GetAll(ctx context.Context, options ...GetTendersOptFunc) ([]models.Tender, error)
	Create(ctx context.Context, data *dto.CreateTenderDTO) (models.Tender, error)
	GetMy(ctx context.Context, username string, options ...PaginationOptFunc) ([]models.Tender, error)
	GetStatus(ctx context.Context, id models.ID, username string) (models.TenderStatus, error)
	SetStatus(ctx context.Context, id models.ID, username string, status models.TenderStatus) (models.Tender, error)
	Update(ctx context.Context, id models.ID, username string, data *dto.UpdateTenderDTO) (models.Tender, error)
	Rollback(ctx context.Context, id models.ID, username string, version int) (models.Tender, error)
}

type TenderRepository interface {
	Create(ctx context.Context, data *models.Tender) (models.Tender, error)
	GetByID(ctx context.Context, id models.ID) (models.Tender, error)
	GetAll(ctx context.Context, options ...GetTendersOptFunc) ([]models.Tender, error)
	GetByOrganizationID(ctx context.Context, organizationID models.ID, options ...PaginationOptFunc) ([]models.Tender, error)
	SetStatus(ctx context.Context, id models.ID, status models.TenderStatus) (models.Tender, error)
	Update(ctx context.Context, id models.ID, data *models.Tender) (models.Tender, error)
	GetVersions(ctx context.Context, id models.ID, options ...PaginationOptFunc) ([]models.Tender, error)
	GetSpecificVersion(ctx context.Context, id models.ID, version int) (models.Tender, error)
	Rollback(ctx context.Context, id models.ID, version int) (models.Tender, error)
	GetLatestVersionNumber(ctx context.Context, id models.ID) (int, error)
}
