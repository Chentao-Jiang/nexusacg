-- NexusACG Database Schema
-- PostgreSQL 15+

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(20) UNIQUE,
    email VARCHAR(255) UNIQUE,
    wechat_open_id VARCHAR(255) UNIQUE,
    qq_open_id VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255),
    nickname VARCHAR(50) NOT NULL DEFAULT '',
    avatar_url VARCHAR(500),
    bio TEXT DEFAULT '',
    role VARCHAR(20) NOT NULL DEFAULT 'user', -- user, admin, moderator, service_provider
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, banned, suspended
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_wechat ON users(wechat_openid);
CREATE INDEX idx_users_qq ON users(qq_openid);

-- Categories (for product zones)
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    zone VARCHAR(20) NOT NULL, -- cosplay, peripheral
    parent_id UUID REFERENCES categories(id),
    icon_url VARCHAR(500),
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_zone ON categories(zone);

-- Products
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    seller_id UUID NOT NULL REFERENCES users(id),
    category_id UUID REFERENCES categories(id),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    original_price DECIMAL(10,2),
    currency VARCHAR(3) NOT NULL DEFAULT 'CNY',
    zone VARCHAR(20) NOT NULL, -- cosplay, peripheral
    source_type VARCHAR(20) NOT NULL DEFAULT 'self_made', -- official, agent, self_made
    images JSONB NOT NULL DEFAULT '[]', -- array of image URLs
    stock INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, draft, sold_out, banned
    tags JSONB NOT NULL DEFAULT '[]',
    character_name VARCHAR(100),
    anime_name VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_zone ON products(zone);
CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_products_seller ON products(seller_id);
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_anime ON products(anime_name);

-- Orders
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    order_no VARCHAR(32) NOT NULL UNIQUE,
    total_amount DECIMAL(10,2) NOT NULL,
    payment_method VARCHAR(20), -- wechat, alipay
    payment_status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, paid, failed, refunded
    order_status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, paid, shipped, received, completed, cancelled
    shipping_address JSONB,
    payment_id VARCHAR(255), -- external payment transaction ID
    idempotency_key VARCHAR(64) UNIQUE,
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(order_status);
CREATE INDEX idx_orders_payment ON orders(payment_status);

-- Order items
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    quantity INT NOT NULL DEFAULT 1,
    price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Posts (community)
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    title VARCHAR(200) NOT NULL DEFAULT '',
    content TEXT NOT NULL,
    images JSONB NOT NULL DEFAULT '[]',
    video_url VARCHAR(500),
    type VARCHAR(20) NOT NULL DEFAULT 'text', -- text, image, video
    tags JSONB NOT NULL DEFAULT '[]',
    like_count INT NOT NULL DEFAULT 0,
    comment_count INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending_review', -- pending_review, approved, rejected
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_posts_user ON posts(user_id);
CREATE INDEX idx_posts_status ON posts(status);
CREATE INDEX idx_posts_created ON posts(created_at DESC);

-- Comments
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    parent_id UUID REFERENCES comments(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending_review',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comments_post ON comments(post_id);
CREATE INDEX idx_comments_user ON comments(user_id);

-- Likes
CREATE TABLE likes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    post_id UUID NOT NULL REFERENCES posts(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, post_id)
);

CREATE INDEX idx_likes_post ON likes(post_id);

-- Groups (interest circles)
CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    cover_url VARCHAR(500),
    owner_id UUID NOT NULL REFERENCES users(id),
    member_count INT NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Group members
CREATE TABLE group_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    role VARCHAR(20) NOT NULL DEFAULT 'member', -- owner, admin, member
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(group_id, user_id)
);

-- Events (comic cons, activities)
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    cover_url VARCHAR(500),
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    address VARCHAR(500) NOT NULL,
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    source VARCHAR(20) NOT NULL DEFAULT 'manual', -- manual, ai_scraped
    status VARCHAR(20) NOT NULL DEFAULT 'upcoming', -- upcoming, ongoing, finished, cancelled
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_events_time ON events(start_time);
CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_location ON events(latitude, longitude);

-- Service providers (makeup artists, photographers)
CREATE TABLE service_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    type VARCHAR(20) NOT NULL, -- makeup, photography, stall
    description TEXT,
    portfolio_images JSONB NOT NULL DEFAULT '[]',
    price_list JSONB NOT NULL DEFAULT '[]',
    rating DECIMAL(3,2) NOT NULL DEFAULT 0,
    review_count INT NOT NULL DEFAULT 0,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_service_providers_type ON service_providers(type);
CREATE INDEX idx_service_providers_rating ON service_providers(rating DESC);

-- Bookings
CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    provider_id UUID NOT NULL REFERENCES service_providers(id),
    event_id UUID REFERENCES events(id),
    booked_time TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, confirmed, completed, cancelled
    notes TEXT,
    amount DECIMAL(10,2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bookings_user ON bookings(user_id);
CREATE INDEX idx_bookings_provider ON bookings(provider_id);

-- Content reports (moderation)
CREATE TABLE content_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reporter_id UUID NOT NULL REFERENCES users(id),
    content_type VARCHAR(20) NOT NULL, -- post, comment, product, profile
    content_id UUID NOT NULL,
    reason VARCHAR(200) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, reviewed, resolved
    moderator_id UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Refresh tokens
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    token VARCHAR(500) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
