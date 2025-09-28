package queries_test

import (
	"context"
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	"testing"
	"time"
)

func TestBasicProductStoreMethods(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	session := db.NewSession()
	ps := queries.NewProductStore(session)

	product := queries.CreateProductRequest{
		Name:        "TestProduct",
		Price:       1999,
		Description: "Test product description",
		Details:     map[string]string{"color": "red", "size": "M"},
		IsAvailable: true,
	}

	if err := ps.Create(ctx, 1, &product); err != nil {
		t.Fatalf("failed to create product, %s", err)
	}

	if _, err := session.Exec(ctx, "UPDATE products SET is_active = true"); err != nil {
		t.Fatalf("failed to activate products")
	}

	list, err := ps.List(ctx, 10, 1)
	if err != nil {
		t.Fatalf("failed to get peoducts list, %s", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 product in list, but got %d", len(list))
	}
	gotSummary := list[0]
	if gotSummary.Name != product.Name || gotSummary.Price != product.Price {
		t.Errorf("list mismatch: got %+v, expected name=%q, price=%d", gotSummary, product.Name, product.Price)
	}

	got, err := ps.Get(ctx, gotSummary.ID)
	if err != nil {
		t.Fatalf("failed to get product, %s", err)
	}
	if got.Name != product.Name || got.Description != product.Description {
		t.Errorf("get mismatch: got %+v, expected name=%q, description=%q",
			got, product.Name, product.Description)
	}

	uProduct := queries.UpdateProductRequest{
		got.ID,
		queries.CreateProductRequest{
			Name:        "TestProduct",
			Price:       1999,
			Description: "New Test product description",
			Details:     map[string]string{"color": "red", "size": "M"},
			IsAvailable: false,
		},
	}
	if err := ps.Update(ctx, 1, &uProduct); err != nil {
		t.Fatalf("failed to update product, %s", err)
	}

	updated, err := ps.Get(ctx, got.ID)
	if err != nil {
		t.Fatalf("failed to get updated product, %s", err)
	}
	if updated.Description != uProduct.Description || updated.IsAvailable != false {
		t.Errorf("update mismatch: got %+v, expected description=%q, isAvailable=false",
			updated, uProduct.Description)
	}

	if err := ps.SetActive(ctx, got.ID, false); err != nil {
		t.Fatalf("failed to set active: %s", err)
	}
	if err := ps.SetAvailable(ctx, got.ID, true); err != nil {
		t.Fatalf("failed to set available, %s", err)
	}

	check, err := ps.Get(ctx, got.ID)
	if err != nil {
		t.Fatalf("failed to get product after state changes: %s", err)
	}
	if check.IsActive != false || check.IsAvailable != true {
		t.Errorf("expected isActive=false, isAvailable=true; got %+v", check)
	}

	if err := ps.Delete(ctx, got.ID, got.UserID); err != nil {
		t.Fatalf("failed to delete product, %s", err)
	}

	_, err = ps.Get(ctx, got.ID)
	if err == nil {
		t.Errorf("expected error when getting deleted product, got nil")
	}
}

func TestFullListProducts(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	session := db.NewSession()
	if _, err := session.Exec(ctx, "TRUNCATE products RESTART IDENTITY CASCADE"); err != nil {
		t.Errorf("failed to truncate products")
	}

	ps := queries.NewProductStore(session)

	description := "some descriptions"
	info := "some info"
	products := []struct {
		userID  int32
		product queries.CreateProductRequest
	}{
		{1, queries.CreateProductRequest{Name: "item 1", Price: 10, Description: description, Details: map[string]string{"info": info}, IsAvailable: true}},
		{1, queries.CreateProductRequest{Name: "item 2", Price: 10, Description: description, Details: map[string]string{"info": info}, IsAvailable: true}},
		{1, queries.CreateProductRequest{Name: "item 3", Price: 10, Description: description, Details: map[string]string{"info": info}, IsAvailable: true}},
		{2, queries.CreateProductRequest{Name: "item 4", Price: 10, Description: description, Details: map[string]string{"info": info}, IsAvailable: true}},
		{3, queries.CreateProductRequest{Name: "item 5", Price: 10, Description: description, Details: map[string]string{"info": info}, IsAvailable: true}},
	}
	for i, product := range products {
		if err := ps.Create(ctx, product.userID, &product.product); err != nil {
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
			flist, err := ps.FullList(ctx, test.id, test.pagination, test.page)
			if err != nil {
				if test.expectedCount == 0 {
					return
				}
				st.Errorf("unexpected error, %s", err)
				return
			}
			if got := len(flist); got != test.expectedCount {
				t.Errorf(`expected "%d" products, but got "%d"`, test.expectedCount, got)
			}
		})
	}
}

func TestBasicCategoryStoreMethods(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

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

func TestSetTagsListPCStore(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

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
	product := queries.CreateProductRequest{Name: "item 1", Price: 10, Description: "some description", Details: map[string]string{"info": "some info"}, IsAvailable: true}
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
