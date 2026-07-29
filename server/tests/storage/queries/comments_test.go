package queries_test

import (
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/queries"
	tu "generic-shop-sample/tests/internal/testutils"
	"testing"
	"time"
)

func TestCommentStore(t *testing.T) {
	ctx := t.Context()
	config := tu.ConfigTestSetup()
	session := tu.DBTestSetup(ctx, config).GetSession()
	log := logger.GetLogger()

	if _, err := session.Exec(ctx, "TRUNCATE user_s.comments, product_s.products CASCADE"); err != nil {
		t.Errorf("error truncate comments and products, %s", err)
		return
	}

	productStore := queries.NewProductStore(session, log)
	if err := productStore.Create(ctx, queries.CreateProductRequest{
		Name:        "new product for comments",
		Description: "lalala",
		CommonDetail: queries.ProductProperty{
			"info": "some info",
		},
	}); err != nil {
		t.Errorf("failed to create product, %s", err)
	}
	products, err := productStore.AdminList(ctx, 20, 1)
	if err != nil {
		t.Errorf("failed to get products full list, %s", err)
	}
	product := products[0]

	username := "adminUser"
	store := queries.NewCommentStore(session, log)
	parentComment := &queries.CommentRequest{
		Parent:   "",
		Referrer: product.ID,
		Body:     "some content",
	}
	if err := store.Create(ctx, username, parentComment); err != nil {
		t.Errorf("failed to create parent comment, %s", err)
	}
	var parentCommentID string
	const getIdQuery = `SELECT id FROM user_s.comments WHERE username = $1 LIMIT 1`
	if err := session.QueryRow(ctx, getIdQuery, username).Scan(&parentCommentID); err != nil {
		t.Errorf("failed to find parent comment id, %s", err)
	}

	body := "some content"
	invalidCommentID := "bad id"
	tests := []struct {
		name       string
		comment    queries.CommentRequest
		isErrExist bool
	}{
		{
			"adminUserCommand",
			queries.CommentRequest{Parent: "", Referrer: product.ID, Body: body},
			false,
		},
		{
			"subVendorUserCommand",
			queries.CommentRequest{Parent: parentCommentID, Referrer: product.ID, Body: body},
			false,
		},
		{
			"costumerUserCommand",
			queries.CommentRequest{Parent: "", Referrer: product.ID, Body: body},
			false,
		},
		{
			"subAdminCommand",
			queries.CommentRequest{Parent: parentCommentID, Referrer: product.ID, Body: body},
			false,
		},
		{
			"invalidParentAdminCommand",
			queries.CommentRequest{Parent: invalidCommentID, Referrer: product.ID, Body: body},
			true,
		},
	}

	var childrenAmount int32
	const getChildrenAmountQuery = `SELECT children_amount FROM user_s.comments WHERE id = $1::UUID`
	for i, test := range tests {
		u := username
		if (i+1)%3 == 0 {
			u = "customerUser"
		}
		if err := store.Create(ctx, u, &test.comment); err != nil && !test.isErrExist {
			t.Errorf(`[%s] unexpected error creating comment, %s`, test.name, err)
		}
		if !test.isErrExist && test.comment.Parent != "" {
			var n int32
			if err := session.QueryRow(ctx, getChildrenAmountQuery, test.comment.Parent).Scan(&n); err != nil {
				t.Errorf(`failed to find children amount from id, %s`, err)
			}
			expected := childrenAmount + 1
			if n == expected {
				childrenAmount = n
			} else {
				t.Errorf(`expected "%d" children for parent comment, but got "%d", name="%s"`, expected, n, test.name)
			}
		}
	}

	if _, err := session.Exec(ctx, "UPDATE user_s.comments SET is_active = true"); err != nil {
		t.Errorf("failed to adtivate comments, %s", err)
		return
	}

	list, err := store.List(ctx, "", product.ID, 3, 1)
	if err != nil {
		t.Errorf("failed to query comments list, %s", err)
		return
	}
	listLen := len(list)
	if listLen != 3 {
		t.Errorf(`expected length of comments list be 2, but got "%d"`, listLen)
	}

	expectedUsers := map[string]bool{
		"adminUser":    true,
		"customerUser": true,
	}
	for _, item := range list {
		isExpectedUser := expectedUsers[item.Username]
		if !isExpectedUser && item.Body != body {
			t.Errorf(`Invalid list's item, isExpectedUser="%v", body="%s"`, isExpectedUser, item.Body)
		}
	}

	if comment, err := store.Get(ctx, parentCommentID); err != nil {
		t.Errorf("failed to get comment, %s", err)
	} else {
		if comment.Body != body && comment.ID != parentCommentID {
			t.Errorf(`unexpected invalid matching, body="%s", id="%s", username="%s"`, comment.Body, comment.ID, comment.Username)
		}
	}

	if err := store.SetActive(ctx, parentCommentID, false); err != nil {
		t.Errorf("bad CommentStore.SetActive, %s", err)
	}

	const isCommentsExistsByID = "SELECT EXISTS(SELECT 1 FROM user_s.comments WHERE id = $1::UUID OR parent = $1::UUID)"
	if err := store.Delete(ctx, parentCommentID); err != nil {
		t.Errorf("failed to delete comment, %s", err)
	} else {
		var isExists bool
		if err := session.QueryRow(ctx, isCommentsExistsByID, parentCommentID).Scan(&isExists); err != nil {
			t.Fatalf("failed to query existing of comments, %s", err)
		}
		if isExists {
			t.Errorf("failed to delete parent comment and all related comments")
		}
	}
}

