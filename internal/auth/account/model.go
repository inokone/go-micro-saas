package account

import (
	"time"

	"github.com/google/uuid"
	"github.com/guregu/null"
)

// Account is a struct to store state for authentication
type Account struct {
	UserID             uuid.UUID `db:"user_id"`
	FailedLoginCounter int       `db:"failed_login_counter"`
	FailedLoginLock    time.Time `db:"failed_login_lock"`
	LastFailedLogin    time.Time `db:"last_failed_login"`
	ConfirmationToken  string    `db:"confirmation_token"`
	ConfirmationTTL    time.Time `db:"confirmation_ttl"`
	Confirmed          bool      `db:"confirmed"`
	RecoveryToken      string    `db:"recovery_token"`
	RecoveryTTL        time.Time `db:"recovery_ttl"`
	LastRecovery       time.Time `db:"last_recovery"`
	CreatedAt          time.Time `db:"created_at"`
	DeletedAt          null.Time `db:"deleted_at"`
}

// ConfirmationResend is a struct for the message body of REST endpoint e-mail confirmation resend
type ConfirmationResend struct {
	Email string `json:"email" binding:"required,email"`
}

// Recovery is a struct for the message body of REST endpoint password reset request
type Recovery struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordReset is a struct for the message body of REST endpoint password reset request
type PasswordReset struct {
	Token    string `json:"token" binding:"required,uuid"`
	Password string `json:"password" binding:"required"`
}

// PasswordChange is a struct for the message body of REST endpoint password change
type PasswordChange struct {
	New string `json:"new" binding:"required"`
	Old string `json:"old" binding:"required"`
}

// Storer is the interface for `Account` persistence
type Storer interface {
	Store(account *Account) error
	Update(account *Account) error
	ByUser(userID uuid.UUID) (*Account, error)
	ByConfirmToken(token string) (*Account, error)
	ByRecoveryToken(token string) (*Account, error)
}
