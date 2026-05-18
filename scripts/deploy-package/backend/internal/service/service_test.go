package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/gorm"
)

func testUser(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()
	user := model.User{
		Nickname: "test_svc_user",
		Role:     "user",
		Status:   "active",
	}
	db.FirstOrCreate(&user, model.User{Nickname: "test_svc_user"})
	return user.ID
}

func TestPost_CreateAndList(t *testing.T) {
	db := testDB(t)
	svc := NewPostService(db)
	ctx := context.Background()

	db.Where("title LIKE 'svc_test%'").Delete(&model.Post{})
	uid := testUser(t, db)

	post, err := svc.Create(ctx, CreatePostInput{
		UserID:  uid,
		Title:   "svc_test post one",
		Content: "test content here",
		Tags:    []string{"test", "golang"},
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}
	if post.ID == uuid.Nil {
		t.Fatal("post ID is nil")
	}
	if post.Status != "pending_review" {
		t.Errorf("expected pending_review, got: %s", post.Status)
	}
	if post.Type != "text" {
		t.Errorf("expected text type, got: %s", post.Type)
	}

	db.Model(&model.Post{}).Where("id = ?", post.ID).Update("status", "approved")

	result, err := svc.List(ctx, PostListInput{PageSize: 20})
	if err != nil {
		t.Fatalf("list posts: %v", err)
	}
	if result.Total < 1 {
		t.Fatal("no posts listed")
	}

	got, err := svc.Get(ctx, post.ID)
	if err != nil {
		t.Fatalf("get post: %v", err)
	}
	if got.ID != post.ID {
		t.Errorf("expected ID %s, got %s", post.ID, got.ID)
	}
}

func TestPost_LikeAndUnlike(t *testing.T) {
	db := testDB(t)
	svc := NewPostService(db)
	ctx := context.Background()

	uid := testUser(t, db)
	db.Where("title LIKE 'svc_test%'").Delete(&model.Post{})
	db.Where("user_id = ?", uid).Where("post_id IS NOT NULL").Delete(&model.Like{})

	post, _ := svc.Create(ctx, CreatePostInput{
		UserID:  uid,
		Title:   "svc_test like post",
		Content: "like test",
	})
	db.Model(&model.Post{}).Where("id = ?", post.ID).Update("status", "approved")

	if err := svc.Like(ctx, uid, post.ID); err != nil {
		t.Fatalf("like post: %v", err)
	}

	if err := svc.Unlike(ctx, uid, post.ID); err != nil {
		t.Fatalf("unlike post: %v", err)
	}

	if err := svc.Unlike(ctx, uid, post.ID); err == nil {
		t.Fatal("double unlike should have failed")
	}
}

func TestPost_CreateComment(t *testing.T) {
	db := testDB(t)
	svc := NewPostService(db)
	ctx := context.Background()

	uid := testUser(t, db)
	db.Where("title LIKE 'svc_test%'").Delete(&model.Post{})
	db.Where("user_id = ?", uid).Delete(&model.Comment{})

	post, _ := svc.Create(ctx, CreatePostInput{
		UserID:  uid,
		Title:   "svc_test comment post",
		Content: "comment test",
	})

	comment, err := svc.CreateComment(ctx, CommentInput{
		PostID:  post.ID,
		UserID:  uid,
		Content: "test comment",
	})
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}
	if comment.Status != "pending_review" {
		t.Errorf("expected pending_review, got: %s", comment.Status)
	}

	// Reply with nil parent (valid - top-level reply to same post)
	_, err = svc.CreateComment(ctx, CommentInput{
		PostID:  post.ID,
		UserID:  uid,
		Content: "another top comment",
	})
	if err != nil {
		t.Fatalf("create top-level comment: %v", err)
	}

	// Reply with valid parent
	parentID := comment.ID
	reply, err := svc.CreateComment(ctx, CommentInput{
		PostID:   post.ID,
		UserID:   uid,
		Content:  "reply comment",
		ParentID: &parentID,
	})
	if err != nil {
		t.Fatalf("create reply: %v", err)
	}
	if *reply.ParentID != parentID {
		t.Error("reply parent_id not set correctly")
	}

	// Reply to wrong post should fail
	otherPost, _ := svc.Create(ctx, CreatePostInput{
		UserID:  uid,
		Title:   "svc_test other",
		Content: "other",
	})
	_, err = svc.CreateComment(ctx, CommentInput{
		PostID:   otherPost.ID,
		UserID:   uid,
		Content:  "wrong post reply",
		ParentID: &parentID,
	})
	if err == nil {
		t.Fatal("expected error for cross-post reply")
	}
}

