package queries_test

import (
	"context"
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	"testing"
	"time"
)

func TestCreateListDeleteComment(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	const getIdQuery = `SELECT id FROM comments WHERE username = $1 LIMIT 1`
	session := db.NewSession()
	cs := queries.NewCommentStore(session)

	parentComment := &queries.RelatedComment{nil, "/", true, queries.Comment{Username: "adminUser", Body: "some content"}}
	if err := cs.Create(ctx, parentComment); err != nil {
		t.Errorf("failed to create parent comment, %s", err)
	}
	if err := session.QueryRow(ctx, getIdQuery, parentComment.Username).Scan(&parentComment.ID); err != nil {
		t.Errorf("failed to find parent comment id, %s", err)
	}

	body := "some content"
	invalidCommentID := "bad id"
	tests := []struct {
		name       string
		comment    queries.RelatedComment
		isErrExist bool
	}{
		{
			"adminUserCommand",
			queries.RelatedComment{nil, "/", true, queries.Comment{Username: "adminUser", Body: body}},
			false,
		},
		{
			"subVendorUserCommand",
			queries.RelatedComment{&parentComment.ID, "/", true, queries.Comment{Username: "vendorUser", Body: body}},
			false,
		},
		{
			"costumerUserCommand",
			queries.RelatedComment{nil, "/", true, queries.Comment{Username: "customerUser", Body: body}},
			false,
		},
		{
			"subAdminCommand",
			queries.RelatedComment{&parentComment.ID, "/", true, queries.Comment{Username: "adminUser", Body: body}},
			false,
		},
		{
			"invalidParentAdminCommand",
			queries.RelatedComment{&invalidCommentID, "/", true, queries.Comment{Username: "adminUser", Body: body}},
			true,
		},
	}

	const getChildrenAmountQuery = `SELECT children_amount FROM comments WHERE id = $1::UUID`
	for _, test := range tests {
		if err := cs.Create(ctx, &test.comment); err != nil && !test.isErrExist {
			t.Errorf(`[%s] unexpected error creating comment, %s`, test.name, err)
		}
		if !test.isErrExist && test.comment.Parent != nil {
			var n int32
			if err := session.QueryRow(ctx, getChildrenAmountQuery, *test.comment.Parent).Scan(&n); err != nil {
				t.Errorf(`failed to find children amount from id, %s`, err)
			}
			expected := parentComment.ChildrenAmount + 1
			if n == expected {
				parentComment.ChildrenAmount = n
			} else {
				t.Errorf(`expected "%d" children for parent comment, but got "%d", name="%s"`, expected, n, test.name)
			}
		}
	}

	list, err := cs.List(ctx, nil, "/", 3, 1)
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

	if comment, err := cs.Get(ctx, parentComment.ID); err != nil {
		t.Errorf("failed to get comment, %s", err)
	} else {
		if comment.Body != body && comment.ID != parentComment.ID && comment.Username != parentComment.Username {
			t.Errorf(`unexpected invalid matching, body="%s", id="%s", username="%s"`, comment.Body, comment.ID, comment.Username)
		}
	}

	if err := cs.SetActive(ctx, parentComment.ID, false); err != nil {
		t.Errorf("bad CommentStore.SetActive, %s", err)
	}

	const isCommentsExistsByID = "SELECT EXISTS(SELECT 1 FROM comments WHERE id = $1::UUID OR parent = $1::UUID)"
	if err := cs.Delete(ctx, parentComment.ID); err != nil {
		t.Errorf("failed to delete comment, %s", err)
	} else {
		var isExists bool
		if err := session.QueryRow(ctx, isCommentsExistsByID, parentComment.ID).Scan(&isExists); err != nil {
			t.Fatalf("failed to query existing of comments, %s", err)
		}
		if isExists {
			t.Errorf("failed to delete parent comment and all related comments")
		}
	}
}

func TestFullListComment(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	session := db.NewSession()
	if _, err := session.Exec(ctx, "TRUNCATE comments"); err != nil {
		t.Errorf("error truncate comments, %s", err)
		return
	}

	cs := queries.NewCommentStore(session)

	comments := []queries.RelatedComment{
		{nil, "/", true, queries.Comment{Username: "adminUser", Body: "admin body"}},
		{nil, "/", true, queries.Comment{Username: "vendorUser", Body: "vendor body"}},
		{nil, "/", false, queries.Comment{Username: "customerUser", Body: "customer body"}},
		{nil, "/", true, queries.Comment{Username: "adminUser", Body: "admin body"}},
	}
	for i, comment := range comments {
		if err := cs.Create(ctx, &comment); err != nil {
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
		t.Run(test.name, func(tt *testing.T) {
			flist, err := cs.FullList(ctx, test.username, test.pagination, test.page)
			if err != nil {
				if test.expectedCount == 0 {
					return
				}
				t.Fatalf("unexpected error: %s", err)
			}
			if got := len(flist); got != test.expectedCount {
				t.Errorf(`expected "%d" result, but got "%d"`, test.expectedCount, got)
			}
		})
	}
}
