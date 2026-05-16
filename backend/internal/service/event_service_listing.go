package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"

	"gorm.io/gorm"
)

type EventServiceListingService struct {
	db *gorm.DB
}

func NewEventServiceListingService(db *gorm.DB) *EventServiceListingService {
	return &EventServiceListingService{db: db}
}

type CreateListingInput struct {
	EventID           uuid.UUID `json:"event_id"`
	ServiceProviderID uuid.UUID `json:"service_provider_id"`
	Price             float64   `json:"price"`
	Description       string    `json:"description"`
}

type CreateScheduleInput struct {
	ServiceProviderID uuid.UUID  `json:"service_provider_id"`
	EventID           *uuid.UUID `json:"event_id,omitempty"`
	Date              string     `json:"date"`
	StartTime         *string    `json:"start_time,omitempty"`
	EndTime           *string    `json:"end_time,omitempty"`
	Status            string     `json:"status"`
	Notes             string     `json:"notes"`
}

type ListingWithProvider struct {
	model.EventServiceListing
	ProviderType string `json:"provider_type"`
	IsVerified   bool   `json:"is_verified"`
}

func (s *EventServiceListingService) CreateListing(ctx context.Context, input CreateListingInput) (*model.EventServiceListing, error) {
	var sp model.ServiceProvider
	if err := s.db.Where("id = ? AND is_verified = ?", input.ServiceProviderID, true).First(&sp).Error; err != nil {
		return nil, fmt.Errorf("service provider not found or not verified")
	}

	listing := model.EventServiceListing{
		ID:                uuid.New(),
		EventID:           input.EventID,
		ServiceProviderID: input.ServiceProviderID,
		Price:             input.Price,
		Description:       input.Description,
		Status:            "active",
	}
	if err := s.db.Create(&listing).Error; err != nil {
		return nil, fmt.Errorf("failed to create listing: %w", err)
	}
	return &listing, nil
}

func (s *EventServiceListingService) ListByEvent(ctx context.Context, eventID uuid.UUID) ([]ListingWithProvider, error) {
	var listings []model.EventServiceListing
	if err := s.db.Where("event_id = ? AND status = ?", eventID, "active").Find(&listings).Error; err != nil {
		return nil, fmt.Errorf("failed to list event services: %w", err)
	}

	var result []ListingWithProvider
	for _, l := range listings {
		var sp model.ServiceProvider
		if err := s.db.Where("id = ?", l.ServiceProviderID).First(&sp).Error; err != nil {
			continue
		}
		result = append(result, ListingWithProvider{
			EventServiceListing: l,
			ProviderType:        sp.Type,
			IsVerified:          sp.IsVerified,
		})
	}
	return result, nil
}

func (s *EventServiceListingService) CreateSchedule(ctx context.Context, input CreateScheduleInput) (*model.ServiceSchedule, error) {
	var sp model.ServiceProvider
	if err := s.db.Where("id = ?", input.ServiceProviderID).First(&sp).Error; err != nil {
		return nil, fmt.Errorf("service provider not found")
	}

	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format, use YYYY-MM-DD")
	}

	status := "available"
	if input.Status != "" {
		status = input.Status
	}

	schedule := model.ServiceSchedule{
		ID:                uuid.New(),
		ServiceProviderID: input.ServiceProviderID,
		EventID:           input.EventID,
		Date:              date,
		Status:            status,
		Notes:             input.Notes,
	}

	if input.StartTime != nil {
		t, err := time.Parse("15:04", *input.StartTime)
		if err != nil {
			return nil, fmt.Errorf("invalid start_time format, use HH:MM")
		}
		schedule.StartTime = &t
	}
	if input.EndTime != nil {
		t, err := time.Parse("15:04", *input.EndTime)
		if err != nil {
			return nil, fmt.Errorf("invalid end_time format, use HH:MM")
		}
		schedule.EndTime = &t
	}

	if err := s.db.Create(&schedule).Error; err != nil {
		return nil, fmt.Errorf("failed to create schedule: %w", err)
	}
	return &schedule, nil
}

func (s *EventServiceListingService) ListSchedules(ctx context.Context, providerID uuid.UUID, eventID *uuid.UUID) ([]model.ServiceSchedule, error) {
	var schedules []model.ServiceSchedule
	query := s.db.Where("service_provider_id = ?", providerID)
	if eventID != nil {
		query = query.Where("event_id = ?", eventID)
	}
	if err := query.Order("date ASC").Find(&schedules).Error; err != nil {
		return nil, fmt.Errorf("failed to list schedules: %w", err)
	}
	return schedules, nil
}