func TestPost_Search(t *testing.T) {
	db := testDB(t)
	svc := NewPostService(db)
	ctx := context.Background()

	uid := testUser(t, db)
	db.Where("title LIKE 'svc_test_search%'").Delete(&model.Post{})

	for _, title := range []string{"svc_test_search anime", "svc_test_search manga", "svc_test_search game"} {
		post, _ := svc.Create(ctx, CreatePostInput{
			UserID:  uid,
			Title:   title,
			Content: "search test content",
		})
		db.Model(&model.Post{}).Where("id = ?", post.ID).Update("status", "approved")
	}

	result, err := svc.List(ctx, PostListInput{Keyword: "anime", PageSize: 20})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if result.Total < 1 {
		t.Fatal("no results for keyword search")
	}

	result, err = svc.List(ctx, PostListInput{Keyword: "search test", PageSize: 20})
	if err != nil {
		t.Fatalf("multi-word search: %v", err)
	}
	if result.Total < 3 {
		t.Fatalf("expected at least 3 results for multi-word search, got %d", result.Total)
	}
}

func TestEvent_CreateWithTimes(t *testing.T) {
	db := testDB(t)
	svc := NewEventService(db)
	ctx := context.Background()

	db.Where("name LIKE 'svc_test_event%'").Delete(&model.Event{})

	event, err := svc.Create(ctx, CreateEventInput{
		Name:      "svc_test_event_2026",
		Address:   "Shanghai Convention Center",
		StartTime: "2026-10-01T09:00:00Z",
		EndTime:   "2026-10-03T18:00:00Z",
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	if event.StartTime.IsZero() {
		t.Fatal("start time is zero")
	}
	if event.EndTime.IsZero() {
		t.Fatal("end time is zero")
	}

	got, err := svc.Get(ctx, event.ID)
	if err != nil {
		t.Fatalf("get event: %v", err)
	}
	if got.Name != event.Name {
		t.Errorf("event name mismatch: expected %s, got %s", event.Name, got.Name)
	}

	result, err := svc.List(ctx, EventListInput{PageSize: 20})
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if result.Total < 1 {
		t.Fatal("no events listed")
	}
}

func TestEvent_InvalidTimes(t *testing.T) {
	db := testDB(t)
	svc := NewEventService(db)
	ctx := context.Background()

	_, err := svc.Create(ctx, CreateEventInput{
		Name:      "bad time event",
		Address:   "Nowhere",
		StartTime: "not-a-date",
		EndTime:   "2026-10-03T18:00:00Z",
	})
	if err == nil {
		t.Fatal("expected error for invalid start time")
	}
}

func TestOrder_CreateAndCancel(t *testing.T) {
	db := testDB(t)
	svc := NewOrderService(db)
	ctx := context.Background()

	uid := testUser(t, db)
	db.Where("name LIKE 'svc_test_order_prod%'").Delete(&model.Product{})
	db.Where("name LIKE 'svc_test_low_stock%'").Delete(&model.Product{})

	product := model.Product{
		ID:       uuid.New(),
		SellerID: uid,
		Name:     "svc_test_order_product",
		Price:    99.0,
		Stock:    10,
		Status:   "active",
		Zone:     "cosplay",
	}
	db.Create(&product)

	order, err := svc.Create(ctx, CreateOrderInput{
		UserID: uid,
		Items: []struct {
			ProductID uuid.UUID `json:"product_id"`
			Quantity  int       `json:"quantity"`
		}{{ProductID: product.ID, Quantity: 2}},
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if order.OrderStatus != "pending" {
		t.Errorf("expected pending status, got: %s", order.OrderStatus)
	}

	var updated model.Product
	db.Where("id = ?", product.ID).First(&updated)
	if updated.Stock != 8 {
		t.Errorf("expected stock 8 after deducting 2, got %d", updated.Stock)
	}

	if err := svc.Cancel(ctx, uid, order.OrderNo); err != nil {
		t.Fatalf("cancel order: %v", err)
	}

	db.Where("id = ?", product.ID).First(&updated)
	if updated.Stock != 10 {
		t.Errorf("expected stock 10 after cancel, got %d", updated.Stock)
	}
}

func TestOrder_CreateWithInsufficientStock(t *testing.T) {
	db := testDB(t)
	svc := NewOrderService(db)
	ctx := context.Background()

	uid := testUser(t, db)

	product := model.Product{
		ID:       uuid.New(),
		SellerID: uid,
		Name:     "svc_test_low_stock",
		Price:    50.0,
		Stock:    1,
		Status:   "active",
		Zone:     "merch",
	}
	db.Create(&product)

	_, err := svc.Create(ctx, CreateOrderInput{
		UserID: uid,
		Items: []struct {
			ProductID uuid.UUID `json:"product_id"`
			Quantity  int       `json:"quantity"`
		}{{ProductID: product.ID, Quantity: 5}},
	})
	if err == nil {
		t.Fatal("expected error for insufficient stock")
	}
}
