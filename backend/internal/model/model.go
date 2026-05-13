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
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	Items           []OrderItem `json:"items,omitempty" gorm:"foreignKey:OrderID"`
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
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Author       *User       `json:"author,omitempty" gorm:"foreignKey:ID;references:UserID"`
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
	Author    *User     `json:"author,omitempty" gorm:"foreignKey:ID;references:UserID"`
}

func (Comment) TableName() string { return "comments" }

type Like struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid"`
	PostID    uuid.UUID `json:"post_id" gorm:"type:uuid;index"`
	CreatedAt time.Time `json:"created_at"`
}

func (Like) TableName() string { return "likes" }

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
