package queries_test

import (
	"context"
	"fmt"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	"testing"
)

func TestProductStore(t *testing.T) {
	ctx := t.Context()
	session := database.GetSession()
	if _, err := session.Exec(ctx, "TRUNCATE products RESTART IDENTITY CASCADE"); err != nil {
		t.Errorf("failed to truncate products, %s", err)
	}
	log := logger.GetLogger()
	store := queries.NewProductStore(session, log)

	productInstance := queries.CreateProductRequest{
		Name:        "TestProduct",
		Price:       1999,
		Description: "Test product description",
		Details:     `{"color": "red", "size": "M"}`,
		IsAvailable: true,
	}
	s := testProductStore{
		ctx:             ctx,
		session:         session,
		store:           store,
		productInstance: productInstance,
	}

	s.create(t)
	s.getList(t)
	s.incrDecrBy(t)
	s.get(t)
	s.update(t)
	s.setActiveAvailable(t)
	s.delete(t)
}

type testProductStore struct {
	ctx             context.Context
	session         database.Session
	store           queries.ProductStore
	productInstance queries.CreateProductRequest
	summaryResponse queries.ProductSummaryResponse
	userID          int32
}

func (s *testProductStore) create(t *testing.T) {
	if err := s.store.Create(s.ctx, 1, &s.productInstance); err != nil {
		t.Fatalf("failed to create product, %s", err)
	}
}

