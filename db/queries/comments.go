package queries

import (
	"context"
	"generic-shop-sample/db"
	"time"

	"github.com/jackc/pgx/v5"
)

type Comment struct {
	ID             string
	Username       string
	PubDate        time.Time
	ChildrenAmount int32
	Body           string
}

type RelatedComment struct {
	Parent   *string
	Referrer string
	Comment
}

type CommentRepository struct {
	session db.Session
}

type CommentStore interface {
	List(context.Context, *string, string, bool) (*[]Comment, error)
	Create(context.Context, *RelatedComment) error
	Delete(context.Context, string) error
}

func NewCommentStore(session db.Session) CommentStore {
	return &CommentRepository{session}
}

func (cr *CommentRepository) List(ctx context.Context, parent *string, referrer string, isActive bool) (*[]Comment, error) {
	const q = `SELECT id, username, pub_date, children_amount, body FROM comments
		WHERE parent = $1::UUID, referrer = $2, is_active = $3`
	return list[Comment](ctx, cr.session, q, parent, referrer, isActive)
}

func (cr *CommentRepository) Create(ctx context.Context, comment *RelatedComment) error {
	const q = `INSERT INTO comments(username, parent, children_amount, referrer, body)
		VALUES (@Username, @Parent::UUID, @ChildrenAmount, @Referrer, @Body)`
	args := pgx.NamedArgs{
		"Username":       comment.Username,
		"Parent":         comment.Parent,
		"ChildrenAmount": comment.ChildrenAmount,
		"Referrer":       comment.Referrer,
		"Body":           comment.Body,
	}
	return execOne(ctx, cr.session, q, args)
}

func (cr *CommentRepository) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM comments WHERE id = $1::UUID`
	return execOne(ctx, cr.session, q, id)
}
