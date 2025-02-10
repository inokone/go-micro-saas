package role

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

	"github.com/inokone/go-micro-saas/internal/common"
)

func setupTestRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestList200ForHappyPath(t *testing.T) {
	mockStorer := new(MockStorer)
	handler := NewHandler(mockStorer)
	router := setupTestRouter(handler)

	roleID := uuid.New()
	testRoles := []Role{
		{
			ID:               roleID,
			AppointmentQuota: 10,
			DisplayName:      "Test Role",
		},
	}

	mockStorer.On("List").Return(testRoles, nil)

	router.GET("/roles", handler.List)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/roles", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []ProfileRole
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, roleID.String(), response[0].ID)
	assert.Equal(t, testRoles[0].AppointmentQuota, response[0].AppointmentQuota)
	assert.Equal(t, testRoles[0].DisplayName, response[0].DisplayName)

	mockStorer.AssertExpectations(t)
}

func TestUpdate200ForHappyPath(t *testing.T) {
	mockStorer := new(MockStorer)
	handler := NewHandler(mockStorer)
	router := setupTestRouter(handler)

	roleID := uuid.New()
	testRole := ProfileRole{
		ID:               roleID.String(),
		AppointmentQuota: 20,
		DisplayName:      "Updated Role",
	}

	mockStorer.On("Update", mock.MatchedBy(func(r ProfileRole) bool {
		return r.ID == testRole.ID &&
			r.AppointmentQuota == testRole.AppointmentQuota &&
			r.DisplayName == testRole.DisplayName
	})).Return(nil)

	router.PUT("/roles/:id", handler.Update)

	body, _ := json.Marshal(testRole)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/roles/"+roleID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response common.StatusMessage
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Role patched!", response.Message)

	mockStorer.AssertExpectations(t)
}

func TestUpdate400ForInvalidInput(t *testing.T) {
	mockStorer := new(MockStorer)
	handler := NewHandler(mockStorer)
	router := setupTestRouter(handler)

	router.PUT("/roles/:id", handler.Update)

	// Test with invalid JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/roles/invalid-uuid", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response common.StatusMessage
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Malformed role data", response.Message)

	mockStorer.AssertNotCalled(t, "Update")
}
