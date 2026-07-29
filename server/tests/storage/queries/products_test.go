package queries_test

import (
	"context"
	"fmt"
	"generic-shop-sample/internal/config"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	tu "generic-shop-sample/tests/internal/testutils"
	"testing"
)

func TestProductStore(t *testing.T) {
	ctx := t.Context()
	config := tu.ConfigTestSetup()
	session := tu.DBTestSetup(ctx, config).GetSession()
	if _, err := session.Exec(ctx, "TRUNCATE product_s.products RESTART IDENTITY CASCADE"); err != nil {
		t.Errorf("failed to truncate products, %s", err)
	}
	log := logger.GetLogger()
	store := queries.NewProductStore(session, log)

	productInstance := queries.CreateProductRequest{
		Name:         "TestProduct",
		Description:  "Test product description",
		CommonDetail: queries.ProductProperty{"color": "red", "size": "M"},
	}
	s := testProductStore{
		ctx:             ctx,
		session:         session,
		store:           store,
		productInstance: productInstance,
	}

	s.create(t)
	s.getList(t)
	s.setVariantDetail(t)
	s.get(t)
	s.update(t)
	s.setActive(t)
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
	if err := s.store.Create(s.ctx, s.productInstance); err != nil {
		t.Fatalf("failed to create product, %s", err)
	}
}

