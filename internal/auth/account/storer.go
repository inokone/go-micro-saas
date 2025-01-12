package account

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

// Store is a method of the `PostgresStorer` struct. Takes a `Account` as parameter and persists it.
func (s *PostgresStorer) Store(account *Account) error {
	query := `INSERT INTO microsaas.accounts (user_id, failed_login_counter, failed_login_lock, last_failed_login, confirmation_token, confirmation_ttl, confirmed, recovery_token, recovery_ttl, last_recovery, created_at, deleted_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := s.db.Exec(
		query,
		account.UserID,
		account.FailedLoginCounter,
		account.FailedLoginLock,
		account.LastFailedLogin,
		account.ConfirmationToken,
		account.ConfirmationTTL,
		account.Confirmed,
		account.RecoveryToken,
		account.RecoveryTTL,
		account.LastRecovery,
		account.CreatedAt,
		account.DeletedAt)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}
	return nil
}

// Update is a method of the `PostgresStorer` struct. Takes a `Account` as parameter and updates it.
func (s *PostgresStorer) Update(account *Account) error {
	query := `UPDATE microsaas.accounts SET failed_login_counter = $1, failed_login_lock = $2, last_failed_login = $3, confirmation_token = $4, confirmation_ttl = $5, confirmed = $6, recovery_token = $7, recovery_ttl = $8, last_recovery = $9, created_at = $10, deleted_at = $11 WHERE user_id = $12`
	_, err := s.db.Exec(query,
		account.FailedLoginCounter,
		account.FailedLoginLock,
		account.LastFailedLogin,
		account.ConfirmationToken,
		account.ConfirmationTTL,
		account.Confirmed,
		account.RecoveryToken,
		account.RecoveryTTL,
		account.LastRecovery,
		account.CreatedAt,
		account.DeletedAt,
		account.UserID)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}
	return nil
}

// ByUser is a method of the `PostgresStorer` struct. Takes a userID as parameter to load a `Account` object from persistence.
func (s *PostgresStorer) ByUser(userID uuid.UUID) (*Account, error) {
	var account Account
	query := `SELECT user_id, failed_login_counter, failed_login_lock, last_failed_login, confirmation_token, confirmation_ttl, confirmed, recovery_token, recovery_ttl, last_recovery, created_at, deleted_at FROM microsaas.accounts WHERE user_id = $1 AND deleted_at is null`
	err := s.db.Get(&account, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account by ID: %w", err)
	}
	return &account, nil
}

// ByConfirmToken is a method of the `PostgresStorer` struct. Takes a confirmation token as parameter to load a `Account` object from persistence.
func (s *PostgresStorer) ByConfirmToken(token string) (*Account, error) {
	var account Account
	query := `SELECT user_id, failed_login_counter, failed_login_lock, last_failed_login, confirmation_token, confirmation_ttl, confirmed, recovery_token, recovery_ttl, last_recovery, created_at, deleted_at FROM microsaas.accounts WHERE confirmation_token = $1 AND deleted_at is null`
	err := s.db.Get(&account, query, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get account by ID: %w", err)
	}
	return &account, nil
}

// ByRecoveryToken is a method of the `PostgresStorer` struct. Takes a recovery token as parameter to load a `Account` object from persistence.
func (s *PostgresStorer) ByRecoveryToken(token string) (*Account, error) {
	var account Account
	query := `SELECT user_id, failed_login_counter, failed_login_lock, last_failed_login, confirmation_token, confirmation_ttl, confirmed, recovery_token, recovery_ttl, last_recovery, created_at, deleted_at FROM microsaas.accounts WHERE recovery_token = $1 AND deleted_at is null`
	err := s.db.Get(&account, query, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get account by ID: %w", err)
	}
	return &account, nil
}
