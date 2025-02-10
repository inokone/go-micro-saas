package history

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/inokone/go-micro-saas/internal/auth/role"
	"github.com/inokone/go-micro-saas/internal/auth/user"
	"github.com/inokone/go-micro-saas/internal/common"
)

var testUser = &user.User{
	ID:        uuid.New(),
	Email:     "test@example.com",
	FirstName: "John",
	LastName:  "Doe",
	Role: &role.Role{
		ID:          uuid.New(),
		DisplayName: "Test Role",
	},
	Status: user.Confirmed,
	Source: "credentials",
}

func setupTestRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestListHistory200ForHappyPath(t *testing.T) {
	mockStorer := new(MockStorer)
	handler := NewHandler(mockStorer)
	router := setupTestRouter(handler)

	testEvents := []common.Event{
		{
			ID:   uuid.New(),
			Type: "test_event",
			Time: time.Now(),
			User: testUser.ID,
			Data: map[string]string{"test": "data"},
		},
	}

	mockStorer.On("List", testUser.ID, historySize).Return(testEvents, nil)

	router.GET("/history", func(c *gin.Context) {
		c.Set("user", testUser)
		handler.List(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/history", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []common.Event
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, testEvents[0].ID, response[0].ID)
	assert.Equal(t, testEvents[0].Type, response[0].Type)
	assert.Equal(t, testEvents[0].User, response[0].User)

	mockStorer.AssertExpectations(t)
}

func TestListHistory401ForInvalidUser(t *testing.T) {
	mockStorer := new(MockStorer)
	handler := NewHandler(mockStorer)
	router := setupTestRouter(handler)

	router.GET("/history", handler.List)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/history", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response common.StatusMessage
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Not authorized!", response.Message)
}

func TestListHistory404ForStorerError(t *testing.T) {
	mockStorer := new(MockStorer)
	handler := NewHandler(mockStorer)
	router := setupTestRouter(handler)

	mockStorer.On("List", testUser.ID, historySize).Return([]common.Event{}, assert.AnError)

	router.GET("/history", func(c *gin.Context) {
		c.Set("user", testUser)
		handler.List(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/history", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response common.StatusMessage
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "User history not found!", response.Message)

	mockStorer.AssertExpectations(t)
}
