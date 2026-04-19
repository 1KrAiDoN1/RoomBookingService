package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type MockConferenceClient struct {
	rng *rand.Rand
}

func NewMockConferenceClient() *MockConferenceClient {
	return &MockConferenceClient{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (m *MockConferenceClient) CreateLink(ctx context.Context, bookingID string) (string, error) {
	latency := time.Duration(20+m.rng.Intn(100)) * time.Millisecond
	if m.rng.Intn(100) < 5 {
		return "", fmt.Errorf("conference service unavailable")
	}
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(latency):
		return fmt.Sprintf("https://meet.example.com/room/%s", bookingID), nil
	}

}
