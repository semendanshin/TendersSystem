package abstraction

import (
	"context"
	"tenderSystem/internal/domain/models"
)

type EmployeeRepository interface {
	GetByUsername(ctx context.Context, username string) (models.Employee, error)
	GetOrganization(ctx context.Context, userID models.ID) (models.Organization, error)
	GetByOrganizationID(ctx context.Context, organizationID models.ID) ([]models.Employee, error)
}
