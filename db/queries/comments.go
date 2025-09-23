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
	IsActive bool
	Comment
}

type CommentRepository struct {
	session db.Session
}

type CommentStore interface {
	Create(context.Context, *RelatedComment) error
	List(context.Context, *string, string, int, int) ([]Comment, error)
	Delete(context.Context, string) error
}

func NewCommentStore(session db.Session) CommentStore {
	return &CommentRepository{session}
}

func (cr *CommentRepository) Create(ctx context.Context, comment *RelatedComment) error {
	const createCommentQuery = `INSERT INTO comments(username, parent, children_amount, referrer, body, is_active)
		VALUES (@Username, @Parent::UUID, 0, @Referrer, @Body, @IsActive)`
	const upadteChildrenCount = `UPDATE comments SET children_amount = children_amount + 1 WHERE id = $1::UUID`

	args := pgx.NamedArgs{
		"Username": comment.Username,
		"Parent":   comment.Parent,
		"Referrer": comment.Referrer,
		"Body":     comment.Body,
		"IsActive": comment.IsActive,
	}
	return pgx.BeginFunc(ctx, cr.session, func(tx pgx.Tx) error {
		cTag, err := tx.Exec(ctx, createCommentQuery, args)
		if err != nil {
			return err
		}
		if cTag.RowsAffected() != 1 {
			return pgx.ErrNoRows
		}

		if comment.Parent != nil {
			if _, err := tx.Exec(ctx, upadteChildrenCount, *comment.Parent); err != nil {
				return err
			}
		}
		return nil
	})
}

func (cr *CommentRepository) List(ctx context.Context, parent *string, referrer string, pagination, page int) ([]Comment, error) {
	const q = `SELECT id, username, pub_date, children_amount, body FROM comments
		WHERE (parent = @Parent::UUID OR (@Parent::UUID IS NULL AND parent IS NULL))
			AND referrer = @Referrer
			AND is_active = true
		ORDER BY pub_date DESC
		LIMIT @Pagniation
		OFFSET @Offset`
	args := pgx.NamedArgs{
		"Parent":     parent,
		"Referrer":   referrer,
		"Pagination": pagination,
		"Offset":     (page - 1) * pagination,
	}
	return list[Comment](ctx, cr.session, q, args)
}

func (cr *CommentRepository) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM comments WHERE id = $1::UUID`
	return execOne(ctx, cr.session, q, id)
}
