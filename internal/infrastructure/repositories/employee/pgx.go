package employee

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"tenderSystem/internal/abstraction"
	"tenderSystem/internal/domain/models"
)

var _ abstraction.EmployeeRepository = &PGXRepository{}

// PGXRepository is a repository for working with employees using pgx driver
type PGXRepository struct {
	conn *pgx.Conn
}

// NewPGXRepository creates a new instance of PGXRepository
func NewPGXRepository(conn *pgx.Conn) *PGXRepository {
	return &PGXRepository{conn: conn}
}

func (P *PGXRepository) GetByUsername(ctx context.Context, username string) (models.Employee, error) {
	const query = `
		SELECT id, username, first_name, last_name, created_at, updated_at
		FROM employee
		WHERE username = $1
	`

	row := P.conn.QueryRow(ctx, query, username)

	var employee models.Employee
	err := row.Scan(&employee.ID, &employee.Username, &employee.FirstName, &employee.LastName, &employee.CreatedAt, &employee.UpdatedAt)
	if err != nil {
		return models.Employee{}, fmt.Errorf("error getting employee by username: %w", err)
	}

	return employee, nil
}

func (P *PGXRepository) GetOrganization(ctx context.Context, userID models.ID) (models.Organization, error) {
	const query = `
		SELECT o.id, o.name, o.description, o.type, o.created_at, o.updated_at
		FROM organization_responsible o_r
		JOIN organization o ON o_r.organization_id = o.id
		WHERE o_r.user_id = $1
	`

	row := P.conn.QueryRow(ctx, query, userID)

	var organization models.Organization
	var description *string

	err := row.Scan(&organization.ID, &organization.Name, &description, &organization.Type, &organization.CreatedAt, &organization.UpdatedAt)
	if err != nil {
		return models.Organization{}, fmt.Errorf("error getting organization by user ID: %w", err)
	}

	if description != nil {
		organization.Description = *description
	}

	return organization, nil
}

func (P *PGXRepository) GetByOrganizationID(ctx context.Context, organizationID models.ID) ([]models.Employee, error) {
	const query = `
		SELECT e.id, e.username, e.first_name, e.last_name, e.created_at, e.updated_at
		FROM organization_responsible o_r
		JOIN employee e ON o_r.user_id = e.id
		WHERE o_r.organization_id = $1
	`

	rows, err := P.conn.Query(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("error getting employees by organization ID: %w", err)
	}
	defer rows.Close()

	var employees []models.Employee
	for rows.Next() {
		var employee models.Employee
		err := rows.Scan(&employee.ID, &employee.Username, &employee.FirstName, &employee.LastName, &employee.CreatedAt, &employee.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning employee: %w", err)
		}

		employees = append(employees, employee)
	}

	return employees, nil
}
