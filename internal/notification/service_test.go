package notification

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inokone/go-micro-saas/internal/common"
	"github.com/inokone/go-micro-saas/internal/mail"
)

// MockMailService implements the mail.Mailer interface for testing
type MockMailService struct {
	mock.Mock
}

func (m *MockMailService) Send(r *mail.SendRequest) error {
	args := m.Called(r)
	return args.Error(0)
}

func (m *MockMailService) EmailConfirmation(recipient string, confirmationURL string) error {
	args := m.Called(recipient, confirmationURL)
	return args.Error(0)
}

func (m *MockMailService) PasswordReset(recipient string, resetURL string) error {
	args := m.Called(recipient, resetURL)
	return args.Error(0)
}

func TestNewServiceInitsMembers(t *testing.T) {
	mockMailer := new(MockMailService)
	source := make(chan common.Event)
	service := NewService(source, mockMailer)

	assert.NotNil(t, service)
	assert.Equal(t, source, service.source)
	assert.Equal(t, mockMailer, service.mailer)
}

func TestGracefulShutdownConsumesEvents(t *testing.T) {
	mockMailer := new(MockMailService)
	source := make(chan common.Event)
	service := NewService(source, mockMailer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the service
	service.Start(ctx)

	// Test handling of events
	event := common.Event{
		ID:   uuid.New(),
		Type: "test_event",
		Time: time.Now(),
		User: uuid.New(),
		Data: map[string]string{"test": "data"},
	}

	// Send event to the service
	source <- event

	// Test graceful shutdown
	cancel()
	time.Sleep(100 * time.Millisecond)

	// Assert that the source channel is empty
	assert.Empty(t, source)
}
