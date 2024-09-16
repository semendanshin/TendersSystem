package usecase

import (
	"context"
	"fmt"
	"tenderSystem/internal/abstraction"
	"tenderSystem/internal/domain"
	"tenderSystem/internal/domain/dto"
	"tenderSystem/internal/domain/models"
)

var _ abstraction.BidUseCaseInterface = &BidUseCase{}

const MaxQuorum = 3

type BidUseCase struct {
	employeeRepo    abstraction.EmployeeRepository
	tenderRepo      abstraction.TenderRepository
	bidRepo         abstraction.BidRepository
	bidFeedbackRepo abstraction.BidFeedbackRepository
	bidDecisionRepo abstraction.BidDecisionRepository
}

func NewBidUseCase(
	employeeRepo abstraction.EmployeeRepository,
	tenderRepo abstraction.TenderRepository,
	bidRepo abstraction.BidRepository,
	bidFeedbackRepo abstraction.BidFeedbackRepository,
	bidDecisionRepo abstraction.BidDecisionRepository,
) *BidUseCase {
	return &BidUseCase{
		employeeRepo:    employeeRepo,
		tenderRepo:      tenderRepo,
		bidRepo:         bidRepo,
		bidFeedbackRepo: bidFeedbackRepo,
		bidDecisionRepo: bidDecisionRepo,
	}
}

func (b *BidUseCase) checkUserIsBidsAuthor(ctx context.Context, bid models.Bid, u models.Employee) error {
	if bid.AuthorType == models.BidAuthorTypeUser {
		if bid.AuthorID != u.ID {
			return fmt.Errorf("user %s is not the author of bid %s:%w", u.Username, bid.ID, domain.ErrForbidden)
		}

		return nil
	}

	if bid.AuthorType == models.BidAuthorTypeOrganization {
		organization, err := b.employeeRepo.GetOrganization(ctx, u.ID)
		if err != nil {
			return err
		}

		if bid.AuthorID != organization.ID {
			return fmt.Errorf("organization %s is not the author of bid %s:%w", u.Username, bid.ID, domain.ErrForbidden)
		}

		return nil
	}

	return fmt.Errorf("unknown author type %s:%w", bid.AuthorType, domain.ErrInternal)
}

func (b *BidUseCase) checkUserHasAccess(ctx context.Context, bid models.Bid, u models.Employee) error {
	if bid.Status == models.BidStatusPublished {
		return nil
	}

	if err := b.checkUserIsBidsAuthor(ctx, bid, u); err != nil {
		return err
	}

	return nil
}

func (b *BidUseCase) Create(ctx context.Context, data *dto.CreateBidDTO) (models.Bid, error) {
	bidModel := models.NewBid(
		data.TenderID, data.AuthorType, data.AuthorID, data.Name, data.Description,
	)

	bid, err := b.bidRepo.Create(ctx, &bidModel)
	if err != nil {
		return models.Bid{}, err
	}

	return bid, nil
}

func (b *BidUseCase) GetMy(ctx context.Context, username string, options ...abstraction.PaginationOptFunc) ([]models.Bid, error) {
	// TODO: Figure out if we need to search only by user or by organization as well (or both)
	u, err := b.employeeRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	bids, err := b.bidRepo.GetByAuthorID(ctx, u.ID, options...)
	if err != nil {
		return nil, err
	}

	return bids, nil
}

