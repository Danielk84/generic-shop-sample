package queries_test

import (
	"fmt"
	"generic-shop-sample/db/database"
	"generic-shop-sample/db/queries"
	"generic-shop-sample/internal"
	"testing"
)

func TestProductStore(t *testing.T) {
	ctx := t.Context()

	session := database.GetSession()
	if _, err := session.Exec(ctx, "TRUNCATE products RESTART IDENTITY CASCADE"); err != nil {
		t.Errorf("failed to truncate products, %s", err)
	}
	ps := queries.NewProductStore(session)

	product := queries.CreateProductRequest{
		Name:        "TestProduct",
		Price:       1999,
		Description: "Test product description",
		Details:     `{"color": "red", "size": "M"}`,
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

	if err := ps.IncrBy(ctx, gotSummary.ID, 1, 10); err != nil {
		t.Errorf("failed to incr by, %s", err)
	}
	if err := ps.DecrBy(ctx, gotSummary.ID, 1, 5); err != nil {
		t.Errorf("failed to decr by, %s", err)
	}

	got, err := ps.Get(ctx, gotSummary.ID)
	if err != nil {
		t.Fatalf("failed to get product, %s", err)
	}
	if got.Name != product.Name || got.Description != product.Description || got.AvailableQuantity != 5 {
		t.Errorf(`get mismatch: got "%+v", expected name="%q", description="%q", available_quantity="%d"`,
			got, product.Name, product.Description, 5)
	}

	uProduct := queries.UpdateProductRequest{
		got.ID,
		queries.CreateProductRequest{
			Name:        "TestProduct",
			Price:       1999,
			Description: "New Test product description",
			Details:     `{"color": "red", "size": "M"}`,
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
	ctx := t.Context()

	session := database.GetSession()
	if _, err := session.Exec(ctx, "TRUNCATE products RESTART IDENTITY CASCADE"); err != nil {
		t.Errorf("failed to truncate products")
	}

	ps := queries.NewProductStore(session)

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
			items, err := ps.FullList(ctx, test.id, test.pagination, test.page)
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
	ps := queries.NewProductStore(session)
	pis := queries.NewProductImagesStore(session)

	if _, err := session.Exec(ctx, "TRUNCATE products RESTART IDENTITY CASCADE"); err != nil {
		t.Errorf("failed to truncate products, %s", err)
	}

	if err := ps.Create(ctx, 1, &queries.CreateProductRequest{
		Name:        "new product",
		Price:       10,
		IsAvailable: true,
	}); err != nil {
		t.Errorf("failed to create product, %s", err)
	}
	if _, err := session.Exec(ctx, "UPDATE products SET is_active = true"); err != nil {
		t.Errorf("failed to activate products, %s", err)
	}
	var product queries.ProductSummaryResponse
	if products, err := ps.List(ctx, 10, 1); err == nil {
		product = products[0]
	} else {
		t.Errorf("failed to get products list, %s", err)
	}

	config := internal.GetConfig()
	tests := make([]string, config.Opt.MaxProductImagesAmount)
	for i := range tests {
		tests[i] = fmt.Sprintf("path/to/img%d.img", i)
	}

	for i, imgPath := range tests {
		if err := pis.Create(ctx, product.ID, imgPath); err != nil {
			if i == config.Opt.MaxProductImagesAmount {
				break
			}
			t.Errorf("failed to create product image row, %s", err)
		}
		if i == config.Opt.MaxProductImagesAmount {
			t.Errorf("failed to return error for creating more than max value")
		}
	}
	productImages, err := pis.List(ctx, product.ID)
	if err != nil {
		t.Errorf("failed to get product images list, %s", err)
	}
	if got := len(productImages); got != len(tests) {
		t.Errorf(`expected list return "%d" product images, but got "%d"`, len(tests), got)
	}
	if imgPath, err := pis.Delete(ctx, productImages[0].ID); err == nil {
		if imgPath != productImages[0].ImgPath {
			t.Error("failed to match imgPath")
		}
	} else {
		t.Errorf("failed to delete product image, %s", err)
	}
}
