package models

import (
	"fmt"
	"tenderSystem/internal/domain"
	"time"
)

type TenderStatus string

const (
	TenderStatusUnknown   TenderStatus = "unknown"
	TenderStatusCreated   TenderStatus = "created"
	TenderStatusPublished TenderStatus = "published"
	TenderStatusClosed    TenderStatus = "closed"
)

func (t TenderStatus) String() string {
	return string(t)
}

func NewTenderStatus(t string) (TenderStatus, error) {
	switch t {
	case "created":
		return TenderStatusCreated, nil
	case "published":
		return TenderStatusPublished, nil
	case "closed":
		return TenderStatusClosed, nil
	default:
		return TenderStatusUnknown, fmt.Errorf("unknown tender status: %w", domain.ErrInvalidArgument)
	}
}

type TenderType string

const (
	TenderTypeUnknown      TenderType = "unknown"
	TenderTypeConstruction TenderType = "construction"
	TenderTypeDelivery     TenderType = "delivery"
	TenderTypeManufacture  TenderType = "manufacture"
)

func (t TenderType) String() string {
	return string(t)
}

func NewTenderType(t string) (TenderType, error) {
	switch t {
	case "construction":
		return TenderTypeConstruction, nil
	case "delivery":
		return TenderTypeDelivery, nil
	case "manufacture":
		return TenderTypeManufacture, nil
	default:
		return TenderTypeUnknown, fmt.Errorf("unknown tender type: %w", domain.ErrInvalidArgument)
	}
}

type Tender struct {
	ID             ID
	Name           string
	Description    string
	Status         TenderStatus
	ServiceType    TenderType
	OrganizationID ID
	Version        int
	CreatedAt      time.Time
}

func NewTender(name, description string, serviceType TenderType, organizationID ID) Tender {
	return Tender{
		ID:             NewID(),
		Name:           name,
		Description:    description,
		ServiceType:    serviceType,
		OrganizationID: organizationID,
		Status:         TenderStatusCreated,
		Version:        1,
		CreatedAt:      time.Now(),
	}
}

func (t *Tender) Publish() {
	t.Status = TenderStatusPublished
}

func (t *Tender) Close() {
	t.Status = TenderStatusClosed
}
