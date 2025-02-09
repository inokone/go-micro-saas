package user

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inokone/go-micro-saas/internal/auth/role"
	"github.com/inokone/go-micro-saas/internal/common"
)

func setupTestRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestProfile(t *testing.T) {
	mockStorer := new(MockStorer)
	handler := NewHandler(mockStorer)
	router := setupTestRouter(handler)

	userID := uuid.New()
	roleID := uuid.New()
	testUser := &User{
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

	router.GET("/profile", func(c *gin.Context) {
		c.Set("user", testUser)
		handler.Profile(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/profile", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Profile
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, testUser.ID.String(), response.ID)
	assert.Equal(t, testUser.Email, response.Email)
	assert.Equal(t, testUser.FirstName, response.FirstName)
	assert.Equal(t, testUser.LastName, response.LastName)
}

func TestList(t *testing.T) {
	mockStorer := new(MockStorer)
	handler := NewHandler(mockStorer)
	router := setupTestRouter(handler)

	userID := uuid.New()
	roleID := uuid.New()
	testUsers := []User{
		{
			ID:        userID,
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Role: &role.Role{
				ID:          roleID,
				DisplayName: "Test Role",
			},
			Status:  Confirmed,
			Source:  "credentials",
			Enabled: true,
		},
	}

	mockStorer.On("List").Return(testUsers, nil)

	router.GET("/users", handler.List)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []AdminView
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, testUsers[0].ID.String(), response[0].ID)
	assert.Equal(t, testUsers[0].Email, response[0].Email)

	mockStorer.AssertExpectations(t)
}

func TestPatch(t *testing.T) {
	mockStorer := new(MockStorer)
	handler := NewHandler(mockStorer)
	router := setupTestRouter(handler)

	testPatch := Patch{
		ID:        uuid.New().String(),
		FirstName: "Updated",
		LastName:  "Name",
		Enabled:   true,
	}

	mockStorer.On("Patch", mock.MatchedBy(func(p Patch) bool {
		return p.ID == testPatch.ID &&
			p.FirstName == testPatch.FirstName &&
			p.LastName == testPatch.LastName &&
			p.Enabled == testPatch.Enabled
	})).Return(nil)

	router.PATCH("/users/:id", handler.Patch)

	body, _ := json.Marshal(testPatch)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/users/"+testPatch.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response common.StatusMessage
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "User patched!", response.Message)

	mockStorer.AssertExpectations(t)
}

func TestSetEnabled(t *testing.T) {
	mockStorer := new(MockStorer)
	handler := NewHandler(mockStorer)
	router := setupTestRouter(handler)

	userID := uuid.New()
	testEnabled := SetEnabled{
		ID:      userID.String(),
		Enabled: true,
	}

	mockStorer.On("SetEnabled", userID, true).Return(nil)

	router.POST("/users/enable", handler.SetEnabled)

	body, _ := json.Marshal(testEnabled)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users/enable", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response common.StatusMessage
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "User updated!", response.Message)

	mockStorer.AssertExpectations(t)
}
