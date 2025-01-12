package user

import (
	"time"

	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/inokone/go-micro-saas/internal/auth/role"
	"golang.org/x/crypto/bcrypt"
)

// User is the user representation for database storage.
type User struct {
	ID        uuid.UUID `db:"user_id"`
	Email     string    `db:"email"`
	PassHash  string    `db:"pass_hash"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Role      *role.Role
	Source    string    `db:"source"`
	Enabled   bool      `db:"enabled"`
	RoleID    uuid.UUID `db:"role_id"`
	Status    Status    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	DeletedAt null.Time `db:"deleted_at"`
}

// Status id the accound status of the user
type Status string

const (
	// Registered is the status for finishing a signup, but not yet confirmed account
	Registered Status = "registered"
	// Confirmed is the status for having full access to the application
	Confirmed Status = "confirmed"
	// Deactivated is the status for a deleted or unregistered user
	Deactivated Status = "deactivated"
)

// NewUser is a function to create a new `User` instance, hashing the password right off the bat
func NewUser(email string, password string, firstName string, lastName string) (*User, error) {
	u := new(User)
	u.ID = uuid.New()
	u.Email = email
	u.Source = "credentials"
	u.Status = Registered
	u.Enabled = true
	u.FirstName = firstName
	u.LastName = lastName
	u.CreatedAt = time.Now()
	err := u.SetPassword(password)
	return u, err
}

// SetPassword sets the password of the target user.
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PassHash = string(hash)
	return nil
}

// VerifyPassword is a method of the `User` struct. It takes a password string as input
// and compares it with the hashed password stored in the `PassHash` field of the `User` struct.
func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PassHash), []byte(password))
	print(err)
	return err == nil
}

// IsActive is a method of `User` returning whether the user is enabled, confirmed and can store data
func (u *User) IsActive() bool {
	return u.Enabled && u.Status == Confirmed
}

// AsProfile is a method of the `User` struct. It converts a `User` object into a `Profile` object.
func (u *User) AsProfile() Profile {
	var r role.ProfileRole

	if u.Role != nil {
		r = u.Role.AsProfileRole()
	} else {
		r = role.ProfileRole{
			ID: u.RoleID.String(),
		}
	}

	return Profile{
		ID:        u.ID.String(),
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      r,
		Status:    string(u.Status),
		Source:    u.Source,
	}
}

// AsAdminView is a method of the `User` struct. It converts a `User` object into a `AdminView` object.
func (u *User) AsAdminView() AdminView {
	var d int
	if u.DeletedAt.IsZero() {
		d = 0
	} else {
		d = int(u.DeletedAt.Time.Unix())
	}
	return AdminView{
		ID:        u.ID.String(),
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Status:    string(u.Status),
		Source:    u.Source,
		Enabled:   u.Enabled,
		Role:      u.Role.AsProfileRole(),
		Created:   int(u.CreatedAt.Unix()),
		Deleted:   d,
	}
}

// Credentials is the JSON user representation for logging in with username and password
type Credentials struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required"`
	Captcha  string `json:"captcha_token" binding:"required"`
}

// Profile is the JSON user representation for authenticated users
type Profile struct {
	ID        string           `json:"id"`
	Email     string           `json:"email" binding:"required,email,max=255"`
	FirstName string           `json:"first_name" binding:"max=255"`
	LastName  string           `json:"last_name" binding:"max=255"`
	Role      role.ProfileRole `json:"role"`
	Status    string           `json:"status" binding:"max=100"`
	Source    string           `json:"source" binding:"max=100"`
}

// SignupRequest is the JSON user representation for signup process
type SignupRequest struct {
	Email     string `json:"email" binding:"required,email,max=255"`
	FirstName string `json:"first_name" binding:"max=255"`
	LastName  string `json:"last_name" binding:"max=255"`
	Password  string `json:"password" binding:"required"`
	Captcha   string `json:"captcha_token" binding:"required"`
}

// AdminView is the user representation for the admin view of the application.
type AdminView struct {
	ID        string           `json:"id"`
	Email     string           `json:"email" binding:"required,email,max=255"`
	FirstName string           `json:"first_name" binding:"max=255"`
	LastName  string           `json:"last_name" binding:"max=255"`
	Status    string           `json:"status" binding:"max=100"`
	Source    string           `json:"source" binding:"max=100"`
	Role      role.ProfileRole `json:"role"`
	Enabled   bool             `json:"enabled"`
	Created   int              `json:"created"`
	Deleted   int              `json:"deleted"`
}

// Patch is the user representation for patching an admin view of the application.
type Patch struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name" binding:"max=255"`
	LastName  string `json:"last_name" binding:"max=255"`
	//	Role      role.ProfileRole `json:"role"` TODO: fix role patching
	Enabled bool `json:"enabled"`
}

// SetEnabled is the user representation for enabling/disabling user authentication.
type SetEnabled struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
}

// RoleUser is aggregated data on the role, with the user count.
type RoleUser struct {
	Role  string `json:"role"`
	Users int    `json:"users"`
}

// Stats is aggregated data on the storer.
type Stats struct {
	TotalUsers   int
	Distribution []RoleUser
}

// Storer is the interface for `User` persistence
type Storer interface {
	Store(user *User) error
	Update(user *User) error
	Patch(usr Patch) error
	Delete(email string) error
	SetEnabled(id uuid.UUID, enabled bool) error
	ByEmail(email string) (*User, error)
	ByID(id uuid.UUID) (*User, error)
	List() ([]User, error)
	Stats() (Stats, error)
}
