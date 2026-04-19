package domain

import "errors"

const (
	ErrInvalidRequest    string = "INVALID_REQUEST"
	ErrUnauthorized      string = "UNAUTHORIZED"
	ErrNotFound          string = "NOT_FOUND"
	ErrRoomNotFound      string = "ROOM_NOT_FOUND"
	ErrSlotNotFound      string = "SLOT_NOT_FOUND"
	ErrSlotAlreadyBooked string = "SLOT_ALREADY_BOOKED"
	ErrBookingNotFound   string = "BOOKING_NOT_FOUND"
	ErrForbidden         string = "FORBIDDEN"
	ErrScheduleExists    string = "SCHEDULE_EXISTS"
	ErrInternalError     string = "INTERNAL_ERROR"
	ErrUserAlreadyExists string = "USER_ALREADY_EXISTS"
)

var (
	ErrorInvalidRequest    = errors.New(ErrInvalidRequest)
	ErrorUnauthorized      = errors.New(ErrUnauthorized)
	ErrorNotFound          = errors.New(ErrNotFound)
	ErrorRoomNotFound      = errors.New(ErrRoomNotFound)
	ErrorSlotNotFound      = errors.New(ErrSlotNotFound)
	ErrorSlotAlreadyBooked = errors.New(ErrSlotAlreadyBooked)
	ErrorBookingNotFound   = errors.New(ErrBookingNotFound)
	ErrorForbidden         = errors.New(ErrForbidden)
	ErrorScheduleExists    = errors.New(ErrScheduleExists)
	ErrorInternalError     = errors.New(ErrInternalError)
	ErrorUserAlreadyExists = errors.New(ErrUserAlreadyExists)
)
