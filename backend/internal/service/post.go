package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/gorm"
)

type PostService struct {
	db *gorm.DB
}

func NewPostService(db *gorm.DB) *PostService {
	return &PostService{db: db}
}

type CreatePostInput struct {
	UserID   uuid.UUID `json:"user_id"`
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	Images   []string  `json:"images"`
	VideoURL *string   `json:"video_url"`
	Type     string    `json:"type"`
	Tags     []string  `json:"tags"`
}

func (s *PostService) Create(ctx context.Context, input CreatePostInput) (*model.Post, error) {
	if input.Images == nil {
		input.Images = []string{}
	}
	if input.Tags == nil {
		input.Tags = []string{}
	}
	post := model.Post{
		ID:      uuid.New(),
		UserID:  input.UserID,
		Title:   input.Title,
		Content: input.Content,
		Images:  input.Images,
		Status:  "pending_review",
	}
	if input.VideoURL != nil {
		post.VideoURL = input.VideoURL
		post.Type = "video"
	} else if len(input.Images) > 0 {
		post.Type = "image"
	} else {
		post.Type = "text"
	}
	if input.Tags != nil {
		post.Tags = input.Tags
	}

	if err := s.db.Create(&post).Error; err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}
	return &post, nil
}

type PostListInput struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
	Tag      string `form:"tag"`
	UserID   string `form:"user_id"`
	Keyword  string `form:"keyword"`
}

type PostListResult struct {
	Items []model.Post `json:"items"`
	Total int64        `json:"total"`
	Page  int          `json:"page"`
	Size  int          `json:"page_size"`
}

func (s *PostService) List(ctx context.Context, input PostListInput) (*PostListResult, error) {
	query := s.db.Model(&model.Post{}).Where("status = ?", "approved")

	if input.Tag != "" {
		tagJSON, _ := json.Marshal([]string{input.Tag})
		query = query.Where("tags @> ?::jsonb", string(tagJSON))
	}
	if input.UserID != "" {
		query = query.Where("user_id = ?", input.UserID)
	}
	if input.Keyword != "" {
		words := strings.Fields(input.Keyword)
		for _, word := range words {
			pattern := "%" + word + "%"
			query = query.Where("title ILIKE ? OR content ILIKE ?", pattern, pattern)
		}
	}

	var total int64
	query.Count(&total)

	offset := (input.Page - 1) * input.PageSize
	if offset < 0 {
		offset = 0
	}

	var posts []model.Post
	if err := query.Preload("Author").Offset(offset).Limit(input.PageSize).Order("created_at DESC").Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}

	return &PostListResult{Items: posts, Total: total, Page: input.Page, Size: input.PageSize}, nil
}

func (s *PostService) Get(ctx context.Context, id uuid.UUID) (*model.Post, error) {
	var post model.Post
	if err := s.db.Preload("Author").Where("id = ? AND status = ?", id, "approved").First(&post).Error; err != nil {
		return nil, fmt.Errorf("post not found")
	}
	return &post, nil
}

func (s *PostService) Like(ctx context.Context, userID, postID uuid.UUID) error {
	like := model.Like{UserID: userID, PostID: postID}
	if err := s.db.Create(&like).Error; err != nil {
		return fmt.Errorf("already liked or invalid")
	}
	s.db.Model(&model.Post{}).Where("id = ?", postID).Update("like_count", gorm.Expr("like_count + 1"))
	return nil
}

func (s *PostService) Unlike(ctx context.Context, userID, postID uuid.UUID) error {
	result := s.db.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&model.Like{})
	if result.RowsAffected == 0 {
		return fmt.Errorf("like not found")
	}
	s.db.Model(&model.Post{}).Where("id = ?", postID).Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)"))
	return nil
}

type CommentInput struct {
	PostID   uuid.UUID `json:"post_id" binding:"required"`
	UserID   uuid.UUID `json:"user_id"`
	Content  string    `json:"content" binding:"required,max=2000"`
	ParentID *uuid.UUID `json:"parent_id"`
}

func (s *PostService) CreateComment(ctx context.Context, input CommentInput) (*model.Comment, error) {
	// Validate parent comment exists and belongs to the same post
	if input.ParentID != nil {
		var parent model.Comment
		if err := s.db.Where("id = ? AND post_id = ?", input.ParentID, input.PostID).First(&parent).Error; err != nil {
			return nil, fmt.Errorf("parent comment not found on this post")
		}
	}

	comment := model.Comment{
		ID:       uuid.New(),
		PostID:   input.PostID,
		UserID:   input.UserID,
		Content:  input.Content,
		ParentID: input.ParentID,
		Status:   "pending_review",
	}
	if err := s.db.Create(&comment).Error; err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}
	s.db.Model(&model.Post{}).Where("id = ?", input.PostID).Update("comment_count", gorm.Expr("comment_count + 1"))
	return &comment, nil
}

type CommentListResult struct {
	Items []model.Comment `json:"items"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"page_size"`
}

func (s *PostService) ListComments(ctx context.Context, postID uuid.UUID, page, pageSize int) (*CommentListResult, error) {
	query := s.db.Model(&model.Comment{}).Where("post_id = ? AND status = ?", postID, "approved")
	var total int64
	query.Count(&total)
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	var comments []model.Comment
	if err := query.Preload("Author").Offset(offset).Limit(pageSize).Order("created_at ASC").Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}
	return &CommentListResult{Items: comments, Total: total, Page: page, Size: pageSize}, nil
}
