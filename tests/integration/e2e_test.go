package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"internship/internal/domain/entity"
	"internship/internal/http-server/handler"
	"internship/internal/http-server/middleware"
	"internship/internal/repository/postgres"
	"internship/internal/service"
	"internship/internal/service/jwt"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"
)

type mockConfClient struct{}

func (m *mockConfClient) CreateLink(ctx context.Context, bookingID string) (string, error) {
	return "https://meet.mock/link-" + bookingID, nil
}

func setupTestRouter(t *testing.T) (*gin.Engine, *pgxpool.Pool) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("Пропуск E2E теста: не задана переменная TEST_DB_DSN")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `TRUNCATE TABLE bookings, slots, schedules, rooms, users RESTART IDENTITY CASCADE;`)
	require.NoError(t, err)

	logger := zap.NewNop()
	getter := trmpgx.DefaultCtxGetter
	trManager := manager.Must(trmpgx.NewDefaultFactory(pool))

	authRepo := postgres.NewAuthRepository(pool, getter)
	roomsRepo := postgres.NewRoomsRepository(pool, getter)
	bookingRepo := postgres.NewBookingRepository(pool, getter)

	jwtManager := jwt.NewJWTManager("test-secret-key", time.Hour)
	authSvc := service.NewAuthService(authRepo, jwtManager, logger)
	confClient := &mockConfClient{}

	roomsSvc := service.NewRoomsService(roomsRepo, trManager, logger)
	bookingSvc := service.NewBookingService(bookingRepo, roomsRepo, trManager, confClient, logger)

	authHandler := handler.NewAuthHandler(authSvc, logger)
	roomsHandler := handler.NewRoomsHandler(roomsSvc, logger)
	bookingHandler := handler.NewBookingHandler(bookingSvc, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/auth/dummy-login", authHandler.DummyLogin)

	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware(jwtManager, logger))

	protected.POST("/rooms", roomsHandler.CreateRoom)
	protected.POST("/rooms/:roomID/schedules", roomsHandler.CreateSchedule)
	protected.GET("/rooms/:roomID/slots", roomsHandler.ListAvailableSlots)

	protected.POST("/bookings", bookingHandler.CreateBooking)
	protected.POST("/bookings/:bookingID/cancel", bookingHandler.CancelBooking)

	return router, pool
}

func TestE2E_BookingScenarios(t *testing.T) {
	router, pool := setupTestRouter(t)
	defer pool.Close()

	makeRequest := func(method, url, token string, body interface{}) *httptest.ResponseRecorder {
		var reqBody []byte
		if body != nil {
			reqBody, _ = json.Marshal(body)
		}
		req, _ := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	adminW := makeRequest("POST", "/auth/dummy-login", "", map[string]string{"role": "admin"})
	require.Equal(t, http.StatusOK, adminW.Code)
	var adminTokenResp map[string]string
	json.Unmarshal(adminW.Body.Bytes(), &adminTokenResp)
	adminToken := adminTokenResp["token"]

	userW := makeRequest("POST", "/auth/dummy-login", "", map[string]string{"role": "user"})
	require.Equal(t, http.StatusOK, userW.Code)
	var userTokenResp map[string]string
	json.Unmarshal(userW.Body.Bytes(), &userTokenResp)
	userToken := userTokenResp["token"]

	roomResp := makeRequest("POST", "/rooms", adminToken, map[string]interface{}{
		"name":        "Big Conference Room",
		"description": "Room with a projector",
		"capacity":    20,
	})
	require.Equal(t, http.StatusCreated, roomResp.Code)
	var roomData struct {
		Room struct {
			ID string `json:"id"`
		} `json:"room"`
	}
	json.Unmarshal(roomResp.Body.Bytes(), &roomData)
	roomID := roomData.Room.ID
	require.NotEmpty(t, roomID)

	scheduleResp := makeRequest("POST", "/rooms/"+roomID+"/schedules", adminToken, map[string]interface{}{
		"daysOfWeek": []int{1, 2, 3, 4, 5, 6, 7},
		"startTime":  "00:00",
		"endTime":    "23:59",
	})
	require.Equal(t, http.StatusCreated, scheduleResp.Code)

	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	slotsResp := makeRequest("GET", "/rooms/"+roomID+"/slots?date="+tomorrow, adminToken, nil)
	require.Equal(t, http.StatusOK, slotsResp.Code)

	var slotsData struct {
		Slots []struct {
			ID string `json:"id"`
		} `json:"slots"`
	}
	json.Unmarshal(slotsResp.Body.Bytes(), &slotsData)
	require.NotEmpty(t, slotsData.Slots, "Ожидается как минимум 1 сгенерированный слот")
	slotID := slotsData.Slots[0].ID

	bookResp := makeRequest("POST", "/bookings", userToken, map[string]interface{}{
		"slotId":               slotID,
		"createConferenceLink": true,
	})
	require.Equal(t, http.StatusCreated, bookResp.Code)

	var bookingData struct {
		Booking entity.Booking `json:"booking"`
	}
	json.Unmarshal(bookResp.Body.Bytes(), &bookingData)
	bookingID := bookingData.Booking.ID

	require.NotEmpty(t, bookingID)
	require.Equal(t, entity.BookingStatusActive, bookingData.Booking.Status)
	require.Contains(t, bookingData.Booking.ConferenceLink, "https://meet.mock")

	bookConflictResp := makeRequest("POST", "/bookings", userToken, map[string]interface{}{
		"slotId": slotID,
	})
	require.Equal(t, http.StatusConflict, bookConflictResp.Code)

	cancelResp := makeRequest("POST", "/bookings/"+bookingID+"/cancel", userToken, nil)
	require.Equal(t, http.StatusOK, cancelResp.Code)

	var canceledData struct {
		Booking entity.Booking `json:"booking"`
	}
	json.Unmarshal(cancelResp.Body.Bytes(), &canceledData)

	require.Equal(t, entity.BookingStatusCancelled, canceledData.Booking.Status)

	slotsAgainResp := makeRequest("GET", "/rooms/"+roomID+"/slots?date="+tomorrow, adminToken, nil)
	require.Equal(t, http.StatusOK, slotsAgainResp.Code)

	var slotsAgainData struct {
		Slots []struct {
			ID string `json:"id"`
		} `json:"slots"`
	}
	json.Unmarshal(slotsAgainResp.Body.Bytes(), &slotsAgainData)

	slotIsAvailableAgain := false
	for _, s := range slotsAgainData.Slots {
		if s.ID == slotID {
			slotIsAvailableAgain = true
			break
		}
	}
	require.True(t, slotIsAvailableAgain, "Слот должен снова появиться в выдаче после отмены брони")
}
