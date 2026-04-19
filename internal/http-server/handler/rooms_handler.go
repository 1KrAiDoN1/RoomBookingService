package handler

import (
	"context"
	"errors"
	"internship/internal/domain"
	dto "internship/internal/models/dto/handler"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RoomsHandler struct {
	roomsService RoomsServiceInterface
	logger       *zap.Logger
}

func NewRoomsHandler(roomsService RoomsServiceInterface,
	logger *zap.Logger) *RoomsHandler {
	return &RoomsHandler{
		roomsService: roomsService,
		logger:       logger,
	}
}

// ListRooms godoc
// @Summary      Список комнат
// @Description  Возвращает список всех доступных комнат
// @Tags         rooms
// @Produce      json
// @Success      200  {object}  dto.ListRoomsResponse
// @Failure      500  {object}  InternalErrorResponse
// @Router       /rooms/list [get]
func (r *RoomsHandler) ListRooms(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), ContextTimeout)
	defer cancel()
	rooms, err := r.roomsService.ListRooms(ctx)
	if err != nil {
		r.logger.Error("listRooms: service error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, InternalErrorResponse{
			Code:    domain.ErrInternalError,
			Message: "internal server error",
		})
		return
	}

	items := make([]dto.RoomData, 0, len(rooms))
	for _, room := range rooms {
		items = append(items, dto.ToRoomDTO(room))
	}

	c.JSON(http.StatusOK, dto.ListRoomsResponse{Rooms: items})
}

// CreateRoom godoc
// @Summary      Создать комнату
// @Description  Добавляет новую комнату
// @Tags         rooms
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateRoomRequest true "Данные комнаты"
// @Success      201  {object}  dto.CreateRoomResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  InternalErrorResponse
// @Router       /rooms/create [post]
func (r *RoomsHandler) CreateRoom(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), ContextTimeout)
	defer cancel()

	var req dto.CreateRoomRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		r.logger.Error("createRoom: bind json", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    domain.ErrInvalidRequest,
			Message: "invalid request body",
		})
		return
	}

	if req.Name == "" || req.Capacity <= 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    domain.ErrInvalidRequest,
			Message: "name and capacity are required, capacity must be greater than 0",
		})
		return
	}

	room, err := r.roomsService.CreateRoom(ctx, req.Name, req.Description, req.Capacity)
	if err != nil {
		r.logger.Error("createRoom: service error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, InternalErrorResponse{
			Code:    domain.ErrInternalError,
			Message: "internal server error",
		})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateRoomResponse{Room: dto.ToRoomDTO(room)})
}

// CreateSchedule godoc
// @Summary      Создать расписание
// @Description  Задает график работы для комнаты
// @Tags         rooms
// @Accept       json
// @Produce      json
// @Param        roomID path string true "ID комнаты"
// @Param        request body dto.CreateScheduleRequest true "Данные расписания"
// @Success      201  {object}  dto.CreateScheduleResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Failure      500  {object}  InternalErrorResponse
// @Router       /rooms/{roomID}/schedule/create [post]
func (r *RoomsHandler) CreateSchedule(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), ContextTimeout)
	defer cancel()

	roomID := c.Param("roomID")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    domain.ErrInvalidRequest,
			Message: "roomId is required",
		})
		return
	}

	var req dto.CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.logger.Error("createSchedule: bind json", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    domain.ErrInvalidRequest,
			Message: "invalid request body",
		})
		return
	}

	if len(req.DaysOfWeek) == 0 || req.StartTime == "" || req.EndTime == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    domain.ErrInvalidRequest,
			Message: "daysOfWeek, startTime and endTime are required",
		})
		return
	}

	schedule, err := r.roomsService.CreateSchedule(ctx, roomID, req.DaysOfWeek, req.StartTime, req.EndTime)
	if err != nil {
		r.logger.Error("createSchedule: service error",
			zap.String("roomID", roomID), zap.Error(err))
		if errors.Is(err, domain.ErrorRoomNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    domain.ErrRoomNotFound,
				Message: "room not found",
			})
		} else if errors.Is(err, domain.ErrorScheduleExists) {
			c.JSON(http.StatusConflict, ErrorResponse{
				Code:    domain.ErrScheduleExists,
				Message: "schedule already exists for this room",
			})
		} else if errors.Is(err, domain.ErrorInvalidRequest) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    domain.ErrInvalidRequest,
				Message: "invalid schedule data",
			})
		} else {
			c.JSON(http.StatusInternalServerError, InternalErrorResponse{
				Code:    domain.ErrInternalError,
				Message: "internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.CreateScheduleResponse{
		Schedule: dto.ToScheduleDTO(schedule),
	})
}

// ListAvailableSlots godoc
// @Summary      Доступные слоты
// @Description  Возвращает свободные слоты для комнаты на конкретную дату
// @Tags         rooms
// @Produce      json
// @Param        roomID path string true "ID комнаты"
// @Param        date query string true "Дата в формате YYYY-MM-DD"
// @Success      200  {object}  dto.ListSlotsResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  InternalErrorResponse
// @Router       /rooms/{roomID}/slots/list [get]
func (r *RoomsHandler) ListAvailableSlots(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), ContextTimeout)
	defer cancel()

	roomID := c.Param("roomID")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    domain.ErrInvalidRequest,
			Message: "roomId is required",
		})
		return
	}

	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    domain.ErrInvalidRequest,
			Message: "date query parameter is required",
		})
		return
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    domain.ErrInvalidRequest,
			Message: "invalid date format, expected YYYY-MM-DD",
		})
		return
	}

	slots, err := r.roomsService.ListAvailableSlots(ctx, roomID, date)
	if err != nil {
		r.logger.Error("listAvailableSlots: service error",
			zap.String("roomID", roomID), zap.String("date", date), zap.Error(err))
		if errors.Is(err, domain.ErrorRoomNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    domain.ErrRoomNotFound,
				Message: "room not found",
			})
		} else if errors.Is(err, domain.ErrorInvalidRequest) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    domain.ErrInvalidRequest,
				Message: "invalid date format",
			})
		} else {
			c.JSON(http.StatusInternalServerError, InternalErrorResponse{
				Code:    domain.ErrInternalError,
				Message: "internal server error",
			})
		}
		return
	}

	items := make([]dto.SlotData, 0, len(slots))
	for _, s := range slots {
		items = append(items, dto.ToSlotDTO(s))
	}

	c.JSON(http.StatusOK, dto.ListSlotsResponse{Slots: items})
}
