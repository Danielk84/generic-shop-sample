package queries_test

import (
	"generic-shop-sample/db/database"
	"generic-shop-sample/db/queries"
	"testing"
	"time"
)

func TestCommentStore(t *testing.T) {
	ctx := t.Context()

	session := database.GetSession()

	if _, err := session.Exec(ctx, "TRUNCATE comments, products CASCADE"); err != nil {
		t.Errorf("error truncate comments and products, %s", err)
		return
	}

	ps := queries.NewProductStore(session)
	if err := ps.Create(ctx, 1, &queries.CreateProductRequest{
		Name:        "new product for comments",
		Price:       1001,
		Description: "lalala",
		Details:     "{}",
		IsAvailable: true,
	}); err != nil {
		t.Errorf("failed to create product, %s", err)
	}
	products, err := ps.FullList(ctx, 0, 20, 1)
	if err != nil {
		t.Errorf("failed to get products full list, %s", err)
	}
	product := products[0]

	username := "adminUser"
	cs := queries.NewCommentStore(session)
	parentComment := &queries.CommentRequest{"", product.ID, "some content"}
	if err := cs.Create(ctx, username, parentComment); err != nil {
		t.Errorf("failed to create parent comment, %s", err)
	}
	var parentCommentID string
	const getIdQuery = `SELECT id FROM comments WHERE username = $1 LIMIT 1`
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
			queries.CommentRequest{"", product.ID, body},
			false,
		},
		{
			"subVendorUserCommand",
			queries.CommentRequest{parentCommentID, product.ID, body},
			false,
		},
		{
			"costumerUserCommand",
			queries.CommentRequest{"", product.ID, body},
			false,
		},
		{
			"subAdminCommand",
			queries.CommentRequest{parentCommentID, product.ID, body},
			false,
		},
		{
			"invalidParentAdminCommand",
			queries.CommentRequest{invalidCommentID, product.ID, body},
			true,
		},
	}

	var childrenAmount int32
	const getChildrenAmountQuery = `SELECT children_amount FROM comments WHERE id = $1::UUID`
	for i, test := range tests {
		u := username
		if (i+1)%3 == 0 {
			u = "customerUser"
		}
		if err := cs.Create(ctx, u, &test.comment); err != nil && !test.isErrExist {
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

	if _, err := session.Exec(ctx, "UPDATE comments SET is_active = true"); err != nil {
		t.Errorf("failed to adtivate comments, %s", err)
		return
	}

	list, err := cs.List(ctx, "", product.ID, 3, 1)
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

	if comment, err := cs.Get(ctx, parentCommentID); err != nil {
		t.Errorf("failed to get comment, %s", err)
	} else {
		if comment.Body != body && comment.ID != parentCommentID {
			t.Errorf(`unexpected invalid matching, body="%s", id="%s", username="%s"`, comment.Body, comment.ID, comment.Username)
		}
	}

	if err := cs.SetActive(ctx, parentCommentID, false); err != nil {
		t.Errorf("bad CommentStore.SetActive, %s", err)
	}

	const isCommentsExistsByID = "SELECT EXISTS(SELECT 1 FROM comments WHERE id = $1::UUID OR parent = $1::UUID)"
	if err := cs.Delete(ctx, parentCommentID); err != nil {
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

	session := database.GetSession()
	if _, err := session.Exec(ctx, "TRUNCATE comments"); err != nil {
		t.Errorf("error truncate comments, %s", err)
		return
	}

	ps := queries.NewProductStore(session)
	products, err := ps.FullList(ctx, 0, 20, 1)
	if err != nil {
		t.Errorf("failed to get products full list, %s", err)
	}
	product := products[0]

	cs := queries.NewCommentStore(session)
	comments := []struct {
		username string
		queries.CommentRequest
	}{
		{"adminUser", queries.CommentRequest{"", product.ID, "admin body"}},
		{"vendorUser", queries.CommentRequest{"", product.ID, "vendor body"}},
		{"customerUser", queries.CommentRequest{"", product.ID, "customer body"}},
		{"adminUser", queries.CommentRequest{"", product.ID, "admin body"}},
	}

	for i, comment := range comments {
		if err := cs.Create(ctx, comment.username, &comment.CommentRequest); err != nil {
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
			flist, err := cs.FullList(ctx, test.username, test.pagination, test.page)
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
