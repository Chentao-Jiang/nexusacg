package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/planforever/nexusacg/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=nexusacg password=nexusacg_dev_pass dbname=nexusacg sslmode=disable TimeZone=Asia/Shanghai"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("database not available: %v", err)
	}

	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	db.AutoMigrate(
		&model.User{}, &model.Product{}, &model.Category{},
		&model.Post{}, &model.Comment{}, &model.Like{},
		&model.Event{}, &model.Order{}, &model.OrderItem{},
		&model.PaymentLog{}, &model.RefreshToken{},
	)

	// Clean tables (reverse FK order)
	db.Exec("DELETE FROM order_items")
	db.Exec("DELETE FROM products")
	db.Exec("DELETE FROM categories")
	db.Exec("DELETE FROM users")

	// Create a test seller
	testUser := model.User{
		Nickname: "test_seller",
		Role:     "seller",
		Status:   "active",
	}
	db.FirstOrCreate(&testUser, model.User{Nickname: "test_seller"})

	return db
}

func testSellerID(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()
	var user model.User
	if err := db.Where("nickname = ?", "test_seller").First(&user).Error; err != nil {
		t.Fatalf("test seller not found: %v", err)
	}
	return user.ID
}

func TestProductSearch_ByKeyword(t *testing.T) {
	db := testDB(t)
	svc := NewProductService(db)
	ctx := context.Background()
	sellerID := testSellerID(t, db)
	svc.Create(ctx, CreateProductInput{
		SellerID: sellerID, Name: "初音ミク cosplay 服装", Description: "初音未来经典款",
		Price: 299, Zone: "cosplay", SourceType: "self_made", Stock: 10,
	})
	svc.Create(ctx, CreateProductInput{
		SellerID: sellerID, Name: "原神 雷电将军 手办", Description: "雷电将军景品",
		Price: 150, Zone: "peripheral", SourceType: "official", Stock: 5,
	})

	// Search for 初音
	result, err := svc.List(ctx, ProductListInput{Keyword: "初音", Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("search by keyword failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 result for '初音', got %d", result.Total)
	}
	if result.Total > 0 && result.Items[0].Name != "初音ミク cosplay 服装" {
		t.Errorf("expected 初音 product, got %s", result.Items[0].Name)
	}

	// Search for 原神
	result, err = svc.List(ctx, ProductListInput{Keyword: "原神", Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("search by keyword failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 result for '原神', got %d", result.Total)
	}

	t.Log("keyword search works correctly")
}

func TestProductSearch_ByTags(t *testing.T) {
	db := testDB(t)
	svc := NewProductService(db)
	ctx := context.Background()

	sellerID := testSellerID(t, db)
	svc.Create(ctx, CreateProductInput{
		SellerID: sellerID, Name: "商品A", Price: 100, Zone: "cosplay", SourceType: "self_made",
		Stock: 10, Tags: []string{"新款", "热卖"},
	})
	svc.Create(ctx, CreateProductInput{
		SellerID: sellerID, Name: "商品B", Price: 200, Zone: "cosplay", SourceType: "self_made",
		Stock: 5, Tags: []string{"限量"},
	})
	svc.Create(ctx, CreateProductInput{
		SellerID: sellerID, Name: "商品C", Price: 300, Zone: "cosplay", SourceType: "self_made",
		Stock: 3, Tags: []string{"新款", "限量"},
	})

	// Filter by tag "新款"
	result, err := svc.List(ctx, ProductListInput{Tags: "新款", Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("filter by tags failed: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected 2 products with tag '新款', got %d", result.Total)
	}

	// Filter by multiple tags "新款" AND "限量"
	result, err = svc.List(ctx, ProductListInput{Tags: "新款,限量", Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("filter by multiple tags failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 product with both tags, got %d", result.Total)
	}

	t.Log("tag filtering works correctly")
}

func TestProductSearch_ByPriceRange(t *testing.T) {
	db := testDB(t)
	svc := NewProductService(db)
	ctx := context.Background()

	sellerID := testSellerID(t, db)
	svc.Create(ctx, CreateProductInput{
		SellerID: sellerID, Name: "低价商品", Price: 50, Zone: "cosplay", SourceType: "self_made", Stock: 10,
	})
	svc.Create(ctx, CreateProductInput{
		SellerID: sellerID, Name: "中等商品", Price: 200, Zone: "cosplay", SourceType: "self_made", Stock: 5,
	})
	svc.Create(ctx, CreateProductInput{
		SellerID: sellerID, Name: "高价商品", Price: 500, Zone: "cosplay", SourceType: "self_made", Stock: 3,
	})

	// Price range 100-300
	result, err := svc.List(ctx, ProductListInput{MinPrice: 100, MaxPrice: 300, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("filter by price range failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 product in price range 100-300, got %d", result.Total)
	}

	t.Log("price range filtering works correctly")
}

func TestProductSearch_CombinedFilters(t *testing.T) {
	db := testDB(t)
	svc := NewProductService(db)
	ctx := context.Background()

	sellerID := testSellerID(t, db)
	svc.Create(ctx, CreateProductInput{
		SellerID: sellerID, Name: "cosplay 热卖 A", Price: 100, Zone: "cosplay",
		SourceType: "self_made", Stock: 10, Tags: []string{"热卖"},
	})
	svc.Create(ctx, CreateProductInput{
		SellerID: sellerID, Name: "cosplay 普通 B", Price: 200, Zone: "cosplay",
		SourceType: "self_made", Stock: 5, Tags: []string{"普通"},
	})
	svc.Create(ctx, CreateProductInput{
		SellerID: sellerID, Name: "peripheral 热卖 C", Price: 150, Zone: "peripheral",
		SourceType: "official", Stock: 3, Tags: []string{"热卖"},
	})

	// Zone + Tag combined
	result, err := svc.List(ctx, ProductListInput{Zone: "cosplay", Tags: "热卖", Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("combined filters failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 product with zone=cosplay AND tag=热卖, got %d", result.Total)
	}

	t.Log("combined filters work correctly")
}

func TestProductSearch_CategoryFilter(t *testing.T) {
	db := testDB(t)
	svc := NewProductService(db)
	catSvc := NewCategoryService(db)
	ctx := context.Background()

	// Create category
	cat, err := catSvc.Create(ctx, CreateCategoryInput{
		Name: "连衣裙", Zone: "cosplay", SortOrder: 1,
	})
	if err != nil {
		t.Fatalf("create category failed: %v", err)
	}

	sellerID := testSellerID(t, db)
	svc.Create(ctx, CreateProductInput{
		SellerID: sellerID, CategoryID: &cat.ID, Name: "连衣裙商品",
		Price: 300, Zone: "cosplay", SourceType: "self_made", Stock: 10,
	})
	svc.Create(ctx, CreateProductInput{
		SellerID: sellerID, Name: "非分类商品",
		Price: 100, Zone: "cosplay", SourceType: "self_made", Stock: 5,
	})

	// Filter by category
	result, err := svc.List(ctx, ProductListInput{CategoryID: cat.ID.String(), Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("filter by category failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 product in category, got %d", result.Total)
	}

	// Verify category_name is populated
	if result.Total > 0 && result.Items[0].CategoryName != "连衣裙" {
		t.Errorf("expected category_name='连衣裙', got '%s'", result.Items[0].CategoryName)
	}

	t.Log("category filtering with join works correctly")
}

func TestCategory_CRUD(t *testing.T) {
	db := testDB(t)
	svc := NewCategoryService(db)
	ctx := context.Background()

	// Create
	cat, err := svc.Create(ctx, CreateCategoryInput{
		Name: "测试分类", Zone: "cosplay", SortOrder: 1,
	})
	if err != nil {
		t.Fatalf("create category failed: %v", err)
	}
	if cat.Name != "测试分类" {
		t.Errorf("expected name '测试分类', got '%s'", cat.Name)
	}

	// Get
	got, err := svc.Get(ctx, cat.ID)
	if err != nil {
		t.Fatalf("get category failed: %v", err)
	}
	if got.Name != cat.Name {
		t.Errorf("expected name '%s', got '%s'", cat.Name, got.Name)
	}

	// Update
	newName := "更新后的分类"
	updated, err := svc.Update(ctx, cat.ID, UpdateCategoryInput{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("update category failed: %v", err)
	}
	if updated.Name != newName {
		t.Errorf("expected updated name '%s', got '%s'", newName, updated.Name)
	}

	// List
	categories, err := svc.List(ctx, CategoryListInput{Zone: "cosplay"})
	if err != nil {
		t.Fatalf("list categories failed: %v", err)
	}
	if len(categories) < 1 {
		t.Errorf("expected at least 1 category, got %d", len(categories))
	}

	// Delete
	if err := svc.Delete(ctx, cat.ID); err != nil {
		t.Fatalf("delete category failed: %v", err)
	}

	// Verify deleted
	_, err = svc.Get(ctx, cat.ID)
	if err == nil {
		t.Error("expected error getting deleted category")
	}

	t.Log("category CRUD works correctly")
}

func TestProductSearch_SortOptions(t *testing.T) {
	db := testDB(t)
	svc := NewProductService(db)
	ctx := context.Background()

	sellerID := testSellerID(t, db)
	svc.Create(ctx, CreateProductInput{
		SellerID: sellerID, Name: "便宜", Price: 10, Zone: "cosplay", SourceType: "self_made", Stock: 10,
	})
	svc.Create(ctx, CreateProductInput{
		SellerID: sellerID, Name: "中等", Price: 100, Zone: "cosplay", SourceType: "self_made", Stock: 5,
	})
	svc.Create(ctx, CreateProductInput{
		SellerID: sellerID, Name: "贵", Price: 500, Zone: "cosplay", SourceType: "self_made", Stock: 3,
	})

	// Sort by price asc
	result, err := svc.List(ctx, ProductListInput{Sort: "price_asc", Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("sort by price_asc failed: %v", err)
	}
	if result.Items[0].Price != 10 {
		t.Errorf("expected first item price 10, got %f", result.Items[0].Price)
	}

	// Sort by price desc
	result, err = svc.List(ctx, ProductListInput{Sort: "price_desc", Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("sort by price_desc failed: %v", err)
	}
	if result.Items[0].Price != 500 {
		t.Errorf("expected first item price 500, got %f", result.Items[0].Price)
	}

	t.Log("sort options work correctly")
}