func (b *BidUseCase) GetByTenderID(ctx context.Context, tenderID models.ID, username string, options ...abstraction.PaginationOptFunc) ([]models.Bid, error) {
	u, err := b.employeeRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	o, err := b.employeeRepo.GetOrganization(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	tender, err := b.tenderRepo.GetByID(ctx, tenderID)
	if err != nil {
		return nil, err
	}

	if tender.OrganizationID != o.ID {
		return nil, fmt.Errorf("organization %s is not the author of tender %s:%w", username, tenderID, domain.ErrForbidden)
	}

	bids, err := b.bidRepo.GetByTenderID(ctx, tenderID, options...)
	if err != nil {
		return nil, err
	}

	return bids, nil
}

func (b *BidUseCase) GetStatus(ctx context.Context, id models.ID, username string) (models.BidStatus, error) {
	u, err := b.employeeRepo.GetByUsername(ctx, username)
	if err != nil {
		return models.BidStatusUnknown, err
	}

	bid, err := b.bidRepo.GetByID(ctx, id)
	if err != nil {
		return models.BidStatusUnknown, err
	}

	err = b.checkUserHasAccess(ctx, bid, u)
	if err != nil {
		return models.BidStatusUnknown, err
	}

	return bid.Status, nil
}

func (b *BidUseCase) SetStatus(ctx context.Context, id models.ID, username string, status models.BidStatus) (models.Bid, error) {
	u, err := b.employeeRepo.GetByUsername(ctx, username)
	if err != nil {
		return models.Bid{}, err
	}

	bid, err := b.bidRepo.GetByID(ctx, id)
	if err != nil {
		return models.Bid{}, err
	}

	err = b.checkUserIsBidsAuthor(ctx, bid, u)

	bid, err = b.bidRepo.SetStatus(ctx, id, status)
	if err != nil {
		return models.Bid{}, err
	}

	return bid, nil
}

func (b *BidUseCase) Update(ctx context.Context, id models.ID, username string, data *dto.UpdateBidDTO) (models.Bid, error) {
	u, err := b.employeeRepo.GetByUsername(ctx, username)
	if err != nil {
		return models.Bid{}, err
	}

	bid, err := b.bidRepo.GetByID(ctx, id)
	if err != nil {
		return models.Bid{}, err
	}

	err = b.checkUserIsBidsAuthor(ctx, bid, u)
	if err != nil {
		return models.Bid{}, err
	}

	latestVersion, err := b.bidRepo.GetLatestVersionNumber(ctx, id)
	if err != nil {
		return models.Bid{}, err
	}

	var input struct {
		Name        string
		Description string
	}
	{
		if data.Name != nil {
			input.Name = *data.Name
		} else {
			input.Name = bid.Name
		}

		if data.Description != nil {
			input.Description = *data.Description
		} else {
			input.Description = bid.Description
		}
	}

	bid.Name = input.Name
	bid.Description = input.Description
	bid.Version = latestVersion + 1

	_, err = b.bidRepo.Update(ctx, id, &bid)
	if err != nil {
		return models.Bid{}, err
	}

	return bid, nil
}

func (b *BidUseCase) getQuorum(ctx context.Context, tenderID models.ID) (int, error) {
	employees, err := b.employeeRepo.GetByOrganizationID(ctx, tenderID)
	if err != nil {
		return 0, err
	}

	return min(MaxQuorum, len(employees)), nil
}

func validateSubmitDecision(tender models.Tender, bid models.Bid, userOrganization models.Organization) error {
	if tender.Status != models.TenderStatusPublished {
		return fmt.Errorf("tender %s is not published:%w", tender.ID, domain.ErrInvalidArgument)
	}

	if bid.Status != models.BidStatusPublished {
		return fmt.Errorf("bid %s is not published", bid.ID)
	}

	if tender.OrganizationID != userOrganization.ID {
		return fmt.Errorf("organization %s is not the author of tender %s:%w", userOrganization.Name, tender.ID, domain.ErrForbidden)
	}

	return nil
}

func (b *BidUseCase) checkTenderDecision(ctx context.Context, tender models.Tender, bid models.Bid, decision models.BidDecisionType) error {
	if decision == models.BidDecisionTypeRejected {
		bid.Status = models.BidStatusRejected
		_, err := b.bidRepo.Update(ctx, bid.ID, &bid)
		if err != nil {
			return err
		}

		return nil
	}

	quorum, err := b.getQuorum(ctx, tender.ID)
	if err != nil {
		return err
	}

	decisions, err := b.bidDecisionRepo.GetByBidID(ctx, bid.ID)
	if err != nil {
		return err
	}

	if len(decisions) >= quorum {
		bid.Status = models.BidStatusApproved
		_, err = b.bidRepo.Update(ctx, bid.ID, &bid)
		if err != nil {
			return err
		}

		tender.Status = models.TenderStatusClosed
		_, err = b.tenderRepo.Update(ctx, tender.ID, &tender)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *BidUseCase) SubmitDecision(ctx context.Context, id models.ID, username string, decision bool) (models.Bid, error) {
	// Get data
	u, err := b.employeeRepo.GetByUsername(ctx, username)
	if err != nil {
		return models.Bid{}, err
	}

	o, err := b.employeeRepo.GetOrganization(ctx, u.ID)
	if err != nil {
		return models.Bid{}, err
	}

	bid, err := b.bidRepo.GetByID(ctx, id)
	if err != nil {
		return models.Bid{}, err
	}

	tender, err := b.tenderRepo.GetByID(ctx, bid.TenderID)
	if err != nil {
		return models.Bid{}, err
	}

	// Validate
	err = validateSubmitDecision(tender, bid, o)
	if err != nil {
		return models.Bid{}, err
	}

	// Create
	var decisionType models.BidDecisionType
	if decision {
		decisionType = models.BidDecisionTypeApproved
	} else {
		decisionType = models.BidDecisionTypeRejected
	}

	decisionModel := models.NewBidDecision(bid.ID, u.ID, decisionType)

	_, err = b.bidDecisionRepo.Create(ctx, &decisionModel)
	if err != nil {
		return models.Bid{}, err
	}

	// Check if all decisions are made
	err = b.checkTenderDecision(ctx, tender, bid, decisionType)
	if err != nil {
		return models.Bid{}, err
	}

	return bid, nil
}

func (b *BidUseCase) LeaveFeedback(ctx context.Context, bidID models.ID, username string, feedback string) (models.Bid, error) {
	u, err := b.employeeRepo.GetByUsername(ctx, username)
	if err != nil {
		return models.Bid{}, err
	}

	o, err := b.employeeRepo.GetOrganization(ctx, u.ID)
	if err != nil {
		return models.Bid{}, err
	}

	bid, err := b.bidRepo.GetByID(ctx, bidID)
	if err != nil {
		return models.Bid{}, err
	}

	tender, err := b.tenderRepo.GetByID(ctx, bid.TenderID)
	if err != nil {
		return models.Bid{}, err
	}

	if tender.OrganizationID != o.ID {
		return models.Bid{}, fmt.Errorf("organization %s is not the author of tender %s:%w", username, tender.ID, domain.ErrForbidden)
	}

	feedbackModel := models.NewBidFeedback(bidID, feedback, u.ID)
	_, err = b.bidFeedbackRepo.Create(ctx, &feedbackModel)
	if err != nil {
		return models.Bid{}, err
	}

	return bid, nil
}

func (b *BidUseCase) Rollback(ctx context.Context, id models.ID, username string, version int) (models.Bid, error) {
	u, err := b.employeeRepo.GetByUsername(ctx, username)
	if err != nil {
		return models.Bid{}, err
	}

	bid, err := b.bidRepo.GetByID(ctx, id)
	if err != nil {
		return models.Bid{}, err
	}

	err = b.checkUserIsBidsAuthor(ctx, bid, u)
	if err != nil {
		return models.Bid{}, err
	}

	newBid, err := b.bidRepo.Rollback(ctx, id, version)
	if err != nil {
		return models.Bid{}, err
	}

	return newBid, nil
}

func (b *BidUseCase) validateGetAuthorsFeedback(ctx context.Context, tenderID models.ID, requesterUsername, authorUsername string) error {
	requester, err := b.employeeRepo.GetByUsername(ctx, requesterUsername)
	if err != nil {
		return err
	}

	requesterOrganization, err := b.employeeRepo.GetOrganization(ctx, requester.ID)
	if err != nil {
		return err
	}

	tender, err := b.tenderRepo.GetByID(ctx, tenderID)
	if err != nil {
		return err
	}

	if tender.OrganizationID != requesterOrganization.ID {
		return fmt.Errorf("organization %s is not the author of tender %s:%w", requesterUsername, tenderID, domain.ErrForbidden)
	}

	author, err := b.employeeRepo.GetByUsername(ctx, authorUsername)
	if err != nil {
		return err
	}

	authorOrganization, err := b.employeeRepo.GetOrganization(ctx, author.ID)
	if err != nil {
		return err
	}

	bids, err := b.bidRepo.GetByTenderID(ctx, tenderID)
	if err != nil {
		return err
	}

	var authorBids []models.Bid
	for _, bid := range bids {
		if bid.AuthorID == author.ID {
			authorBids = append(authorBids, bid)
		}
	}

	if len(authorBids) == 0 {
		return fmt.Errorf("author %s has no bids in tender %s:%w", authorUsername, tenderID, domain.ErrNotFound)
	}

	var matched bool
	for _, bid := range authorBids {
		if bid.AuthorType == models.BidAuthorTypeUser {
			if bid.AuthorID == author.ID {
				matched = true
				break
			}
		}
		if bid.AuthorType == models.BidAuthorTypeOrganization {
			if bid.AuthorID == authorOrganization.ID {
				matched = true
				break
			}
		}
	}

	if !matched {
		return fmt.Errorf("author %s has no bids in tender %s:%w", authorUsername, tenderID, domain.ErrNotFound)
	}

	return nil
}

func (b *BidUseCase) GetAuthorsFeedback(ctx context.Context, tenderID models.ID, requesterUsername, authorUsername string, options ...abstraction.PaginationOptFunc) ([]models.BidFeedback, error) {
	err := b.validateGetAuthorsFeedback(ctx, tenderID, requesterUsername, authorUsername)
	if err != nil {
		return nil, err
	}

	author, err := b.employeeRepo.GetByUsername(ctx, authorUsername)
	if err != nil {
		return nil, err
	}

	feedbacks, err := b.bidFeedbackRepo.GetByAuthorID(ctx, author.ID, options...)
	if err != nil {
		return nil, err
	}

	return feedbacks, nil
}
