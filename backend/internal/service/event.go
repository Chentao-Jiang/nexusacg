package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/gorm"
)

type EventService struct {
	db *gorm.DB
}

func NewEventService(db *gorm.DB) *EventService {
	return &EventService{db: db}
}

type EventListInput struct {
	Status   string `form:"status"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
}

type EventListResult struct {
	Items []model.Event `json:"items"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
	Size  int           `json:"page_size"`
}

func (s *EventService) List(ctx context.Context, input EventListInput) (*EventListResult, error) {
	query := s.db.Model(&model.Event{})

	if input.Status != "" {
		query = query.Where("status = ?", input.Status)
	}

	var total int64
	query.Count(&total)

	offset := (input.Page - 1) * input.PageSize
	if offset < 0 {
		offset = 0
	}

	var events []model.Event
	if err := query.Offset(offset).Limit(input.PageSize).Order("start_time ASC").Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	return &EventListResult{Items: events, Total: total, Page: input.Page, Size: input.PageSize}, nil
}

func (s *EventService) Get(ctx context.Context, id uuid.UUID) (*model.Event, error) {
	var event model.Event
	if err := s.db.Where("id = ?", id).First(&event).Error; err != nil {
		return nil, fmt.Errorf("event not found")
	}
	return &event, nil
}

type CreateEventInput struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CoverURL    *string  `json:"cover_url"`
	StartTime   string   `json:"start_time"`
	EndTime     string   `json:"end_time"`
	Address     string   `json:"address"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
}

func (s *EventService) Create(ctx context.Context, input CreateEventInput) (*model.Event, error) {
	startTime, err := time.Parse(time.RFC3339, input.StartTime)
	if err != nil {
		startTime, err = time.Parse("2006-01-02 15:04:05", input.StartTime)
	}
	if err != nil {
		return nil, fmt.Errorf("invalid start_time format: %s", input.StartTime)
	}

	endTime, err := time.Parse(time.RFC3339, input.EndTime)
	if err != nil {
		endTime, err = time.Parse("2006-01-02 15:04:05", input.EndTime)
	}
	if err != nil {
		return nil, fmt.Errorf("invalid end_time format: %s", input.EndTime)
	}

	event := model.Event{
		ID:          uuid.New(),
		Name:        input.Name,
		Description: input.Description,
		CoverURL:    input.CoverURL,
		StartTime:   startTime,
		EndTime:     endTime,
		Address:     input.Address,
		Latitude:    input.Latitude,
		Longitude:   input.Longitude,
		Source:      "manual",
		Status:      "upcoming",
	}

	if err := s.db.Create(&event).Error; err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}
	return &event, nil
}


func (s *EventService) Register(eventID, userID uuid.UUID) error {
	var event model.Event
	if err := s.db.Where("id = ?", eventID).First(&event).Error; err != nil {
		return fmt.Errorf("event not found")
	}
	uid := userID.String()
	for _, id := range event.RegisteredUserIDs {
		if id == uid {
			return fmt.Errorf("already registered")
		}
	}
	event.RegisteredUserIDs = append(event.RegisteredUserIDs, uid)
	return s.db.Model(&event).Update("registered_user_ids", event.RegisteredUserIDs).Error
}

func (s *EventService) GetMyRegistrations(userID uuid.UUID) ([]model.Event, error) {
	var events []model.Event
	uid := userID.String()
	// PostgreSQL JSONB contains operator
	err := s.db.Where("status != ? AND registered_user_ids @> ?", "cancelled", fmt.Sprintf(`["%s"]`, uid)).
		Order("start_time DESC").Find(&events).Error
	return events, err
}
