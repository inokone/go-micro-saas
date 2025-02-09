package role

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorer is a mock implementation of the Storer interface
type MockStorer struct {
	mock.Mock
}

func (m *MockStorer) Update(role ProfileRole) error {
	args := m.Called(role)
	return args.Error(0)
}

func (m *MockStorer) ByID(id uuid.UUID) (*Role, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Role), args.Error(1)
}

func (m *MockStorer) List() ([]Role, error) {
	args := m.Called()
	return args.Get(0).([]Role), args.Error(1)
}

func TestIsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		role     Role
		expected bool
	}{
		{
			name: "Admin role",
			role: Role{
				ID: RoleAdmin,
			},
			expected: true,
		},
		{
			name: "Customer admin role",
			role: Role{
				ID: RoleCustomerAdmin,
			},
			expected: false,
		},
		{
			name: "Customer user role",
			role: Role{
				ID: RoleCustomerUser,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.role.IsAdmin())
		})
	}
}

func TestAsProfileRole(t *testing.T) {
	roleID := uuid.New()
	testRole := &Role{
		ID:               roleID,
		AppointmentQuota: 10,
		DisplayName:      "Test Role",
	}

	profileRole := testRole.AsProfileRole()

	assert.Equal(t, roleID.String(), profileRole.ID)
	assert.Equal(t, testRole.AppointmentQuota, profileRole.AppointmentQuota)
	assert.Equal(t, testRole.DisplayName, profileRole.DisplayName)
}

func TestPredefinedRoles(t *testing.T) {
	// Test that predefined role UUIDs are valid and different
	assert.NotEqual(t, RoleAdmin, RoleCustomerAdmin)
	assert.NotEqual(t, RoleAdmin, RoleCustomerUser)
	assert.NotEqual(t, RoleCustomerAdmin, RoleCustomerUser)

	// Test that predefined role UUIDs are valid
	assert.Equal(t, "b6d0a023-86db-4480-9dd9-532a4d4b1fbb", RoleAdmin.String())
	assert.Equal(t, "3dae67da-21bd-4c1f-ac35-b3e79c4a4225", RoleCustomerAdmin.String())
	assert.Equal(t, "0d83a7d4-24e3-4dd4-9b0a-d65379225abc", RoleCustomerUser.String())
}
