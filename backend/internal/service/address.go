package service

import (
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/gorm"
)

type AddressService struct{ db *gorm.DB }

func NewAddressService(db *gorm.DB) *AddressService { return &AddressService{db: db} }

func (s *AddressService) List(userID uuid.UUID) ([]model.Address, error) {
	var addrs []model.Address
	err := s.db.Where("user_id = ?", userID).Order("is_default DESC, created_at DESC").Find(&addrs).Error
	return addrs, err
}

func (s *AddressService) Create(addr *model.Address) error {
	if addr.IsDefault {
		s.db.Model(&model.Address{}).Where("user_id = ?", addr.UserID).Update("is_default", false)
	}
	return s.db.Create(addr).Error
}

func (s *AddressService) Update(addr *model.Address) error {
	if addr.IsDefault {
		s.db.Model(&model.Address{}).Where("user_id = ?", addr.UserID).Update("is_default", false)
	}
	return s.db.Updates(addr).Error
}

func (s *AddressService) Delete(id, userID uuid.UUID) error {
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Address{}).Error
}
