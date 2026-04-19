package service

import (
	"context"
	"fmt"
	"internship/internal/domain"
	"internship/internal/domain/entity"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type bookingService struct {
	bookingRepository BookingRepositoryInterface
	roomsRepository   RoomsRepositoryInterface
	trManager         TransactionManagerInterface
	conferenceClient  ConferenceClientInterface
	logger            *zap.Logger
}

func NewBookingService(bookingRepository BookingRepositoryInterface, roomsRepository RoomsRepositoryInterface, trManager TransactionManagerInterface, conferenceClient ConferenceClientInterface, logger *zap.Logger) *bookingService {
	return &bookingService{
		bookingRepository: bookingRepository,
		roomsRepository:   roomsRepository,
		trManager:         trManager,
		conferenceClient:  conferenceClient,
		logger:            logger,
	}
}

func (s *bookingService) CreateBooking(ctx context.Context, userID, slotID string, createConferenceLink bool) (*entity.Booking, error) {
	slot, err := s.roomsRepository.GetSlotByID(ctx, slotID)
	if err != nil {
		return nil, fmt.Errorf("get slot: %w", err)
	}
	if slot == nil {
		return nil, domain.ErrorSlotNotFound
	}

	startTime, err := time.Parse(time.RFC3339, slot.Start)
	if err != nil {
		return nil, fmt.Errorf("parse slot start: %w", err)
	}
	if !startTime.After(time.Now().UTC()) {
		return nil, domain.ErrorInvalidRequest
	}
	booking := &entity.Booking{
		ID:        uuid.New().String(),
		SlotID:    slotID,
		UserID:    userID,
		Status:    entity.BookingStatusActive,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	err = s.trManager.Do(ctx, func(txCtx context.Context) error {
		available, err := s.roomsRepository.IsAvailableSlot(txCtx, slotID)
		if err != nil {
			return fmt.Errorf("check slot availability: %w", err)
		}
		if !available {
			return domain.ErrorSlotAlreadyBooked
		}
		if err = s.bookingRepository.CreateBooking(txCtx, booking); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if createConferenceLink {
		link, linkErr := s.conferenceClient.CreateLink(ctx, booking.ID)
		if linkErr != nil {
			s.logger.Error("createBooking: conference link failed",
				zap.String("bookingID", booking.ID), zap.Error(linkErr))
		} else {
			booking.ConferenceLink = link
		}
	}

	return booking, nil
}

func (s *bookingService) ListAllBookings(ctx context.Context, page, pageSize int) ([]*entity.Booking, *entity.Pagination, error) {
	bookings, total, err := s.bookingRepository.ListAllBookings(ctx, page, pageSize)
	if err != nil {
		return nil, nil, fmt.Errorf("list all bookings: %w", err)
	}
	if bookings == nil {
		bookings = []*entity.Booking{}
	}
	return bookings, &entity.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

func (s *bookingService) ListMyBookings(ctx context.Context, userID string) ([]*entity.Booking, error) {
	time := time.Now()
	bookings, err := s.bookingRepository.ListMyBookings(ctx, userID, time) // только будущие слоты
	if err != nil {
		return nil, fmt.Errorf("list my bookings: %w", err)
	}
	if bookings == nil {
		bookings = []*entity.Booking{}
	}
	return bookings, nil
}

func (s *bookingService) CancelBooking(ctx context.Context, bookingID, userID string) (*entity.Booking, error) {
	booking, err := s.bookingRepository.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, fmt.Errorf("get booking: %w", err)
	}
	if booking == nil {
		return nil, domain.ErrorBookingNotFound
	}

	if booking.UserID != userID {
		return nil, domain.ErrorForbidden
	}

	if booking.Status == entity.BookingStatusCancelled {
		return booking, nil
	}

	updated, err := s.bookingRepository.CancelBooking(ctx, bookingID)
	if err != nil {
		return nil, fmt.Errorf("cancel booking: %w", err)
	}
	return updated, nil
}
