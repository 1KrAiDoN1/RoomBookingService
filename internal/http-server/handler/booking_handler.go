package handler

import (
	"context"
	"errors"
	"internship/internal/domain"
	dto "internship/internal/models/dto/handler"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BookingHandler struct {
	bookingService BookingServiceInterface
	logger         *zap.Logger
}

func NewBookingHandler(bookingService BookingServiceInterface,
	logger *zap.Logger) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
		logger:         logger,
	}
}

// CreateBooking godoc
// @Summary      Создать бронирование
// @Description  Бронирует конкретный слот для авторизованного пользователя
// @Tags         bookings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateBookingRequest true "Данные бронирования"
// @Success      201  {object}  dto.CreateBookingResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /bookings/create [post]
func (b *BookingHandler) CreateBooking(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), ContextTimeout)
	defer cancel()

	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    domain.ErrUnauthorized,
			Message: "unauthorized",
		})
		return
	}

	var req dto.CreateBookingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		b.logger.Error("createBooking: bind json", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    domain.ErrInvalidRequest,
			Message: "invalid request body",
		})
		return
	}
	if req.SlotID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    domain.ErrInvalidRequest,
			Message: "slotId is required",
		})
		return
	}

	booking, err := b.bookingService.CreateBooking(ctx, userID.(string), req.SlotID, req.CreateConferenceLink)
	if err != nil {
		b.logger.Error("createBooking: service error", zap.Error(err))
		if errors.Is(err, domain.ErrorSlotNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    domain.ErrSlotNotFound,
				Message: "slot not found",
			})
		} else if errors.Is(err, domain.ErrorSlotAlreadyBooked) {
			c.JSON(http.StatusConflict, ErrorResponse{
				Code:    domain.ErrSlotAlreadyBooked,
				Message: "slot is already booked",
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    domain.ErrInternalError,
				Message: "internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.CreateBookingResponse{
		Booking: dto.ToBookingDTO(booking),
	})

}

// ListAllBookings godoc
// @Summary      Список всех бронирований
// @Description  Возвращает все бронирования с пагинацией
// @Tags         bookings
// @Security     BearerAuth
// @Produce      json
// @Param        page query int false "Номер страницы" default(1)
// @Param        pageSize query int false "Количество на странице" default(20)
// @Success      200  {object}  dto.ListBookingsResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /bookings/list [get]
func (b *BookingHandler) ListAllBookings(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), ContextTimeout)
	defer cancel()

	page := queryIntParam(c, "page", 1)
	pageSize := queryIntParam(c, "pageSize", 20)

	if page < 1 || pageSize < 1 || pageSize > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    domain.ErrInvalidRequest,
			Message: "invalid pagination parameters",
		})
		return
	}

	bookings, pagination, err := b.bookingService.ListAllBookings(ctx, page, pageSize)
	if err != nil {
		b.logger.Error("listAllBookings: service error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    domain.ErrInternalError,
			Message: "internal server error",
		})
		return
	}

	items := make([]dto.BookingData, 0, len(bookings))
	for _, bk := range bookings {
		items = append(items, dto.ToBookingDTO(bk))
	}

	c.JSON(http.StatusOK, dto.ListBookingsResponse{
		Bookings: items,
		Pagination: dto.PaginationQuery{
			Page:     pagination.Page,
			PageSize: pagination.PageSize,
		},
	})
}

// ListMyBookings godoc
// @Summary      Мои бронирования
// @Description  Возвращает список бронирований текущего пользователя
// @Tags         bookings
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string][]dto.BookingData "Массив бронирований"
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /bookings/my [get]
func (b *BookingHandler) ListMyBookings(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), ContextTimeout)
	defer cancel()

	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    domain.ErrUnauthorized,
			Message: "unauthorized",
		})
		return
	}

	bookings, err := b.bookingService.ListMyBookings(ctx, userID.(string))
	if err != nil {
		b.logger.Error("listMyBookings: service error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    domain.ErrInternalError,
			Message: "internal server error",
		})
		return
	}

	items := make([]dto.BookingData, 0, len(bookings))
	for _, bk := range bookings {
		items = append(items, dto.ToBookingDTO(bk))
	}

	c.JSON(http.StatusOK, gin.H{"bookings": items})
}

// CancelBooking godoc
// @Summary      Отменить бронирование
// @Description  Отменяет бронирование по его ID
// @Tags         bookings
// @Security     BearerAuth
// @Produce      json
// @Param        bookingID path string true "ID бронирования"
// @Success      200  {object}  dto.CancelBookingResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /bookings/{bookingID}/cancel [post]
func (b *BookingHandler) CancelBooking(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), ContextTimeout)
	defer cancel()

	bookingID := c.Param("bookingID")
	if bookingID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    domain.ErrInvalidRequest,
			Message: "bookingId is required",
		})
		return
	}

	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    domain.ErrUnauthorized,
			Message: "unauthorized",
		})
		return
	}

	booking, err := b.bookingService.CancelBooking(ctx, bookingID, userID.(string))
	if err != nil {
		b.logger.Error("cancelBooking: service error",
			zap.String("bookingId", bookingID), zap.Error(err))
		if errors.Is(err, domain.ErrorBookingNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    domain.ErrBookingNotFound,
				Message: "booking not found",
			})
		} else if errors.Is(err, domain.ErrorForbidden) {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Code:    domain.ErrForbidden,
				Message: "cannot cancel another user's booking",
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    domain.ErrInternalError,
				Message: "internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, dto.CancelBookingResponse{
		Booking: dto.ToBookingDTO(booking),
	})
}

func queryIntParam(c *gin.Context, key string, defaultVal int) int {
	s := c.Query(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}
