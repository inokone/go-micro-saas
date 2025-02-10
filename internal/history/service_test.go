package history

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inokone/go-micro-saas/internal/common"
)

var testEvent = common.Event{
	ID:   uuid.New(),
	Type: "test_event",
	Time: time.Now(),
	User: uuid.New(),
	Data: map[string]string{"test": "data"},
}

// MockStorer is a mock implementation of the Storer interface
type MockStorer struct {
	mock.Mock
}

func (m *MockStorer) Store(event *common.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockStorer) List(usr uuid.UUID, limit int) ([]common.Event, error) {
	args := m.Called(usr, limit)
	return args.Get(0).([]common.Event), args.Error(1)
}

func TestNewServiceInitsMembers(t *testing.T) {
	mockStorer := new(MockStorer)
	source := make(chan common.Event)
	service := NewService(source, mockStorer)

	assert.NotNil(t, service)
	assert.Equal(t, source, service.source)
	assert.Equal(t, mockStorer, service.events)
}

func TestEventsFromSourceAreStored(t *testing.T) {
	mockStorer := new(MockStorer)
	source := make(chan common.Event)
	service := NewService(source, mockStorer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up mock expectations
	mockStorer.On("Store", &testEvent).Return(nil)

	// Start the service
	service.Start(ctx)

	// Send event to the service
	source <- testEvent

	// Give some time for the event to be processed
	time.Sleep(100 * time.Millisecond)
	mockStorer.AssertCalled(t, "Store", &testEvent)
}

func TestGracefulShutdownConsumesEvents(t *testing.T) {
	mockStorer := new(MockStorer)
	source := make(chan common.Event)
	service := NewService(source, mockStorer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up mock expectations
	mockStorer.On("Store", &testEvent).Return(nil)

	// Start the service
	service.Start(ctx)

	// Send event to the service
	source <- testEvent

	// Test graceful shutdown
	cancel()

	// Give some time for the event to be processed
	time.Sleep(100 * time.Millisecond)

	// Assert that the source channel is empty
	assert.Empty(t, source)
}
