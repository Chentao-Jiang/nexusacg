package service

import (
	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/gorm"
)

type MessageService struct{ db *gorm.DB }

func NewMessageService(db *gorm.DB) *MessageService { return &MessageService{db: db} }

func (s *MessageService) Send(msg *model.Message) error { return s.db.Create(msg).Error }

type ConversationInfo struct {
	OtherUserID   uuid.UUID `json:"other_user_id"`
	OtherNickname string    `json:"other_nickname"`
	OtherAvatar   string    `json:"other_avatar_url"`
	LastMessage   string    `json:"last_message"`
	LastTime      string    `json:"last_time"`
	Unread        int64     `json:"unread"`
}

func (s *MessageService) GetConversations(userID uuid.UUID) ([]ConversationInfo, error) {
	var infos []ConversationInfo
	s.db.Raw(`
		SELECT 
			CASE WHEN m.sender_id = ? THEN m.receiver_id ELSE m.sender_id END as other_user_id,
			u.nickname as other_nickname,
			u.avatar_url as other_avatar_url,
			last_msg.content as last_message,
			last_msg.created_at as last_time,
			COALESCE(unread.cnt, 0) as unread
		FROM messages m
		JOIN users u ON u.id = CASE WHEN m.sender_id = ? THEN m.receiver_id ELSE m.sender_id END
		JOIN LATERAL (
			SELECT content, created_at FROM messages m2
			WHERE (m2.sender_id = m.sender_id AND m2.receiver_id = m.receiver_id)
			   OR (m2.sender_id = m.receiver_id AND m2.receiver_id = m.sender_id)
			ORDER BY created_at DESC LIMIT 1
		) last_msg ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*) as cnt FROM messages m3
			WHERE m3.receiver_id = ? AND m3.sender_id = CASE WHEN m.sender_id = ? THEN m.receiver_id ELSE m.sender_id END AND m3.is_read = false
		) unread ON true
		WHERE m.sender_id = ? OR m.receiver_id = ?
		GROUP BY other_user_id, u.nickname, u.avatar_url, last_msg.content, last_msg.created_at, unread.cnt
		ORDER BY last_msg.created_at DESC
	`, userID, userID, userID, userID, userID, userID).Scan(&infos)
	return infos, nil
}

func (s *MessageService) GetMessages(userID, otherID uuid.UUID, page, pageSize int) ([]model.Message, error) {
	if pageSize <= 0 { pageSize = 50 }
	if page <= 0 { page = 1 }
	var msgs []model.Message
	err := s.db.Where(
		"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		userID, otherID, otherID, userID,
	).Order("created_at DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&msgs).Error
	// Mark as read
	s.db.Model(&model.Message{}).Where("sender_id = ? AND receiver_id = ? AND is_read = false", otherID, userID).Update("is_read", true)
	return msgs, err
}
