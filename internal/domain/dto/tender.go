package dto

import "tenderSystem/internal/domain/models"

//type TenderResponseDTO struct {
//	ID          string
//	Name        string
//	Description string
//	Status      string
//	ServiceType string
//	Version     int
//	CreatedAt   time.Time
//}

type CreateTenderDTO struct {
	Name            string
	Description     string
	ServiceType     models.TenderType
	OrganizationID  models.ID
	CreatorUsername string
}

type UpdateTenderDTO struct {
	Name        *string
	Description *string
	ServiceType *models.TenderType
}
