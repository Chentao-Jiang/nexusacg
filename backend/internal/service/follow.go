package service

import (
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FollowService struct {
	db *gorm.DB
}

func NewFollowService(db *gorm.DB) *FollowService {
	return &FollowService{db: db}
}

func (s *FollowService) Follow(followerID, followingID uuid.UUID) error {
	if followerID == followingID {
		return gorm.ErrInvalidData
	}
	f := model.Follow{FollowerID: followerID, FollowingID: followingID}
	return s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&f).Error
}

func (s *FollowService) Unfollow(followerID, followingID uuid.UUID) error {
	return s.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&model.Follow{}).Error
}

func (s *FollowService) IsFollowing(followerID, followingID uuid.UUID) (bool, error) {
	var count int64
	err := s.db.Model(&model.Follow{}).Where("follower_id = ? AND following_id = ?", followerID, followingID).Count(&count).Error
	return count > 0, err
}

type FollowListResult struct {
	Items    []FollowUserInfo `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

type FollowUserInfo struct {
	UserID   uuid.UUID `json:"user_id"`
	Nickname string    `json:"nickname"`
	Avatar   string    `json:"avatar_url"`
}

func (s *FollowService) GetFollowers(userID uuid.UUID, page, pageSize int) (*FollowListResult, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	var total int64
	s.db.Model(&model.Follow{}).Where("following_id = ?", userID).Count(&total)
	var infos []FollowUserInfo
	err := s.db.Model(&model.Follow{}).
		Select("users.id as user_id, users.nickname, users.avatar_url").
		Joins("JOIN users ON users.id = follows.follower_id").
		Where("follows.following_id = ?", userID).
		Order("follows.created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&infos).Error
	return &FollowListResult{Items: infos, Total: total, Page: page, PageSize: pageSize}, err
}

func (s *FollowService) GetFollowing(userID uuid.UUID, page, pageSize int) (*FollowListResult, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	var total int64
	s.db.Model(&model.Follow{}).Where("follower_id = ?", userID).Count(&total)
	var infos []FollowUserInfo
	err := s.db.Model(&model.Follow{}).
		Select("users.id as user_id, users.nickname, users.avatar_url").
		Joins("JOIN users ON users.id = follows.following_id").
		Where("follows.follower_id = ?", userID).
		Order("follows.created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&infos).Error
	return &FollowListResult{Items: infos, Total: total, Page: page, PageSize: pageSize}, err
}