func (s *testProductStore) getList(t *testing.T) {
	if _, err := s.session.Exec(s.ctx, "UPDATE products SET is_active = true"); err != nil {
		t.Fatalf("failed to activate products")
	}

	list, err := s.store.List(s.ctx, 10, 1)
	if err != nil {
		t.Fatalf("failed to get peoducts list, %s", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 product in list, but got %d", len(list))
	}
	s.summaryResponse = list[0]
	if s.summaryResponse.Name != s.productInstance.Name || s.summaryResponse.Price != s.productInstance.Price {
		t.Errorf("list mismatch: got %+v, expected name=%q, price=%d", s.summaryResponse, s.productInstance.Name, s.productInstance.Price)
	}
}

func (s *testProductStore) incrDecrBy(t *testing.T) {
	if err := s.store.IncrBy(s.ctx, s.summaryResponse.ID, 1, 10); err != nil {
		t.Errorf("failed to incr by, %s", err)
	}
	if err := s.store.DecrBy(s.ctx, s.summaryResponse.ID, 1, 5); err != nil {
		t.Errorf("failed to decr by, %s", err)
	}
}

func (s *testProductStore) get(t *testing.T) {
	got, err := s.store.Get(s.ctx, s.summaryResponse.ID)
	if err != nil {
		t.Fatalf("failed to get product, %s", err)
	}
	if got.Name != s.productInstance.Name || got.Description != s.productInstance.Description || got.AvailableQuantity != 5 {
		t.Errorf(`get mismatch: got "%+v", expected name="%q", description="%q", available_quantity="%d"`,
			got, s.productInstance.Name, s.productInstance.Description, 5)
	}
}

func (s *testProductStore) update(t *testing.T) {
	uProduct := queries.UpdateProductRequest{
		s.summaryResponse.ID,
		queries.CreateProductRequest{
			Name:        "TestProduct",
			Price:       1999,
			Description: "New Test product description",
			Details:     `{"color": "red", "size": "M"}`,
			IsAvailable: false,
		},
	}
	if err := s.store.Update(s.ctx, 1, &uProduct); err != nil {
		t.Fatalf("failed to update product, %s", err)
	}

	updated, err := s.store.Get(s.ctx, s.summaryResponse.ID)
	if err != nil {
		t.Fatalf("failed to get updated product, %s", err)
	}
	if updated.Description != uProduct.Description || updated.IsAvailable != false {
		t.Errorf("update mismatch: got %+v, expected description=%q, isAvailable=false",
			updated, uProduct.Description)
	}
}

func (s *testProductStore) setActiveAvailable(t *testing.T) {
	if err := s.store.SetActive(s.ctx, s.summaryResponse.ID, false); err != nil {
		t.Fatalf("failed to set active: %s", err)
	}
	if err := s.store.SetAvailable(s.ctx, s.summaryResponse.ID, true); err != nil {
		t.Fatalf("failed to set available, %s", err)
	}

	check, err := s.store.Get(s.ctx, s.summaryResponse.ID)
	if err != nil {
		t.Fatalf("failed to get product after state changes: %s", err)
	}
	if check.IsActive != false || check.IsAvailable != true {
		t.Errorf("expected isActive=false, isAvailable=true; got %+v", check)
	}
	s.userID = check.UserID
}

func (s *testProductStore) delete(t *testing.T) {
	if err := s.store.Delete(s.ctx, s.summaryResponse.ID, s.userID); err != nil {
		t.Fatalf("failed to delete product, %s", err)
	}

	_, err := s.store.Get(s.ctx, s.summaryResponse.ID)
	if err == nil {
		t.Errorf("expected error when getting deleted product, got nil")
	}
}

func TestFullListProducts(t *testing.T) {
	ctx := t.Context()
	session := database.GetSession()
	if _, err := session.Exec(ctx, "TRUNCATE products RESTART IDENTITY CASCADE"); err != nil {
		t.Errorf("failed to truncate products")
	}
	log := logger.GetLogger()
	store := queries.NewProductStore(session, log)

	description := "some descriptions"
	info := `{"info": "some info"}`
	products := []struct {
		userID  int32
		product queries.CreateProductRequest
	}{
		{1, queries.CreateProductRequest{Name: "item 1", Price: 10, Description: description, Details: info, IsAvailable: true}},
		{1, queries.CreateProductRequest{Name: "item 2", Price: 10, Description: description, Details: info, IsAvailable: true}},
		{1, queries.CreateProductRequest{Name: "item 3", Price: 10, Description: description, Details: info, IsAvailable: true}},
		{2, queries.CreateProductRequest{Name: "item 4", Price: 10, Description: description, Details: info, IsAvailable: true}},
		{3, queries.CreateProductRequest{Name: "item 5", Price: 10, Description: description, Details: info, IsAvailable: true}},
	}
	for i, product := range products {
		if err := store.Create(ctx, product.userID, &product.product); err != nil {
			t.Errorf(`failed to create product "%d", %s`, i, err)
		}
	}

	if _, err := session.Exec(ctx, "UPDATE products SET is_active = true"); err != nil {
		t.Errorf("failed to active products, %s", err)
	}

	tests := []struct {
		name          string
		id            int32
		pagination    int
		page          int
		expectedCount int
	}{
		{"all products", 0, 10, 1, 5},
		{"adminUser products", 1, 4, 1, 3},
		{"invalidUser products", 999, 10, 1, 0},
		{"pagination test", 1, 10, 3, 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(st *testing.T) {
			items, err := store.FullList(ctx, test.id, test.pagination, test.page)
			if err != nil {
				if test.expectedCount == 0 {
					return
				}
				st.Errorf("unexpected error, %s", err)
				return
			}
			if got := len(items); got != test.expectedCount {
				t.Errorf(`expected "%d" products, but got "%d"`, test.expectedCount, got)
			}
		})
	}
}

func TestProductImagesStore(t *testing.T) {
	ctx := t.Context()
	session := database.GetSession()
	log := logger.GetLogger()
	config := internal.GetConfig()
	productStore := queries.NewProductStore(session, log)
	store := queries.NewProductImagesStore(session, log)

	if _, err := session.Exec(ctx, "TRUNCATE products RESTART IDENTITY CASCADE"); err != nil {
		t.Errorf("failed to truncate products, %s", err)
	}

	if err := productStore.Create(ctx, 1, &queries.CreateProductRequest{
		Name:        "new product",
		Price:       10,
		IsAvailable: true,
	}); err != nil {
		t.Errorf("failed to create product, %s", err)
	}
	if _, err := session.Exec(ctx, "UPDATE products SET is_active = true"); err != nil {
		t.Errorf("failed to activate products, %s", err)
	}
	var summaryResponse queries.ProductSummaryResponse
	if products, err := productStore.List(ctx, 10, 1); err == nil {
		summaryResponse = products[0]
	} else {
		t.Errorf("failed to get products list, %s", err)
	}

	s := testProductImagesStore{ctx, session, config, store, summaryResponse}

	s.create(t)
	s.list(t)
}

type testProductImagesStore struct {
	ctx             context.Context
	session         database.Session
	config          *internal.Config
	store           queries.ProductImagesStore
	summaryResponse queries.ProductSummaryResponse
}

func (s *testProductImagesStore) create(t *testing.T) {
	productImages := make([]string, s.config.Opt.MaxProductImagesAmount)
	for i := range productImages {
		productImages[i] = fmt.Sprintf("path/to/img%d.img", i)
	}

	for i, imgPath := range productImages {
		if err := s.store.Create(s.ctx, s.summaryResponse.ID, imgPath); err != nil {
			if i == s.config.Opt.MaxProductImagesAmount {
				break
			}
			t.Errorf("failed to create product image row, %s", err)
		}
		if i == s.config.Opt.MaxProductImagesAmount {
			t.Errorf("failed to return error for creating more than max value")
		}
	}
}

func (s *testProductImagesStore) list(t *testing.T) {
	productImages, err := s.store.List(s.ctx, s.summaryResponse.ID)
	if err != nil {
		t.Errorf("failed to get product images list, %s", err)
	}
	if got := len(productImages); got != s.config.Opt.MaxProductImagesAmount {
		t.Errorf(`expected list return "%d" product images, but got "%d"`, s.config.Opt.MaxProductImagesAmount, got)
	}
	if imgPath, err := s.store.Delete(s.ctx, productImages[0].ID); err == nil {
		if imgPath != productImages[0].ImgPath {
			t.Error("failed to match imgPath")
		}
	} else {
		t.Errorf("failed to delete product image, %s", err)
	}
}
