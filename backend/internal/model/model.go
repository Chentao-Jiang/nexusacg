package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Phone        *string   `json:"phone,omitempty" gorm:"uniqueIndex"`
	Email        *string   `json:"email,omitempty" gorm:"uniqueIndex"`
	WechatOpenID *string   `json:"-" gorm:"uniqueIndex"`
	QQOpenID     *string   `json:"-" gorm:"uniqueIndex"`
	PasswordHash *string   `json:"-"`
	Nickname     string    `json:"nickname" gorm:"default:''"`
	AvatarURL    *string   `json:"avatar_url,omitempty"`
	Bio          string    `json:"bio" gorm:"default:''"`
	Role         string    `json:"role" gorm:"default:user"`
	Status       string    `json:"status" gorm:"default:active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (User) TableName() string { return "users" }

type Category struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string     `json:"name"`
	Zone      string     `json:"zone" gorm:"index"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	IconURL   *string    `json:"icon_url,omitempty"`
	SortOrder int        `json:"sort_order" gorm:"default:0"`
	CreatedAt time.Time  `json:"created_at"`
}

func (Category) TableName() string { return "categories" }

type Product struct {
	ID            uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SellerID      uuid.UUID    `json:"seller_id" gorm:"type:uuid;index"`
	CategoryID    *uuid.UUID   `json:"category_id,omitempty" gorm:"type:uuid"`
	Name          string       `json:"name"`
	Description   string       `json:"description"`
	Price         float64      `json:"price"`
	OriginalPrice *float64     `json:"original_price,omitempty"`
	Currency      string       `json:"currency" gorm:"default:CNY"`
	Zone          string       `json:"zone" gorm:"index"`
	SourceType    string       `json:"source_type" gorm:"default:self_made"`
	SellerType    string       `json:"seller_type" gorm:"default:uncertified;index"`
	Images        StringArray  `json:"images" gorm:"type:jsonb;default:'[]'"`
	Stock         int          `json:"stock" gorm:"default:0"`
	Status        string       `json:"status" gorm:"default:active;index"`
	Tags          StringArray  `json:"tags" gorm:"type:jsonb;default:'[]'"`
	CharacterName *string      `json:"character_name,omitempty"`
	AnimeName     *string      `json:"anime_name,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

func (Product) TableName() string { return "products" }

type Order struct {
	ID              uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          uuid.UUID   `json:"user_id" gorm:"type:uuid;index"`
	OrderNo         string      `json:"order_no" gorm:"uniqueIndex"`
	TotalAmount     float64     `json:"total_amount"`
	PaymentMethod   *string     `json:"payment_method,omitempty"`
	PaymentStatus   string      `json:"payment_status" gorm:"default:pending;index"`
	OrderStatus     string      `json:"order_status" gorm:"default:pending;index"`
	ShippingAddress *string     `json:"shipping_address,omitempty" gorm:"type:jsonb"`
	PaymentID       *string     `json:"payment_id,omitempty"`
	IdempotencyKey  *string     `json:"-" gorm:"uniqueIndex"`
	PaidAt          *time.Time  `json:"paid_at,omitempty"`
	ShippedAt       *time.Time  `json:"shipped_at,omitempty"`
	CompletedAt     *time.Time  `json:"completed_at,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	Items           []OrderItem `json:"items,omitempty" gorm:"foreignKey:OrderID"`
	DisputeStatus   string      `json:"dispute_status" gorm:"default:none;index"` // none | disputed | resolved
}

func (Order) TableName() string { return "orders" }

