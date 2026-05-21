package service

import (
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BookmarkService struct{ db *gorm.DB }

func NewBookmarkService(db *gorm.DB) *BookmarkService { return &BookmarkService{db: db} }

func (s *BookmarkService) Bookmark(userID, postID uuid.UUID) error {
	b := model.Bookmark{UserID: userID, PostID: postID}
	return s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&b).Error
}

func (s *BookmarkService) Unbookmark(userID, postID uuid.UUID) error {
	return s.db.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&model.Bookmark{}).Error
}

func (s *BookmarkService) IsBookmarked(userID, postID uuid.UUID) (bool, error) {
	var count int64
	err := s.db.Model(&model.Bookmark{}).Where("user_id = ? AND post_id = ?", userID, postID).Count(&count).Error
	return count > 0, err
}

type BookmarkListResult struct {
	Items    []BookmarkPostInfo `json:"items"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

type BookmarkPostInfo struct {
	PostID    uuid.UUID `json:"post_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Images    string    `json:"images"`
	VideoURL  *string   `json:"video_url"`
	LikeCount int       `json:"like_count"`
}

func (s *BookmarkService) GetMyBookmarks(userID uuid.UUID, page, pageSize int) (*BookmarkListResult, error) {
	if pageSize <= 0 { pageSize = 20 }
	if page <= 0 { page = 1 }
	var total int64
	s.db.Model(&model.Bookmark{}).Where("user_id = ?", userID).Count(&total)
	var infos []BookmarkPostInfo
	err := s.db.Model(&model.Bookmark{}).
		Select("posts.id as post_id, posts.title, posts.content, posts.images, posts.video_url, posts.like_count").
		Joins("JOIN posts ON posts.id = bookmarks.post_id").
		Where("bookmarks.user_id = ?", userID).
		Order("bookmarks.created_at DESC").
		Offset((page-1)*pageSize).Limit(pageSize).
		Scan(&infos).Error
	return &BookmarkListResult{Items: infos, Total: total, Page: page, PageSize: pageSize}, err
}
