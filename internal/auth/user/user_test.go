package user

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inokone/go-micro-saas/internal/auth/role"
)

// MockStorer is a mock implementation of the Storer interface
type MockStorer struct {
	mock.Mock
}

func (m *MockStorer) Store(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockStorer) Update(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockStorer) Patch(usr Patch) error {
	args := m.Called(usr)
	return args.Error(0)
}

func (m *MockStorer) Delete(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *MockStorer) SetEnabled(id uuid.UUID, enabled bool) error {
	args := m.Called(id, enabled)
	return args.Error(0)
}

func (m *MockStorer) ByEmail(email string) (*User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockStorer) ByID(id uuid.UUID) (*User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockStorer) List() ([]User, error) {
	args := m.Called()
	return args.Get(0).([]User), args.Error(1)
}

func (m *MockStorer) Stats() (Stats, error) {
	args := m.Called()
	return args.Get(0).(Stats), args.Error(1)
}

func TestNewUserSetsMembers(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		password  string
		firstName string
		lastName  string
		wantErr   bool
	}{
		{
			name:      "Valid user creation",
			email:     "test@example.com",
			password:  "password123",
			firstName: "John",
			lastName:  "Doe",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.email, tt.password, tt.firstName, tt.lastName)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.email, user.Email)
			assert.Equal(t, tt.firstName, user.FirstName)
			assert.Equal(t, tt.lastName, user.LastName)
			assert.True(t, user.Enabled)
			assert.Equal(t, Registered, user.Status)
			assert.NotEmpty(t, user.PassHash)
		})
	}
}

func TestVerifyPasswordSuccessfulForMatching(t *testing.T) {
	user := &User{}
	password := "testpassword123"

	err := user.SetPassword(password)
	assert.NoError(t, err)

	assert.True(t, user.VerifyPassword(password))
}

func TestVerifyPasswordFailsForNonMatching(t *testing.T) {
	user := &User{}
	password := "testpassword123"

	err := user.SetPassword(password)
	assert.NoError(t, err)

	assert.False(t, user.VerifyPassword("wrongpassword"))
}

func TestIsActiveReturnsTrueForEnabledAndConfirmedUser(t *testing.T) {
	tests := []struct {
		name     string
		user     User
		expected bool
	}{
		{
			name: "Active user",
			user: User{
				Enabled: true,
				Status:  Confirmed,
			},
			expected: true,
		},
		{
			name: "Disabled user",
			user: User{
				Enabled: false,
				Status:  Confirmed,
			},
			expected: false,
		},
		{
			name: "Unconfirmed user",
			user: User{
				Enabled: true,
				Status:  Registered,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.user.IsActive())
		})
	}
}

func TestAsProfileRetainsFields(t *testing.T) {
	userID := uuid.New()
	roleID := uuid.New()
	testUser := User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role: &role.Role{
			ID:          roleID,
			DisplayName: "Test Role",
		},
		Status: Confirmed,
		Source: "credentials",
	}

	profile := testUser.AsProfile()

	assert.Equal(t, userID.String(), profile.ID)
	assert.Equal(t, testUser.Email, profile.Email)
	assert.Equal(t, testUser.FirstName, profile.FirstName)
	assert.Equal(t, testUser.LastName, profile.LastName)
	assert.Equal(t, roleID.String(), profile.Role.ID)
	assert.Equal(t, testUser.Role.DisplayName, profile.Role.DisplayName)
	assert.Equal(t, string(testUser.Status), profile.Status)
	assert.Equal(t, testUser.Source, profile.Source)
}

func TestAsAdminViewRetainsFields(t *testing.T) {
	now := time.Now()
	deletedTime := null.TimeFrom(now.Add(time.Hour))
	userID := uuid.New()
	roleID := uuid.New()

	testUser := User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role: &role.Role{
			ID:          roleID,
			DisplayName: "Test Role",
		},
		Status:    Confirmed,
		Source:    "credentials",
		Enabled:   true,
		CreatedAt: now,
		DeletedAt: deletedTime,
	}

	adminView := testUser.AsAdminView()

	assert.Equal(t, userID.String(), adminView.ID)
	assert.Equal(t, testUser.Email, adminView.Email)
	assert.Equal(t, testUser.FirstName, adminView.FirstName)
	assert.Equal(t, testUser.LastName, adminView.LastName)
	assert.Equal(t, roleID.String(), adminView.Role.ID)
	assert.Equal(t, testUser.Role.DisplayName, adminView.Role.DisplayName)
	assert.Equal(t, string(testUser.Status), adminView.Status)
	assert.Equal(t, testUser.Source, adminView.Source)
	assert.Equal(t, testUser.Enabled, adminView.Enabled)
	assert.Equal(t, int(now.Unix()), adminView.Created)
	assert.Equal(t, int(deletedTime.Time.Unix()), adminView.Deleted)
}