type OrderItem struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrderID   uuid.UUID `json:"order_id" gorm:"type:uuid"`
	ProductID uuid.UUID `json:"product_id" gorm:"type:uuid"`
	Quantity  int       `json:"quantity" gorm:"default:1"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}

func (OrderItem) TableName() string { return "order_items" }

type Post struct {
	ID           uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       uuid.UUID   `json:"user_id" gorm:"type:uuid;index"`
	Title        string      `json:"title"`
	Content      string      `json:"content"`
	Images       StringArray `json:"images" gorm:"type:jsonb;default:'[]'"`
	VideoURL     *string     `json:"video_url,omitempty"`
	Type         string      `json:"type" gorm:"default:text"`
	Tags         StringArray `json:"tags" gorm:"type:jsonb;default:'[]'"`
	LikeCount    int         `json:"like_count" gorm:"default:0"`
	CommentCount int         `json:"comment_count" gorm:"default:0"`
	Status       string      `json:"status" gorm:"default:pending_review;index"`
	Visibility   string      `json:"visibility" gorm:"default:public;index"` // public | followers | private
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Author       *User       `json:"author,omitempty" gorm:"foreignKey:UserID;references:ID"`
}

func (Post) TableName() string { return "posts" }

type Comment struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PostID    uuid.UUID `json:"post_id" gorm:"type:uuid;index"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	Content   string    `json:"content"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty" gorm:"type:uuid"`
	Status    string    `json:"status" gorm:"default:pending_review"`
	CreatedAt time.Time `json:"created_at"`
	Author    *User     `json:"author,omitempty" gorm:"foreignKey:UserID;references:ID"`
}

func (Comment) TableName() string { return "comments" }

type Like struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid"`
	PostID    uuid.UUID `json:"post_id" gorm:"type:uuid;index"`
	CreatedAt time.Time `json:"created_at"`
}

func (Like) TableName() string { return "likes" }

type Follow struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FollowerID  uuid.UUID `json:"follower_id" gorm:"type:uuid;uniqueIndex:idx_follow;not null"`
	FollowingID uuid.UUID `json:"following_id" gorm:"type:uuid;uniqueIndex:idx_follow;not null"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Follow) TableName() string { return "follows" }

type Group struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CoverURL    *string   `json:"cover_url,omitempty"`
	OwnerID     uuid.UUID `json:"owner_id" gorm:"type:uuid"`
	MemberCount int       `json:"member_count" gorm:"default:1"`
	Status      string    `json:"status" gorm:"default:active"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Group) TableName() string { return "groups" }

type Event struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string    `json:"name"`
	Description string  `json:"description"`
	CoverURL  *string   `json:"cover_url,omitempty"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Address   string    `json:"address"`
	Latitude  *float64  `json:"latitude,omitempty"`
	Longitude *float64  `json:"longitude,omitempty"`
	Source    string    `json:"source" gorm:"default:manual"`
	RegisteredUserIDs StringArray `json:"registered_user_ids,omitempty" gorm:"type:jsonb;default:'[]'"`
	Status    string    `json:"status" gorm:"default:upcoming;index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Event) TableName() string { return "events" }

