package service

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/gorm"
)

type GroupService struct{ db *gorm.DB }

func NewGroupService(db *gorm.DB) *GroupService { return &GroupService{db: db} }

type GroupListInput struct {
	Search   string `form:"search"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
	Sort     string `form:"sort"` // popular, newest
}

type GroupListResult struct {
	Items []model.Group `json:"items"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
	Size  int           `json:"page_size"`
}

func (s *GroupService) List(input GroupListInput) (*GroupListResult, error) {
	if input.PageSize <= 0 { input.PageSize = 20 }
	if input.Page <= 0 { input.Page = 1 }
	q := s.db.Model(&model.Group{}).Where("status = ?", "active")
	if input.Search != "" {
		q = q.Where("name ILIKE ? OR description ILIKE ?", "%"+input.Search+"%", "%"+input.Search+"%")
	}
	var total int64
	q.Count(&total)
	order := "member_count DESC"
	if input.Sort == "newest" { order = "created_at DESC" }
	var items []model.Group
	err := q.Order(order).Offset((input.Page-1)*input.PageSize).Limit(input.PageSize).Find(&items).Error
	return &GroupListResult{Items: items, Total: total, Page: input.Page, Size: input.PageSize}, err
}

func (s *GroupService) Get(id uuid.UUID) (*model.Group, error) {
	var g model.Group
	if err := s.db.First(&g, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("group not found")
	}
	return &g, nil
}

func (s *GroupService) Create(g *model.Group) error { return s.db.Create(g).Error }

func (s *GroupService) Update(g *model.Group) error {
	return s.db.Model(&model.Group{}).Where("id = ?", g.ID).Updates(map[string]interface{}{
		"name": g.Name, "description": g.Description, "cover_url": g.CoverURL,
	}).Error
}

func (s *GroupService) Join(groupID, userID uuid.UUID) error {
	tx := s.db.Begin()
	m := model.GroupMember{GroupID: groupID, UserID: userID, Role: "member"}
	if err := tx.Create(&m).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("already joined or invalid")
	}
	if err := tx.Model(&model.Group{}).Where("id = ?", groupID).
		UpdateColumn("member_count", gorm.Expr("member_count + 1")).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (s *GroupService) Leave(groupID, userID uuid.UUID) error {
	tx := s.db.Begin()
	res := tx.Where("group_id = ? AND user_id = ? AND role != ?", groupID, userID, "owner").Delete(&model.GroupMember{})
	if res.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("cannot leave (owner or not member)")
	}
	if err := tx.Model(&model.Group{}).Where("id = ?", groupID).UpdateColumn("member_count", gorm.Expr("member_count - 1")).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (s *GroupService) GetMembers(groupID uuid.UUID, page, pageSize int) ([]model.GroupMember, int64, error) {
	if pageSize <= 0 { pageSize = 20 }
	var total int64
	s.db.Model(&model.GroupMember{}).Where("group_id = ?", groupID).Count(&total)
	var members []model.GroupMember
	err := s.db.Preload("User").Where("group_id = ?", groupID).
		Order("role ASC, joined_at ASC").Offset((page-1)*pageSize).Limit(pageSize).Find(&members).Error
	return members, total, err
}

func (s *GroupService) GetMyGroups(userID uuid.UUID) ([]model.Group, error) {
	var groups []model.Group
	err := s.db.Joins("JOIN group_members ON groups.id = group_members.group_id").
		Where("group_members.user_id = ?", userID).Order("group_members.joined_at DESC").Find(&groups).Error
	return groups, err
}

func (s *GroupService) GetGroupPosts(groupID uuid.UUID, page, pageSize int) ([]model.Post, int64, error) {
	if pageSize <= 0 { pageSize = 20 }
	var total int64
	s.db.Model(&model.Post{}).Where("group_id = ? AND status = ?", groupID, "approved").Count(&total)
	var posts []model.Post
	err := s.db.Preload("Author").Where("group_id = ? AND status = ?", groupID, "approved").
		Order("created_at DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&posts).Error
	return posts, total, err
}

func (s *GroupService) CreateMember(groupID, userID uuid.UUID, role string) error {
	return s.db.Create(&model.GroupMember{GroupID: groupID, UserID: userID, Role: role}).Error
}
