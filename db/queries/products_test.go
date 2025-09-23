package queries_test

import (
	"context"
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	"testing"
	"time"
)

func TestBasicProductRepositoryMethod(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	session := db.NewSession()
	ps := queries.NewProductStore(session)

	description := "Test product description"
	product := queries.OwnedProduct{
		UserID: 1,
		Product: queries.Product{
			IsAvailable: true,
			IsActive:    true,
			Description: &description,
			Details:     map[string]string{"color": "red", "size": "M"},
			ProductSummary: queries.ProductSummary{
				Name:  "TestProduct",
				Price: 1999,
			},
		},
	}

	if err := ps.Create(ctx, &product); err != nil {
		t.Fatalf("failed to create product, %s", err)
	}

	list, err := ps.List(ctx, 10, 1)
	if err != nil {
		t.Fatalf("failed to list products, %s", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 product in list, but got %d", len(list))
	}
	gotSummary := list[0]
	if gotSummary.Name != product.Name || gotSummary.Price != product.Price {
		t.Errorf("list mismatch: got %+v, expected name=%q, price=%d", gotSummary, product.Name, product.Price)
	}

	product.ID = gotSummary.ID
	got, err := ps.Get(ctx, product.ID)
	if err != nil {
		t.Fatalf("failed to get product, %s", err)
	}
	if got.Name != product.Name || *got.Description != *product.Description {
		t.Errorf("get mismatch: got %+v, expected name=%q, description=%q",
			got, product.Name, *product.Description)
	}

	newDescription := "Updated description"
	product.Description = &newDescription
	product.IsAvailable = false
	if err := ps.Update(ctx, &product.Product); err != nil {
		t.Fatalf("failed to update product, %s", err)
	}

	updated, err := ps.Get(ctx, product.ID)
	if err != nil {
		t.Fatalf("failed to get updated product, %s", err)
	}
	if *updated.Description != newDescription || updated.IsAvailable != false {
		t.Errorf("update mismatch: got %+v, expected description=%q, isAvailable=false",
			updated, newDescription)
	}

	if err := ps.SetActive(ctx, product.ID, false); err != nil {
		t.Fatalf("failed to set active: %s", err)
	}
	if err := ps.SetAvailable(ctx, product.ID, true); err != nil {
		t.Fatalf("failed to set available, %s", err)
	}

	check, err := ps.Get(ctx, product.ID)
	if err != nil {
		t.Fatalf("failed to get product after state changes: %s", err)
	}
	if check.IsActive != false || check.IsAvailable != true {
		t.Errorf("expected isActive=false, isAvailable=true; got %+v", check)
	}

	if err := ps.Delete(ctx, product.ID); err != nil {
		t.Fatalf("failed to delete product, %s", err)
	}

	_, err = ps.Get(ctx, product.ID)
	if err == nil {
		t.Errorf("expected error when getting deleted product, got none")
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
	products := []queries.OwnedProduct{
		{1, queries.Product{true, true, &description, map[string]string{"info": info}, queries.ProductSummary{Name: "item 1", Price: 10}}},
		{1, queries.Product{true, true, &description, map[string]string{"info": info}, queries.ProductSummary{Name: "item 2", Price: 10}}},
		{1, queries.Product{true, true, &description, map[string]string{"info": info}, queries.ProductSummary{Name: "item 3", Price: 10}}},
		{2, queries.Product{true, true, &description, map[string]string{"info": info}, queries.ProductSummary{Name: "item 4", Price: 10}}},
		{3, queries.Product{true, true, &description, map[string]string{"info": info}, queries.ProductSummary{Name: "item 5", Price: 10}}},
	}
	for i, product := range products {
		if err := ps.Create(ctx, &product); err != nil {
			t.Errorf(`failed to create product "%d", %s`, i, err)
		}
	}

	tests := []struct {
		name          string
		id            int
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
