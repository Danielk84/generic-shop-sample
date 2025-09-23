package queries_test

import (
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	"testing"
)

func TestCreateListDeleteComment(t *testing.T) {
	const getIdQuery = `SELECT id FROM comments WHERE username = $1 LIMIT 1`
	session := db.NewSession()
	cs := queries.NewCommentStore(session)

	parentComment := &queries.RelatedComment{nil, "/", true, queries.Comment{Username: "adminUser", Body: "some content"}}
	if err := cs.Create(t.Context(), parentComment); err != nil {
		t.Errorf("failed to create parent comment, %s", err)
	}
	if err := session.QueryRow(t.Context(), getIdQuery, parentComment.Username).Scan(&parentComment.ID); err != nil {
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
		if err := cs.Create(t.Context(), &test.comment); err != nil && !test.isErrExist {
			t.Errorf(`[%s] unexpected error creating comment, %s`, test.name, err)
		}
		if !test.isErrExist && test.comment.Parent != nil {
			var n int32
			if err := session.QueryRow(t.Context(), getChildrenAmountQuery, *test.comment.Parent).Scan(&n); err != nil {
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

	list, err := cs.List(t.Context(), nil, "/", 3, 1)
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

	if err := cs.Delete(t.Context(), parentComment.ID); err != nil {
		t.Errorf("failed to delete comment, %s", err)
	}
}