type ServiceProvider struct {
	ID              uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          uuid.UUID   `json:"user_id" gorm:"type:uuid"`
	Type            string      `json:"type" gorm:"index"`
	Description     string      `json:"description"`
	PortfolioImages StringArray `json:"portfolio_images" gorm:"type:jsonb;default:'[]'"`
	PriceList       StringArray `json:"price_list" gorm:"type:jsonb;default:'[]'"`
	Rating          float64     `json:"rating" gorm:"default:0"`
	ReviewCount     int         `json:"review_count" gorm:"default:0"`
	IsVerified      bool        `json:"is_verified" gorm:"default:false"`
	Status          string      `json:"status" gorm:"default:active"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

func (ServiceProvider) TableName() string { return "service_providers" }

type RefreshToken struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	Token     string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

// PaymentLog records every payment provider callback for audit and idempotency.
type PaymentLog struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrderID        uuid.UUID `json:"order_id" gorm:"type:uuid;index"`
	PaymentMethod  string    `json:"payment_method" gorm:"index"`            // wechat, alipay
	TransactionID  string    `json:"transaction_id" gorm:"uniqueIndex"`     // provider's transaction ID
	Amount         float64   `json:"amount"`
	Status         string    `json:"status" gorm:"index"`                    // success, failed, refunded
	RawPayload     string    `json:"-" gorm:"type:text"`                    // raw notification for debugging
	SignatureValid bool      `json:"signature_valid"`
	Processed      bool      `json:"processed" gorm:"default:false"`        // whether business logic was applied
	CreatedAt      time.Time `json:"created_at"`
}

func (PaymentLog) TableName() string { return "payment_logs" }

type ProfitShareRecord struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrderID       uuid.UUID  `json:"order_id" gorm:"type:uuid;index"`
	TotalAmount   float64    `json:"total_amount"`
	PlatformFee   float64    `json:"platform_fee"`
	SellerAmount  float64    `json:"seller_amount"`
	Status        string     `json:"status" gorm:"default:pending"`
	PaymentMethod string     `json:"payment_method"`
	ReleasedAt    *time.Time `json:"released_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (ProfitShareRecord) TableName() string { return "profit_share_records" }

// CertificationApplication stores merchant/service_provider certification requests.
type CertificationApplication struct {
	ID                 uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID             uuid.UUID   `json:"user_id" gorm:"type:uuid;index"`
	Type               string      `json:"type" gorm:"index"` // merchant | service_provider
	BusinessLicenseURL *string     `json:"business_license_url,omitempty"`
	ProductCategory    *string     `json:"product_category,omitempty"`
	StoreName          *string     `json:"store_name,omitempty"`
	ProviderType       *string     `json:"provider_type,omitempty"` // makeup_artist | wig_stylist | photographer | post_editor | props_maker
	PortfolioImages    StringArray `json:"portfolio_images" gorm:"type:jsonb;default:'[]'"`
	Status             string      `json:"status" gorm:"default:pending;index"` // pending | approved | rejected
	ReviewedBy         *uuid.UUID  `json:"reviewed_by,omitempty" gorm:"type:uuid"`
	ReviewedAt         *time.Time  `json:"reviewed_at,omitempty"`
	RejectionReason    *string     `json:"rejection_reason,omitempty"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
}

func (CertificationApplication) TableName() string { return "certification_applications" }

// ServiceSchedule stores makeup_artist/photographer availability slots.
type ServiceSchedule struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ServiceProviderID uuid.UUID  `json:"service_provider_id" gorm:"type:uuid;index"`
	EventID           *uuid.UUID `json:"event_id,omitempty" gorm:"type:uuid"`
	Date              time.Time  `json:"date"`
	StartTime         *time.Time `json:"start_time,omitempty"`
	EndTime           *time.Time `json:"end_time,omitempty"`
	Status            string     `json:"status" gorm:"default:available"` // available | booked | unavailable
	Notes             string     `json:"notes" gorm:"default:''"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (ServiceSchedule) TableName() string { return "service_schedules" }

// EventServiceListing links a service provider to an event with pricing.
type EventServiceListing struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EventID           uuid.UUID `json:"event_id" gorm:"type:uuid;index"`
	ServiceProviderID uuid.UUID `json:"service_provider_id" gorm:"type:uuid;index"`
	Price             float64   `json:"price"`
	Description       string    `json:"description"`
	Status            string    `json:"status" gorm:"default:active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (EventServiceListing) TableName() string { return "event_service_listings" }

// EmailVerificationToken stores email verification tokens (24h TTL).
type EmailVerificationToken struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;uniqueIndex"`
	Token     string    `json:"-" gorm:"uniqueIndex"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (EmailVerificationToken) TableName() string { return "email_verification_tokens" }

// ServiceProduct represents a service offering in the service zone.
// Unlike physical products, availability is schedule-based and display includes
// portfolio images of completed work.
type ServiceProduct struct {
	ID                uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ServiceProviderID uuid.UUID   `json:"service_provider_id" gorm:"type:uuid;index"`
	UserID            uuid.UUID   `json:"user_id" gorm:"type:uuid;index"`
	CategoryID        *uuid.UUID  `json:"category_id,omitempty" gorm:"type:uuid"`
	Name              string      `json:"name"`
	Description       string      `json:"description"`
	Price             float64     `json:"price"`
	OriginalPrice     *float64    `json:"original_price,omitempty"`
	Currency          string      `json:"currency" gorm:"default:CNY"`
	ServiceType       string      `json:"service_type" gorm:"index"` // makeup_artist | wig_stylist | photographer | post_editor | props_maker
	Images            StringArray `json:"images" gorm:"type:jsonb;default:'[]'"`
	PortfolioImages   StringArray `json:"portfolio_images" gorm:"type:jsonb;default:'[]'"`
	Status            string      `json:"status" gorm:"default:active;index"`
	Tags              StringArray `json:"tags" gorm:"type:jsonb;default:'[]'"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

func (ServiceProduct) TableName() string { return "service_products" }

// PromotionApplication tracks traffic promotion (投流) requests.
type PromotionApplication struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          uuid.UUID  `json:"user_id" gorm:"type:uuid;index"`
	TargetType      string     `json:"target_type" gorm:"index"` // product | service_product
	TargetID        uuid.UUID  `json:"target_id" gorm:"type:uuid;index"`
	Budget          float64    `json:"budget"`
	Duration        int        `json:"duration"` // days
	Reason          string     `json:"reason"`
	Status          string     `json:"status" gorm:"default:pending;index"` // pending | approved | rejected | active | expired
	ReviewedBy      *uuid.UUID `json:"reviewed_by,omitempty" gorm:"type:uuid"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
	ActivatedAt     *time.Time `json:"activated_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	LinkedPostID    *uuid.UUID `json:"linked_post_id,omitempty" gorm:"type:uuid"` // Phase 2
	Impressions     int        `json:"impressions" gorm:"default:0"`
	Clicks          int        `json:"clicks" gorm:"default:0"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (PromotionApplication) TableName() string { return "promotion_applications" }

// Dispute tracks buyer-seller disputes over orders.
type Dispute struct {
	ID             uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrderID        uuid.UUID   `json:"order_id" gorm:"type:uuid;index"`
	OrderNo        string      `json:"order_no" gorm:"index"`
	InitiatorID    uuid.UUID   `json:"initiator_id" gorm:"type:uuid"`
	SellerID       uuid.UUID   `json:"seller_id" gorm:"type:uuid;index"`
	Reason         string      `json:"reason"`
	Description    string      `json:"description"`
	EvidenceImages StringArray `json:"evidence_images" gorm:"type:jsonb;default:'[]'"`
	RefundAmount   float64     `json:"refund_amount"` // 0 = full refund
	Status         string      `json:"status" gorm:"default:pending;index"` // pending | seller_responded | admin_review | resolved | buyer_won | seller_won
	AdminDecision  string      `json:"admin_decision,omitempty"` // full_refund | partial_refund | reject
	AdminNote      *string     `json:"admin_note,omitempty"`
	RespondedBy    *uuid.UUID  `json:"responded_by,omitempty" gorm:"type:uuid"`
	RespondedAt    *time.Time  `json:"responded_at,omitempty"`
	ResolvedBy     *uuid.UUID  `json:"resolved_by,omitempty" gorm:"type:uuid"`
	ResolvedAt     *time.Time  `json:"resolved_at,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

func (Dispute) TableName() string { return "disputes" }

// DisputeMessage stores communication within a dispute.
type DisputeMessage struct {
	ID          uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	DisputeID   uuid.UUID   `json:"dispute_id" gorm:"type:uuid;index"`
	SenderID    uuid.UUID   `json:"sender_id" gorm:"type:uuid"`
	SenderRole  string      `json:"sender_role"` // buyer | seller | admin
	Content     string      `json:"content"`
	Images      StringArray `json:"images" gorm:"type:jsonb;default:'[]'"`
	CreatedAt   time.Time   `json:"created_at"`
}

func (DisputeMessage) TableName() string { return "dispute_messages" }

// RefundApplication tracks buyer-initiated refund/return requests for orders.
type RefundApplication struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrderID      uuid.UUID  `json:"order_id" gorm:"type:uuid;index"`
	OrderNo      string     `json:"order_no" gorm:"index"`
	UserID       uuid.UUID  `json:"user_id" gorm:"type:uuid;index"`
	SellerID     uuid.UUID  `json:"seller_id" gorm:"type:uuid;index"`
	RefundType   string     `json:"refund_type" gorm:"index"` // refund_only | return_refund
	Reason       string     `json:"reason"`
	EvidenceURLs StringArray `json:"evidence_urls" gorm:"type:jsonb;default:'[]'"`
	Amount       float64    `json:"amount"` // requested refund amount; 0 = full refund
	Status       string     `json:"status" gorm:"default:pending;index"` // pending | seller_review | approved | rejected | completed
	SellerNote   *string    `json:"seller_note,omitempty"`
	AdminNote    *string    `json:"admin_note,omitempty"`
	ReviewedBy   *uuid.UUID `json:"reviewed_by,omitempty" gorm:"type:uuid"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (RefundApplication) TableName() string { return "refund_applications" }