func TestFullListComments(t *testing.T) {
	ctx := t.Context()
	config := tu.ConfigTestSetup()
	session := tu.DBTestSetup(ctx, config).GetSession()
	if _, err := session.Exec(ctx, "TRUNCATE user_s.comments"); err != nil {
		t.Errorf("error truncate comments, %s", err)
		return
	}
	log := logger.GetLogger()

	productStore := queries.NewProductStore(session, log)
	products, err := productStore.AdminList(ctx, 20, 1)
	if err != nil {
		t.Errorf("failed to get products full list, %s", err)
	}
	product := products[0]

	store := queries.NewCommentStore(session, log)
	comments := []struct {
		username string
		queries.CommentRequest
	}{
		{"adminUser", queries.CommentRequest{Parent: "", Referrer: product.ID, Body: "admin body"}},
		{"vendorUser", queries.CommentRequest{Parent: "", Referrer: product.ID, Body: "vendor body"}},
		{"customerUser", queries.CommentRequest{Parent: "", Referrer: product.ID, Body: "customer body"}},
		{"adminUser", queries.CommentRequest{Parent: "", Referrer: product.ID, Body: "admin body"}},
	}

	for i, comment := range comments {
		if err := store.Create(ctx, comment.username, &comment.CommentRequest); err != nil {
			t.Errorf(`failed to create comment "%d", %s`, i, err)
		}
	}

	tests := []struct {
		name          string
		username      string
		pagination    int
		page          int
		expectedCount int
	}{
		{
			"list all comments",
			"",
			2,
			2,
			2,
		},
		{
			"list by username",
			"vendorUser",
			4,
			1,
			1,
		},
		{
			"no match username",
			"ghostUser",
			10,
			1,
			0,
		},
		{
			"test pagination",
			"adminUser",
			10,
			3,
			0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(st *testing.T) {
			flist, err := store.FullList(ctx, test.username, test.pagination, test.page)
			if err != nil {
				if test.expectedCount == 0 {
					return
				}
				st.Fatalf("unexpected error: %s", err)
			}
			if got := len(flist); got != test.expectedCount {
				st.Errorf(`expected "%d" result, but got "%d"`, test.expectedCount, got)
			}
		})
		time.Sleep(1 * time.Second)
	}
}
