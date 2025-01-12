package role

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// PostgresStorer is the `Storer` implementation based on pq library.
type PostgresStorer struct {
	db *sqlx.DB
}

// NewPostgresStorer creates a new `PostgresStorer` instance based on the pq library.
func NewPostgresStorer(db *sqlx.DB) *PostgresStorer {
	return &PostgresStorer{
		db: db,
	}
}

func (s *PostgresStorer) ByID(id uuid.UUID) (*Role, error) {
	var role Role
	query := `SELECT role_id, appointment_quota, display_name FROM microsaas.roles WHERE role_id = $1`
	err := s.db.Get(&role, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get role by ID: %w", err)
	}
	return &role, nil
}

// Update is a method of the `PostgresStorer` struct. Takes a `Role` and updates settings (quota and display name) for it.
func (s *PostgresStorer) Update(role ProfileRole) error {
	query := `UPDATE microsaas.roles SET appointment_quota = $1, display_name = $2 WHERE role_id = $3`
	_, err := s.db.Exec(query, role.AppointmentQuota, role.DisplayName, role.ID)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}
	return nil
}

// List is a method of the `PostgresStorer` struct. Loads all `Role` objects from persistence.
func (s *PostgresStorer) List() ([]Role, error) {
	var roles []Role
	query := `SELECT role_id, appointment_quota, display_name FROM microsaas.roles`
	err := s.db.Select(&roles, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all roles: %w", err)
	}
	return roles, nil
}
