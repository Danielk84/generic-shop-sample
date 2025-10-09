package queries

import (
	"context"
	"generic-shop-sample/db"
	"time"

	"github.com/jackc/pgx/v5"
)

type CommentRequest struct {
	Username string `json:"username" binding:"required,username"`
	Parent   string `json:"parent" binding:"required,uuid"`
	Referrer string `json:"referrer" binding:"required"`
	Body     string `json:"body"`
}

type CommentResponse struct {
	ID             string    `json:"id"`
	Username       string    `json:"username"`
	PubDate        time.Time `json:"pub_date"`
	ChildrenAmount int32     `json:"children_amount"`
	Body           string    `json:"body"`
}

type RelatedCommentResponse struct {
	CommentResponse
	Parent   string `json:"parent"`
	Referrer string `json:"referrer"`
	IsActive bool   `json:"is_active"`
}

type CommentRepository struct {
	session db.Session
}

type CommentStore interface {
	Create(ctx context.Context, comment *CommentRequest) error
	Get(ctx context.Context, id string) (*RelatedCommentResponse, error)
	List(ctx context.Context, parent string, referrer string, pagination, page int) ([]CommentResponse, error)
	FullList(ctx context.Context, username string, pagination, page int) ([]RelatedCommentResponse, error)
	Delete(ctx context.Context, id string) error
	SetActive(ctx context.Context, id string, isActive bool) error
}

func NewCommentStore(session db.Session) CommentStore {
	return &CommentRepository{session}
}

func (cr *CommentRepository) Create(ctx context.Context, comment *CommentRequest) error {
	const createCommentQuery = `INSERT INTO comments(username, parent, children_amount, referrer, body)
		VALUES (@Username, NULLIF(@Parent, '')::UUID, 0, @Referrer, @Body)`
	const upadteChildrenCount = `UPDATE comments SET children_amount = children_amount + 1 WHERE id = $1::UUID`

	args := pgx.NamedArgs{
		"Username": comment.Username,
		"Parent":   comment.Parent,
		"Referrer": comment.Referrer,
		"Body":     comment.Body,
	}
	return pgx.BeginFunc(ctx, cr.session, func(tx pgx.Tx) error {
		cTag, err := tx.Exec(ctx, createCommentQuery, args)
		if err != nil {
			return err
		}
		if cTag.RowsAffected() != 1 {
			return pgx.ErrNoRows
		}

		if comment.Parent != "" {
			if _, err := tx.Exec(ctx, upadteChildrenCount, comment.Parent); err != nil {
				return err
			}
		}
		return nil
	})
}

func (cr *CommentRepository) Get(ctx context.Context, id string) (*RelatedCommentResponse, error) {
	const q = `SELECT id, username, pub_date, COALESCE(parent::TEXT, '') AS parent, children_amount, referrer, body, is_active FROM comments
		WHERE id = $1::UUID`
	return get[RelatedCommentResponse](ctx, cr.session, q, id)
}

func (cr *CommentRepository) List(ctx context.Context, parent string, referrer string, pagination, page int) ([]CommentResponse, error) {
	const q = `SELECT id, username, pub_date, children_amount, body FROM comments
		WHERE COALESCE(parent::TEXT, '') = @Parent
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
	return list[CommentResponse](ctx, cr.session, q, args)
}

func (cr *CommentRepository) FullList(ctx context.Context, username string, pagination, page int) ([]RelatedCommentResponse, error) {
	const baseQuery = "SELECT id, username, pub_date, COALESCE(parent::TEXT, '') AS parent, children_amount, referrer, body, is_active FROM comments"
	const limitOffset = ` LIMIT @Pagination OFFSET @Offset`
	args := pgx.NamedArgs{
		"Pagination": pagination,
		"Offset":     (page - 1) * pagination,
	}

	if username == "" {
		q := baseQuery + " ORDER BY pub_date DESC, is_active" + limitOffset
		return list[RelatedCommentResponse](ctx, cr.session, q, args)
	}
	args["Username"] = username
	q := baseQuery + " WHERE username = @Username ORDER BY pub_date DESC" + limitOffset
	return list[RelatedCommentResponse](ctx, cr.session, q, args)
}

func (cr *CommentRepository) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM comments WHERE id = $1::UUID OR parent = $1::UUID`
	_, err := cr.session.Exec(ctx, q, id)
	return err
}

func (cr *CommentRepository) SetActive(ctx context.Context, id string, isActive bool) error {
	const q = `UPDATE comments SET is_active = $1 WHERE id = $2::UUID`
	return execOne(ctx, cr.session, q, isActive, id)
}
