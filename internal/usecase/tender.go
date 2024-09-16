package usecase

import (
	"context"
	"fmt"
	"tenderSystem/internal/abstraction"
	"tenderSystem/internal/domain"
	"tenderSystem/internal/domain/dto"
	"tenderSystem/internal/domain/models"
)

var _ abstraction.TenderUseCaseInterface = &TenderUseCase{}

type TenderUseCase struct {
	tenderRepo abstraction.TenderRepository

	employeeRepo abstraction.EmployeeRepository
}

func (t *TenderUseCase) SetStatus(ctx context.Context, id models.ID, username string, status models.TenderStatus) (models.Tender, error) {
	_, _, tender, err := t.authorizeUser(ctx, id, username)
	if err != nil {
		return models.Tender{}, err
	}

	_, err = t.tenderRepo.SetStatus(ctx, id, status)
	if err != nil {
		return models.Tender{}, err
	}

	return tender, nil
}

func (t *TenderUseCase) GetAll(ctx context.Context, options ...abstraction.GetTendersOptFunc) ([]models.Tender, error) {
	return t.tenderRepo.GetAll(ctx, options...)
}

func (t *TenderUseCase) Create(ctx context.Context, data *dto.CreateTenderDTO) (models.Tender, error) {
	tenderModel := models.NewTender(
		data.Name, data.Description, data.ServiceType, data.OrganizationID,
	)

	_, err := t.tenderRepo.Create(ctx, &tenderModel)
	if err != nil {
		return models.Tender{}, err
	}

	return tenderModel, nil
}

func (t *TenderUseCase) GetMy(ctx context.Context, username string, options ...abstraction.PaginationOptFunc) ([]models.Tender, error) {
	//TODO: Learn if we need to return all organization tenders or tenders created by the user
	u, err := t.employeeRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	o, err := t.employeeRepo.GetOrganization(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	return t.tenderRepo.GetByOrganizationID(ctx, o.ID, options...)
}

func (t *TenderUseCase) authorizeUser(ctx context.Context, tenderID models.ID, username string) (models.Employee, models.Organization, models.Tender, error) {

	u, err := t.employeeRepo.GetByUsername(ctx, username)
	if err != nil {
		return models.Employee{}, models.Organization{}, models.Tender{}, err
	}

	o, err := t.employeeRepo.GetOrganization(ctx, u.ID)
	if err != nil {
		return models.Employee{}, models.Organization{}, models.Tender{}, err
	}

	tender, err := t.tenderRepo.GetByID(ctx, tenderID)
	if err != nil {
		return models.Employee{}, models.Organization{}, models.Tender{}, err
	}

	if tender.OrganizationID != o.ID {
		return models.Employee{}, models.Organization{}, models.Tender{},
			fmt.Errorf("tender does not belong to the organization: %w", domain.ErrForbidden)
	}

	return u, o, tender, nil
}

func (t *TenderUseCase) GetStatus(ctx context.Context, id models.ID, username string) (models.TenderStatus, error) {
	_, _, tender, err := t.authorizeUser(ctx, id, username)
	if err != nil {
		return models.TenderStatusUnknown, err
	}

	return tender.Status, nil
}

func (t *TenderUseCase) Update(ctx context.Context, tenderID models.ID, username string, data *dto.UpdateTenderDTO) (models.Tender, error) {
	_, _, tender, err := t.authorizeUser(ctx, tenderID, username)
	if err != nil {
		return models.Tender{}, err
	}

	latestVersion, err := t.tenderRepo.GetLatestVersionNumber(ctx, tenderID)
	if err != nil {
		return models.Tender{}, err
	}

	var input struct {
		Name        string
		Description string
		ServiceType models.TenderType
	}
	{
		if data.Name != nil {
			input.Name = *data.Name
		} else {
			input.Name = tender.Name
		}

		if data.Description != nil {
			input.Description = *data.Description
		} else {
			input.Description = tender.Description
		}

		if data.ServiceType != nil {
			input.ServiceType = *data.ServiceType
		} else {
			input.ServiceType = tender.ServiceType
		}
	}

	tender.Name = input.Name
	tender.Description = input.Description
	tender.ServiceType = input.ServiceType
	tender.Version = latestVersion + 1

	return t.tenderRepo.Update(ctx, tenderID, &tender)
}

func (t *TenderUseCase) Rollback(ctx context.Context, tenderID models.ID, username string, version int) (models.Tender, error) {
	_, _, _, err := t.authorizeUser(ctx, tenderID, username)
	if err != nil {
		return models.Tender{}, err
	}

	return t.tenderRepo.Rollback(ctx, tenderID, version)
}

func NewTenderUseCase(tenderRepo abstraction.TenderRepository, employeeRepo abstraction.EmployeeRepository) *TenderUseCase {
	return &TenderUseCase{
		tenderRepo:   tenderRepo,
		employeeRepo: employeeRepo,
	}
}
