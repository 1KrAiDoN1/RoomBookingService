package service

import (
	"context"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMockConferenceClient_CreateLink(t *testing.T) {
	client := NewMockConferenceClient()

	t.Run("success", func(t *testing.T) {
		bookingID := "test-booking-123"
		link, err := client.CreateLink(context.Background(), bookingID)
		require.NoError(t, err)
		require.NotEmpty(t, link)
		require.True(t, strings.Contains(link, bookingID))
	})

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := client.CreateLink(ctx, "test-booking-456")
		require.Error(t, err)
		require.Equal(t, context.Canceled, err)
	})

	t.Run("service unavailable", func(t *testing.T) {
		// This is tricky to test reliably because of the randomness.
		// We'll run it a few times to increase the chance of hitting the error.
		client_with_error := &MockConferenceClient{
			rng: rand.New(rand.NewSource(1)),
		}
		var err error
		for i := 0; i < 20; i++ {
			_, err = client_with_error.CreateLink(context.Background(), "test-booking-789")
			if err != nil {
				break
			}
		}
		// The error message is "conference service unavailable"
		require.Error(t, err)
		require.Equal(t, "conference service unavailable", err.Error())
	})

	t.Run("context timeout", func(t *testing.T) {
		client := NewMockConferenceClient()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		time.Sleep(20 * time.Millisecond) // Ensure the context is timed out

		_, err := client.CreateLink(ctx, "booking-id-for-timeout")
		require.Error(t, err)
		require.EqualError(t, err, context.DeadlineExceeded.Error())
	})
}
