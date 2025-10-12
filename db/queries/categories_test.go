package queries_test

import (
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	"testing"
)

func TestCategoryStore(t *testing.T) {
	ctx := t.Context()

	session := db.NewSession()
	cs := queries.NewCategoryStore(session)

	if _, err := session.Exec(ctx, "TRUNCATE categories RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate categories: %s", err)
	}

	c1 := &queries.Category{CategoryTag: queries.CategoryTag{"electronics"}}
	if err := cs.Create(ctx, c1.Tag); err != nil {
		t.Fatalf("failed to create category: %s", err)
	}
	c2 := &queries.CategoryTag{Tag: "books"}
	if err := cs.Create(ctx, c2.Tag); err != nil {
		t.Fatalf("failed to create category: %s", err)
	}

	list, err := cs.List(ctx)
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
			c1.ID = cat.ID
		}
	}

	if err := cs.Delete(ctx, c1.ID); err != nil {
		t.Fatalf("failed to delete category id=%d, %s", c1.ID, err)
	}

	list, err = cs.List(ctx)
	if err != nil {
		t.Fatalf("failed to list categories after delete, %s", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 category after deletion, got %d", len(list))
	}
	if list[0].Tag != "books" {
		t.Errorf("expected remaining tag 'books', got %q", list[0].Tag)
	}

	if err := cs.Delete(ctx, 9999); err == nil {
		t.Errorf("expected error deleting non-existing category, but got nil")
	}
}

func TestPCStore(t *testing.T) {
	ctx := t.Context()

	session := db.NewSession()

	if _, err := session.Exec(ctx, "TRUNCATE products, categories RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate products_categories: %s", err)
	}

	cs := queries.NewCategoryStore(session)
	for _, tag := range []string{"1", "2", "3", "4", "5"} {
		if err := cs.Create(ctx, tag); err != nil {
			t.Error("failed to create category tags", err)
		}
	}

	ps := queries.NewProductStore(session)
	product := queries.CreateProductRequest{Name: "item 1", Price: 10, Description: "some description", Details: `{"info": "some info"}`, IsAvailable: true}
	if err := ps.Create(ctx, 1, &product); err != nil {
		t.Error("failed to create product", err)
	}

	if _, err := session.Exec(ctx, "UPDATE products SET is_active = true"); err != nil {
		t.Errorf("failed to active products, %s", err)
	}
	products, err := ps.List(ctx, 1, 1)
	if err != nil {
		t.Error("unexpected error in list method, ", err)
	}
	productID := products[0].ID

	pcs := queries.NewPCStore(session)
	initialTags := []string{"1", "2", "3"}
	if err := pcs.SetTags(ctx, productID, initialTags); err != nil {
		t.Fatalf("SetTags failed: %s", err)
	}

	got, err := pcs.List(ctx, productID)
	if err != nil {
		t.Fatalf("List failed: %s", err)
	}
	if len(got) != len(initialTags) {
		t.Fatalf("expected %d tags, got %d", len(initialTags), len(got))
	}

	expected := map[string]bool{"1": true, "2": true, "3": true}
	for _, tag := range got {
		if !expected[tag] {
			t.Errorf("unexpected tag %q found", tag)
		}
	}
	newTags := []string{"4", "5"}
	if err := pcs.SetTags(ctx, productID, newTags); err != nil {
		t.Fatalf("SetTags overwrite failed: %s", err)
	}

	got, err = pcs.List(ctx, productID)
	if err != nil {
		t.Fatalf("List failed after overwrite: %s", err)
	}
	if len(got) != len(newTags) {
		t.Fatalf("expected %d tags after overwrite, got %d", len(newTags), len(got))
	}
	expected = map[string]bool{"4": true, "5": true}
	for _, tag := range got {
		if !expected[tag] {
			t.Errorf("unexpected tag after overwrite: %q", tag)
		}
	}
}
