package services

import (
	"errors"
	"fmt"
	"goproxy/domain"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTrafficReporter_AddInBytes(t *testing.T) {
	mockBus := new(MockMessageBusService)
	mockBus.On("Produce", mock.Anything, mock.Anything).Return(nil)

	reporter := &TrafficReporter{
		userId:         1,
		thresholdBytes: 10,
		interval:       time.Hour,
		messageBus:     mockBus,
		lastSent:       time.Now().UTC(),
	}

	// Increase incoming bytes counter by less than threshold.
	// This should not trigger a flushing send logic, so incoming bytes counter must not be flushed.
	reporter.AddInBytes(reporter.thresholdBytes - 1)
	assert.Equal(t, reporter.thresholdBytes-1, atomic.LoadInt64(&reporter.inBytes))

	// Increase incoming bytes counter again
	// This should trigger a flushing send logic, so incoming bytes counter must be flushed.
	reporter.AddInBytes(reporter.thresholdBytes)
	assert.Equal(t, int64(0), atomic.LoadInt64(&reporter.inBytes))

	// should send event
	mockBus.AssertCalled(t, "Produce", fmt.Sprintf("%s", domain.PLAN), mock.Anything)
}

func TestTrafficReporter_AddOutBytes(t *testing.T) {
	mockBus := new(MockMessageBusService)
	mockBus.On("Produce", mock.Anything, mock.Anything).Return(nil)

	reporter := &TrafficReporter{
		userId:         1,
		thresholdBytes: 10,
		interval:       time.Hour,
		messageBus:     mockBus,
		lastSent:       time.Now().UTC(),
	}

	// Increase incoming bytes counter by less than threshold.
	// This should not trigger a flushing send logic, so incoming bytes counter must not be flushed.
	reporter.AddOutBytes(reporter.thresholdBytes - 1)
	assert.Equal(t, reporter.thresholdBytes-1, atomic.LoadInt64(&reporter.outBytes))

	// Increase outgoing bytes counter again
	// This should trigger a flushing send logic, so outgoing bytes counter must be flushed.
	reporter.AddOutBytes(reporter.thresholdBytes)
	assert.Equal(t, int64(0), atomic.LoadInt64(&reporter.outBytes))

	// should send event
	mockBus.AssertCalled(t, "Produce", fmt.Sprintf("%s", domain.PLAN), mock.Anything)
}

func TestTrafficReporter_SendIntermediate(t *testing.T) {
	mockBus := new(MockMessageBusService)
	mockBus.On("Produce", fmt.Sprintf("%s", domain.PLAN), mock.Anything).Return(nil)

	reporter := &TrafficReporter{
		userId:     1,
		messageBus: mockBus,
		lastSent:   time.Now().UTC().Add(-time.Minute),
	}
	reporter.AddInBytes(100)
	reporter.AddOutBytes(200)

	// will trigger counters flush and event producing
	reporter.SendIntermediate(100, 200)
	// was event produced?
	mockBus.AssertCalled(t, "Produce", fmt.Sprintf("%s", domain.PLAN), mock.Anything)
	// was counters flushed?
	assert.Equal(t, int64(0), atomic.LoadInt64(&reporter.outBytes))
	assert.Equal(t, int64(0), atomic.LoadInt64(&reporter.inBytes))

}

func TestTrafficReporter_ProduceTrafficConsumedEvent(t *testing.T) {
	mockBus := new(MockMessageBusService)
	mockBus.On("Produce", fmt.Sprintf("%s", domain.PLAN), mock.Anything).Return(nil)

	reporter := &TrafficReporter{
		userId:     1,
		messageBus: mockBus,
	}

	err := reporter.ProduceTrafficConsumedEvent(100, 200)
	assert.NoError(t, err)
	mockBus.AssertCalled(t, "Produce", fmt.Sprintf("%s", domain.PLAN), mock.Anything)
}

func TestTrafficReporter_ProduceTrafficConsumedEvent_Error(t *testing.T) {
	mockBus := new(MockMessageBusService)
	mockBus.On("Produce", mock.Anything, mock.Anything).Return(errors.New("produce error"))

	reporter := &TrafficReporter{
		userId:     1,
		messageBus: mockBus,
	}

	err := reporter.ProduceTrafficConsumedEvent(100, 200)
	assert.Error(t, err)
	mockBus.AssertCalled(t, "Produce", fmt.Sprintf("%s", domain.PLAN), mock.Anything)
}
