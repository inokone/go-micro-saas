package user

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/inokone/go-micro-saas/internal/auth/role"
)

// PostgresStorer is the `Storer` implementation based on sqlx library.
type PostgresStorer struct {
	db    *sqlx.DB
	roles role.Storer
}

// NewPostgresStorer creates a new `PostgresStorer` instance based on the sqlx library.
func NewPostgresStorer(db *sqlx.DB, roles role.Storer) *PostgresStorer {
	return &PostgresStorer{
		db:    db,
		roles: roles,
	}
}

// Store is a method of the `PostgresStorer` struct. Takes a `User` as parameter and persists it.
func (s *PostgresStorer) Store(user *User) error {
	query := "INSERT INTO microsaas.users(user_id, email, pass_hash, first_name, last_name, role_id, enabled, status, source, created_at, deleted_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)"
	_, err := s.db.Exec(
		query,
		user.ID,
		user.Email,
		user.PassHash,
		user.FirstName,
		user.LastName,
		user.RoleID,
		user.Enabled,
		user.Status,
		user.Source,
		user.CreatedAt,
		user.DeletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// ByEmail is a method of the `PostgresStorer` struct. Takes an email as parameter to load a `User` object from persistence.
func (s *PostgresStorer) ByEmail(email string) (*User, error) {
	var user User
	query := `SELECT user_id, email, pass_hash, first_name, last_name, role_id, enabled, status, source, created_at, deleted_at FROM microsaas.users WHERE email = $1 AND deleted_at is null`
	if err := s.db.Get(&user, query, email); err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	if err := s.attachRole(&user); err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

// ByID is a method of the `PostgresStorer` struct. Takes an UUID as parameter to load a `User` object from persistence.
func (s *PostgresStorer) ByID(id uuid.UUID) (*User, error) {
	var user User
	query := `SELECT user_id, email, pass_hash, first_name, last_name, role_id, enabled, status, source, created_at, deleted_at FROM microsaas.users WHERE user_id = $1 AND deleted_at is null`
	err := s.db.Get(&user, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	err = s.attachRole(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &user, nil
}

func (s *PostgresStorer) attachRole(user *User) error {
	var err error
	user.Role, err = s.roles.ByID(user.RoleID)
	if err != nil {
		return fmt.Errorf("failed to get user role: %w", err)
	}
	return nil
}

// List is a method of the `PostgresStorer` struct. Loads all `User` objects from persistence.
func (s *PostgresStorer) List() ([]User, error) {
	roleMap, err := s.mapRoles()
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	var users []User
	query := `SELECT user_id, email, pass_hash, first_name, last_name, role_id, enabled, status, source, created_at, deleted_at FROM microsaas.users WHERE deleted_at is null`
	err = s.db.Select(&users, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	for i := 0; i < len(users); i++ {
		if role, ok := roleMap[users[i].RoleID.String()]; ok {
			users[i].Role = &role
		}
	}
	return users, nil
}

func (s *PostgresStorer) mapRoles() (map[string]role.Role, error) {
	roleList, err := s.roles.List()
	if err != nil {
		return nil, fmt.Errorf("failed to map roles: %w", err)
	}
	res := make(map[string]role.Role, len(roleList))
	for _, role := range roleList {
		res[role.ID.String()] = role
	}
	return res, nil
}

// Delete is a method of the `PostgresStorer` struct. Takes an email as parameter and deletes the corresponding `User` from persistence.
func (s *PostgresStorer) Delete(email string) error {
	query := `UPDATE microsaas.users SET deleted_at = $1 WHERE email = $2`
	_, err := s.db.Exec(query, time.Now(), email)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// Patch is a method of the `PostgresStorer` struct. Takes a `Patch` and updates settings for it.
func (s *PostgresStorer) Patch(usr Patch) error {
	query := `UPDATE microsaas.users SET first_name = $1, last_name = $2, enabled = $3 WHERE user_id = $4`
	_, err := s.db.Exec(query, usr.FirstName, usr.LastName, usr.Enabled, usr.ID)
	if err != nil {
		return fmt.Errorf("failed to patch user: %w", err)
	}
	return nil
}

// SetEnabled is a method of the `PostgresStorer` struct. Takes a user and updates if it is enabled.
func (s *PostgresStorer) SetEnabled(id uuid.UUID, enabled bool) error {
	query := `UPDATE microsaas.users SET enabled = $1 WHERE user_id = $2`
	_, err := s.db.Exec(query, enabled, id)
	if err != nil {
		return fmt.Errorf("failed to enable/disable user: %w", err)
	}
	return nil
}

// Update is a method of the `PostgresStorer` struct. Takes a `User` and updates it.
func (s *PostgresStorer) Update(usr *User) error {
	values := map[string]interface{}{
		"user_id":    usr.ID,
		"first_name": usr.FirstName,
		"last_name":  usr.LastName,
		"role_id":    usr.RoleID,
		"enabled":    usr.Enabled,
		"status":     usr.Status,
		"source":     usr.Source,
	}
	query := "UPDATE microsaas.users SET user_id = :user_id, first_name = :first_name, last_name = :last_name, role_id = :role_id, enabled = :enabled, status = :status, source = :source WHERE user_id = :user_id"
	_, err := s.db.NamedExec(query, values)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// Stats is a method of `PostgresStorer` for collecting aggregated data on the storer.
func (s *PostgresStorer) Stats() (Stats, error) {
	var (
		distribution []RoleUser
		count        int
	)
	rows, err := s.db.Query("SELECT r.display_name as role, count(u.user_id) as users FROM microsaas.roles r JOIN microsaas.users u on u.role_id = r.role_id and u.deleted_at is null GROUP BY r.display_name")
	if err != nil {
		return Stats{}, fmt.Errorf("failed to get stats: %w", err)
	}
	for rows.Next() {
		var role string
		var users int
		err = rows.Scan(&role, &users)
		distribution = append(distribution, RoleUser{
			Role:  role,
			Users: users,
		})
		count += users
	}

	return Stats{
		TotalUsers:   count,
		Distribution: distribution,
	}, err
}
