package mail

import (
	"testing"
	"time"

	"github.com/cskr/pubsub/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/mail.v2"

	"github.com/inokone/go-micro-saas/internal/common"
)

type MockDialer struct {
	mock.Mock
}

func (m *MockDialer) DialAndSend(msg ...*mail.Message) error {
	args := m.Called(msg[0])
	return args.Error(0)
}

func setupTestService() (*Service, *MockDialer, *pubsub.PubSub[string, common.Event]) {
	config := &common.MailConfig{
		SMTPAddress:     "smtp.example.com",
		SMTPPort:        587,
		SMTPUser:        "user",
		SMTPPassword:    "pass",
		NoReplyAddress:  "noreply@example.com",
		ApplicationName: "Test App",
	}

	ps := pubsub.New[string, common.Event](0)
	service := NewService(config, ps)
	mockDialer := new(MockDialer)
	service.dialer = mockDialer
	return service, mockDialer, ps
}

func TestSend(t *testing.T) {
	service, mockDialer, _ := setupTestService()
	mockDialer.On("DialAndSend", mock.Anything).Return(nil)

	tests := []struct {
		name        string
		request     *SendRequest
		expectError bool
	}{
		{
			name: "Valid confirmation email",
			request: &SendRequest{
				UserID:    uuid.New(),
				Recipient: "test@example.com",
				Subject:   "Test Subject",
				Template:  confirmation,
				Data: templateData{
					Link: "http://example.com/confirm",
					App:  "Test App",
				},
			},
			expectError: false,
		},
		{
			name: "Valid password reset email",
			request: &SendRequest{
				UserID:    uuid.New(),
				Recipient: "test@example.com",
				Subject:   "Test Subject",
				Template:  pwdReset,
				Data: templateData{
					Link: "http://example.com/reset",
					App:  "Test App",
				},
			},
			expectError: false,
		},
		{
			name: "Invalid template",
			request: &SendRequest{
				UserID:    uuid.New(),
				Recipient: "test@example.com",
				Subject:   "Test Subject",
				Template:  "nonexistent",
				Data: templateData{
					Link: "http://example.com",
					App:  "Test App",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Send(tt.request)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	mockDialer.AssertExpectations(t)
}

func TestEmailConfirmation(t *testing.T) {
	service, mockDialer, _ := setupTestService()
	mockDialer.On("DialAndSend", mock.Anything).Return(nil)

	err := service.EmailConfirmation("test@example.com", "http://example.com/confirm")
	assert.NoError(t, err)

	mockDialer.AssertExpectations(t)
}

func TestPasswordReset(t *testing.T) {
	service, mockDialer, _ := setupTestService()
	mockDialer.On("DialAndSend", mock.Anything).Return(nil)

	err := service.PasswordReset("test@example.com", "http://example.com/reset")
	assert.NoError(t, err)

	mockDialer.AssertExpectations(t)
}

func TestHistoryEvent(t *testing.T) {
	service, mockDialer, ps := setupTestService()
	mockDialer.On("DialAndSend", mock.Anything).Return(nil)

	// Subscribe to history events
	ch := ps.Sub(common.HistoryTopic)
	defer ps.Unsub(ch)

	userID := uuid.New()
	recipient := "test@example.com"
	subject := "Test Subject"
	body := "Test Body"

	err := service.send(recipient, subject, body, userID)
	assert.NoError(t, err)

	// Check if event was published
	select {
	case event := <-ch:
		assert.Equal(t, common.EmailSent, event.Type)
		assert.Equal(t, userID, event.User)

		emailData, ok := event.Data.(common.EmailData)
		assert.True(t, ok)
		assert.Equal(t, service.config.NoReplyAddress, emailData.From)
		assert.Equal(t, recipient, emailData.To)
		assert.Equal(t, subject, emailData.Subject)
		assert.Equal(t, body, emailData.Body)
	case <-time.After(time.Second):
		t.Error("No event received")
	}

	mockDialer.AssertExpectations(t)
}
