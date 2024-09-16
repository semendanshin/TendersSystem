package abstraction

import (
	"context"
	"tenderSystem/internal/domain/dto"
	"tenderSystem/internal/domain/models"
)

type BidUseCaseInterface interface {
	Create(ctx context.Context, data *dto.CreateBidDTO) (models.Bid, error)
	GetMy(ctx context.Context, username string, options ...PaginationOptFunc) ([]models.Bid, error)
	GetByTenderID(ctx context.Context, tenderID models.ID, username string, options ...PaginationOptFunc) ([]models.Bid, error)
	GetStatus(ctx context.Context, id models.ID, username string) (models.BidStatus, error)
	SetStatus(ctx context.Context, id models.ID, username string, status models.BidStatus) (models.Bid, error)
	Update(ctx context.Context, id models.ID, username string, data *dto.UpdateBidDTO) (models.Bid, error)
	SubmitDecision(ctx context.Context, id models.ID, username string, decision models.BidDecisionType) (models.Bid, error)
	LeaveFeedback(ctx context.Context, id models.ID, username string, feedback string) (models.Bid, error)
	Rollback(ctx context.Context, id models.ID, username string, version int) (models.Bid, error)
	GetAuthorsFeedback(ctx context.Context, id models.ID, requesterUsername, authorUsername string, options ...PaginationOptFunc) ([]models.BidFeedback, error)
}

type BidRepository interface {
	Create(ctx context.Context, data *models.Bid) (models.Bid, error)
	GetByID(ctx context.Context, id models.ID) (models.Bid, error)
	GetAll(ctx context.Context, options ...PaginationOptFunc) ([]models.Bid, error)
	GetByAuthorID(ctx context.Context, authorID models.ID, options ...PaginationOptFunc) ([]models.Bid, error)
	GetByTenderID(ctx context.Context, tenderID models.ID, options ...PaginationOptFunc) ([]models.Bid, error)
	SetStatus(ctx context.Context, id models.ID, status models.BidStatus) (models.Bid, error)
	Update(ctx context.Context, id models.ID, data *models.Bid) (models.Bid, error)
	Rollback(ctx context.Context, id models.ID, version int) (models.Bid, error)
	GetLatestVersionNumber(ctx context.Context, id models.ID) (int, error)
}

type BidFeedbackRepository interface {
	Create(ctx context.Context, data *models.BidFeedback) (models.BidFeedback, error)
	GetByAuthorID(ctx context.Context, authorID models.ID, options ...PaginationOptFunc) ([]models.BidFeedback, error)
}

type BidDecisionRepository interface {
	Create(ctx context.Context, data *models.BidDecision) (models.BidDecision, error)
	GetByBidID(ctx context.Context, bidID models.ID) ([]models.BidDecision, error)
}
