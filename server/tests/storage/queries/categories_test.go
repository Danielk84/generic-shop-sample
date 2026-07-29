package queries_test

import (
	"context"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	tu "generic-shop-sample/tests/internal/testutils"
	"testing"
)

func TestCategoryStore(t *testing.T) {
	ctx := t.Context()
	config := tu.ConfigTestSetup()
	session := tu.DBTestSetup(ctx, config).GetSession()
	log := logger.GetLogger()
	store := queries.NewCategoryStore(session, log)

	if _, err := session.Exec(ctx, "TRUNCATE product_s.categories RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate categories: %s", err)
	}

	s := testCategoryStore{ctx: ctx, session: session, store: store}

	s.create(t)
	s.list(t)
	s.delete(t)
}

type testCategoryStore struct {
	ctx      context.Context
	session  database.Session
	store    queries.CategoryStore
	category queries.Category
}

func (s *testCategoryStore) create(t *testing.T) {
	c1 := &queries.Category{CategoryTag: queries.CategoryTag{Tag: "electronics"}}
	if err := s.store.Create(s.ctx, c1.Tag); err != nil {
		t.Fatalf("failed to create category: %s", err)
	}
	c2 := &queries.CategoryTag{Tag: "books"}
	if err := s.store.Create(s.ctx, c2.Tag); err != nil {
		t.Fatalf("failed to create category: %s", err)
	}
}

func (s *testCategoryStore) list(t *testing.T) {
	list, err := s.store.List(s.ctx)
	if err != nil {
		t.Fatalf("failed to categories list: %s", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(list))
	}

	expectedTags := map[string]bool{"electronics": true, "books": true}
	for _, cat := range list {
		if !expectedTags[cat.Tag] {
			t.Errorf("unexpected category tag in list: %q", cat.Tag)
		}
		if cat.Tag == "electronics" {
			s.category.ID = cat.ID
		}
	}
}

func (s *testCategoryStore) delete(t *testing.T) {
	if err := s.store.Delete(s.ctx, s.category.ID); err != nil {
		t.Fatalf("failed to delete category id=%d, %s", s.category.ID, err)
	}

	list, err := s.store.List(s.ctx)
	if err != nil {
		t.Fatalf("failed to list categories after delete, %s", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 category after deletion, got %d", len(list))
	}
	if list[0].Tag != "books" {
		t.Errorf("expected remaining tag 'books', got %q", list[0].Tag)
	}

	if err := s.store.Delete(s.ctx, 9999); err == nil {
		t.Errorf("expected error deleting non-existing category, but got nil")
	}
}

func TestPCStore(t *testing.T) {
	ctx := t.Context()
	config := tu.ConfigTestSetup()
	session := tu.DBTestSetup(ctx, config).GetSession()
	log := logger.GetLogger()

	_, err := session.Exec(ctx, "TRUNCATE product_s.products, product_s.categories RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("failed to truncate products_categories: %s", err)
	}

	categoryStore := queries.NewCategoryStore(session, log)
	for _, tag := range []string{"1", "2", "3", "4", "5"} {
		if err := categoryStore.Create(ctx, tag); err != nil {
			t.Fatalf("failed to create category tags: %s", err)
		}
	}

	productStore := queries.NewProductStore(session, log)
	product := queries.CreateProductRequest{
		Name:        "item 1",
		Description: "some description",
		CommonDetail: queries.ProductProperty{
			"info": "some info",
		},
	}
	if err := productStore.Create(ctx, product); err != nil {
		t.Fatalf("failed to create product: %s", err)
	}

	if _, err := session.Exec(ctx, "UPDATE product_s.products SET is_active = true"); err != nil {
		t.Fatalf("failed to active products, %s", err)
	}
	products, err := productStore.List(ctx, 1, 1)
	if err != nil {
		t.Fatalf("unexpected error in list method: %s", err)
	}
	productID := products[0].ID

	store := queries.NewPCStore(session, log)
	initialTags := []string{"1", "2", "3"}
	if err := store.SetTags(ctx, productID, initialTags); err != nil {
		t.Fatalf("SetTags failed: %s", err)
	}

	got, err := store.List(ctx, productID)
	if err != nil {
		t.Fatalf("List failed: %s", err)
	}
	if len(got) != len(initialTags) {
		t.Fatalf("expected %d tags, got %d", len(initialTags), len(got))
	}

	expected := map[string]bool{"1": true, "2": true, "3": true}
	for _, tag := range got {
		if !expected[tag] {
			t.Fatalf("unexpected tag %q found", tag)
		}
	}
	newTags := []string{"4", "5"}
	if err := store.SetTags(ctx, productID, newTags); err != nil {
		t.Fatalf("SetTags overwrite failed: %s", err)
	}

	got, err = store.List(ctx, productID)
	if err != nil {
		t.Fatalf("List failed after overwrite: %s", err)
	}
	if len(got) != len(newTags) {
		t.Fatalf("expected %d tags after overwrite, got %d", len(newTags), len(got))
	}
	expected = map[string]bool{"4": true, "5": true}
	for _, tag := range got {
		if !expected[tag] {
			t.Fatalf("unexpected tag after overwrite: %q", tag)
		}
	}
}
