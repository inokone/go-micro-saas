package role

import (
	"github.com/google/uuid"
)

var (
	// RoleAdmin is the administrator role type.
	RoleAdmin = uuid.MustParse("b6d0a023-86db-4480-9dd9-532a4d4b1fbb")
	// RoleCustomerAdmin is the administrator of a schedule app.
	RoleCustomerAdmin = uuid.MustParse("3dae67da-21bd-4c1f-ac35-b3e79c4a4225")
	// RoleCustomerUser is a user of a schedule app.
	RoleCustomerUser = uuid.MustParse("0d83a7d4-24e3-4dd4-9b0a-d65379225abc")
)

// Role is a struct representing the user role representation for database storage.
type Role struct {
	ID               uuid.UUID `db:"role_id"`
	AppointmentQuota int       `db:"appointment_quota"`
	DisplayName      string    `db:"display_name"`
}

func (r Role) IsAdmin() bool {
	return r.ID == RoleAdmin
}

// ProfileRole is a struct, the JSON representation of the `Role` entity for profile and admin views.
type ProfileRole struct {
	ID               string `json:"id"`
	AppointmentQuota int    `json:"quota"`
	DisplayName      string `json:"name" binding:"required,max=100"`
}

// AsProfileRole is a method of the `Role` struct. It converts a `Role` object into a `ProfileRole` object.
func (u *Role) AsProfileRole() ProfileRole {
	return ProfileRole{
		ID:               u.ID.String(),
		AppointmentQuota: u.AppointmentQuota,
		DisplayName:      u.DisplayName,
	}
}

// Storer is the interface for `Role` persistence
type Storer interface {
	Update(role ProfileRole) error
	ByID(id uuid.UUID) (*Role, error)
	List() ([]Role, error)
}