func (s *testProductStore) getList(t *testing.T) {
	if _, err := s.session.Exec(s.ctx, "UPDATE product_s.products SET is_active = true"); err != nil {
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
	if s.summaryResponse.Name != s.productInstance.Name {
		t.Errorf("list mismatch: got %+v, expected name=%q", s.summaryResponse, s.productInstance.Name)
	}
}

func (s *testProductStore) setVariantDetail(t *testing.T) {
	property := queries.ProductProperty{"info": "some info"}
	err := s.store.SetVariantDetail(s.ctx, s.summaryResponse.ID, []queries.ProductVariantDetail{
		{
			ProductPropertyRequest: queries.ProductPropertyRequest{
				Property: property,
			},
			Price: 100,
			Vendors: []queries.ProductVendor{
				{
					UserID:   "123",
					Quantity: 5,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to set variant detail: %s", err)
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
		ProductIDRequest: queries.ProductIDRequest{
			ID: s.summaryResponse.ID,
		},
		CreateProductRequest: queries.CreateProductRequest{
			Name:         "TestProduct",
			Description:  "New Test product description",
			CommonDetail: queries.ProductProperty{"color": "red", "size": "M"},
		},
	}
	if err := s.store.Update(s.ctx, uProduct); err != nil {
		t.Fatalf("failed to update product, %s", err)
	}

	updated, err := s.store.Get(s.ctx, s.summaryResponse.ID)
	if err != nil {
		t.Fatalf("failed to get updated product, %s", err)
	}
	if updated.Description != uProduct.Description || updated.IsAvailable != true || updated.Price != 100 {
		t.Errorf("update mismatch: got %+v, expected description=%q, isAvailable=true, price=100",
			updated, uProduct.Description)
	}
}

func (s *testProductStore) setActive(t *testing.T) {
	if err := s.store.SetActive(s.ctx, s.summaryResponse.ID, false); err != nil {
		t.Fatalf("failed to set active: %s", err)
	}

	check, err := s.store.Get(s.ctx, s.summaryResponse.ID)
	if err != nil {
		t.Fatalf("failed to get product after state changes: %s", err)
	}
	if check.IsActive != false || check.IsAvailable != true {
		t.Errorf("expected isActive=false, isAvailable=true; got %+v", check)
	}
}

func (s *testProductStore) delete(t *testing.T) {
	if err := s.store.Delete(s.ctx, s.summaryResponse.ID); err != nil {
		t.Fatalf("failed to delete product, %s", err)
	}

	_, err := s.store.Get(s.ctx, s.summaryResponse.ID)
	if err == nil {
		t.Errorf("expected error when getting deleted product, got nil")
	}
}

func TestFullListProducts(t *testing.T) {
	ctx := t.Context()
	config := tu.ConfigTestSetup()
	session := tu.DBTestSetup(ctx, config).GetSession()
	if _, err := session.Exec(ctx, "TRUNCATE product_s.products RESTART IDENTITY CASCADE"); err != nil {
		t.Errorf("failed to truncate products")
	}
	log := logger.GetLogger()
	store := queries.NewProductStore(session, log)

	description := "some descriptions"
	info := queries.ProductProperty{"info": "some info"}
	products := []struct {
		product queries.CreateProductRequest
	}{
		{queries.CreateProductRequest{Name: "item 1", Description: description, CommonDetail: info}},
		{queries.CreateProductRequest{Name: "item 2", Description: description, CommonDetail: info}},
		{queries.CreateProductRequest{Name: "item 3", Description: description, CommonDetail: info}},
		{queries.CreateProductRequest{Name: "item 4", Description: description, CommonDetail: info}},
		{queries.CreateProductRequest{Name: "item 5", Description: description, CommonDetail: info}},
	}
	for i, product := range products {
		if err := store.Create(ctx, product.product); err != nil {
			t.Errorf(`failed to create product "%d", %s`, i, err)
		}
	}

	items, err := store.AdminList(ctx, 6, 1)
	if err != nil {
		t.Fatalf("failed to get product admin list: %s", err)
	}
	if l := len(items); l != 5 {
		t.Fatalf(`invalid product list length, got="%d"`, l)
	}
}

func TestProductImagesStore(t *testing.T) {
	ctx := t.Context()
	config := tu.ConfigTestSetup()
	session := tu.DBTestSetup(ctx, config).GetSession()
	log := logger.GetLogger()
	productStore := queries.NewProductStore(session, log)
	store := queries.NewProductImagesStore(session, log, config.ProductImage)

	if _, err := session.Exec(ctx, "TRUNCATE product_s.products RESTART IDENTITY CASCADE"); err != nil {
		t.Errorf("failed to truncate products, %s", err)
	}

	if err := productStore.Create(ctx, queries.CreateProductRequest{
		Name:         "new product",
		Description:  "some desc",
		CommonDetail: queries.ProductProperty{"info": "some info"},
	}); err != nil {
		t.Errorf("failed to create product, %s", err)
	}
	if _, err := session.Exec(ctx, "UPDATE product_s.products SET is_active = true"); err != nil {
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
	config          config.Config
	store           queries.ProductImagesStore
	summaryResponse queries.ProductSummaryResponse
}

func (s *testProductImagesStore) create(t *testing.T) {
	productImages := make([]string, s.config.ProductImage.MaxProductImagesAmount)
	for i := range productImages {
		productImages[i] = fmt.Sprintf("path/to/img%d.img", i)
	}

	for i, imgPath := range productImages {
		if err := s.store.Create(s.ctx, s.summaryResponse.ID, imgPath); err != nil {
			if i == s.config.ProductImage.MaxProductImagesAmount {
				break
			}
			t.Errorf("failed to create product image row, %s", err)
		}
		if i == s.config.ProductImage.MaxProductImagesAmount {
			t.Errorf("failed to return error for creating more than max value")
		}
	}
}

func (s *testProductImagesStore) list(t *testing.T) {
	productImages, err := s.store.List(s.ctx, s.summaryResponse.ID)
	if err != nil {
		t.Errorf("failed to get product images list, %s", err)
	}
	if got := len(productImages); got != s.config.ProductImage.MaxProductImagesAmount {
		t.Errorf(`expected list return "%d" product images, but got "%d"`, s.config.ProductImage.MaxProductImagesAmount, got)
	}
	if imgPath, err := s.store.Delete(s.ctx, productImages[0].ID); err == nil {
		if imgPath != productImages[0].ImgPath {
			t.Error("failed to match imgPath")
		}
	} else {
		t.Errorf("failed to delete product image, %s", err)
	}
}
