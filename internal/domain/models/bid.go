package models

import (
	"fmt"
	"tenderSystem/internal/domain"
	"time"
)

type BidStatus string

const (
	BidStatusUnknown   BidStatus = "unknown"
	BidStatusCreated   BidStatus = "created"
	BidStatusPublished BidStatus = "published"
	BidStatusCanceled  BidStatus = "canceled"
	BidStatusApproved  BidStatus = "approved"
	BidStatusRejected  BidStatus = "rejected"
)

func (b BidStatus) String() string {
	return string(b)
}

func NewBidStatus(b string) (BidStatus, error) {
	switch b {
	case "created":
		return BidStatusCreated, nil
	case "published":
		return BidStatusPublished, nil
	case "canceled":
		return BidStatusCanceled, nil
	case "approved":
		return BidStatusApproved, nil
	case "rejected":
		return BidStatusRejected, nil
	default:
		return BidStatusUnknown, fmt.Errorf("unknown bid status: %w", domain.ErrInvalidArgument)
	}
}

type BidAuthorType string

const (
	BidAuthorTypeUnknown      BidAuthorType = "unknown"
	BidAuthorTypeOrganization BidAuthorType = "organization"
	BidAuthorTypeUser         BidAuthorType = "user"
)

func (b BidAuthorType) String() string {
	return string(b)
}

func NewBidAuthorType(b string) (BidAuthorType, error) {
	switch b {
	case "organization":
		return BidAuthorTypeOrganization, nil
	case "user":
		return BidAuthorTypeUser, nil
	default:
		return BidAuthorTypeUnknown, fmt.Errorf("unknown bid author type: %w", domain.ErrInvalidArgument)
	}
}

type Bid struct {
	ID          ID
	TenderID    ID
	Status      BidStatus
	AuthorType  BidAuthorType
	AuthorID    ID
	Name        string
	Description string
	Version     int
	CreatedAt   time.Time
}

func NewBid(tenderID ID, authorType BidAuthorType, authorID ID, name, description string) Bid {
	return Bid{
		ID:          NewID(),
		TenderID:    tenderID,
		Status:      BidStatusCreated,
		AuthorType:  authorType,
		AuthorID:    authorID,
		Name:        name,
		Description: description,
		Version:     1,
		CreatedAt:   time.Now(),
	}
}

type BidFeedback struct {
	ID          ID
	BidID       ID
	Description string
	AuthorID    ID
	CreatedAt   time.Time
}

func NewBidFeedback(bidID ID, description string, authorID ID) BidFeedback {
	return BidFeedback{
		ID:          NewID(),
		BidID:       bidID,
		Description: description,
		AuthorID:    authorID,
		CreatedAt:   time.Now(),
	}
}

type BidDecisionType string

const (
	BidDecisionTypeUnknown  BidDecisionType = "unknown"
	BidDecisionTypeApproved BidDecisionType = "approved"
	BidDecisionTypeRejected BidDecisionType = "rejected"
)

func (b BidDecisionType) String() string {
	return string(b)
}

func NewBidDecisionType(b string) (BidDecisionType, error) {
	switch b {
	case "approved":
		return BidDecisionTypeApproved, nil
	case "rejected":
		return BidDecisionTypeRejected, nil
	default:
		return BidDecisionTypeUnknown, fmt.Errorf("unknown bid decision type: %w", domain.ErrInvalidArgument)
	}
}

type BidDecision struct {
	ID         ID
	BidID      ID
	Decision   BidDecisionType
	EmployeeID ID
	CreatedAt  time.Time
}

func NewBidDecision(bidID, employeeID ID, decision BidDecisionType) BidDecision {
	return BidDecision{
		ID:         NewID(),
		BidID:      bidID,
		Decision:   decision,
		EmployeeID: employeeID,
		CreatedAt:  time.Now(),
	}
}
