package models

import (
	"github.com/google/uuid"

	"fmt"
	"tenderSystem/internal/domain"
	"time"
)

type OrganizationType string

const (
	OrganizationTypeUnknown                 OrganizationType = "unknown"
	OrganizationTypeIndividualEntrepreneur  OrganizationType = "IP"
	OrganizationTypeLimitedLiabilityCompany OrganizationType = "LLC"
	OrganizationTypeJointStockCompany       OrganizationType = "JSC"
)

func (o OrganizationType) String() string {
	return string(o)
}

func NewOrganizationType(o string) (OrganizationType, error) {
	switch o {
	case "IP":
		return OrganizationTypeIndividualEntrepreneur, nil
	case "LLC":
		return OrganizationTypeLimitedLiabilityCompany, nil
	case "JSC":
		return OrganizationTypeJointStockCompany, nil
	default:
		return OrganizationTypeUnknown, fmt.Errorf("unknown organization type: %s", domain.ErrInvalidArgument)
	}
}

type ID uuid.UUID

func (i ID) String() string {
	return uuid.UUID(i).String()
}

func ParseID(i string) (ID, error) {
	id, err := uuid.Parse(i)
	if err != nil {
		return ID{}, fmt.Errorf("invalid ID: %s", domain.ErrInvalidArgument)
	}

	return ID(id), nil
}

func NewID() ID {
	return ID(uuid.New())
}

type Organization struct {
	ID          ID
	Name        string
	Description string
	Type        OrganizationType
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Employee struct {
	ID        ID
	Username  string
	FirstName string
	LastName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
