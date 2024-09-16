package dto

import "tenderSystem/internal/domain/models"

//type BidResponseDTO struct {
//	ID         string
//	Name       string
//	Status     models.BidStatus
//	AuthorType models.BidAuthorType
//	AuthorID   string
//	Version    int
//	CreatedAt  time.Time
//}

type CreateBidDTO struct {
	Name        string
	Description string
	TenderID    models.ID
	AuthorID    models.ID
	AuthorType  models.BidAuthorType
}

type UpdateBidDTO struct {
	Name        *string
	Description *string
}
