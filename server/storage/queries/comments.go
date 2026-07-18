package queries

import (
	"context"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/database"
	"time"

	"github.com/jackc/pgx/v5"
)

type CommentRequest struct {
	Parent   string `json:"parent" binding:"omitempty,uuid"`
	Referrer string `json:"referrer" binding:"required,uuid"`
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
	session database.Session
	log     logger.Logger
}

type CommentStore interface {
	Create(ctx context.Context, username string, comment *CommentRequest) error
	Get(ctx context.Context, id string) (RelatedCommentResponse, error)
	List(ctx context.Context, parent string, referrer string, pagination, page int) ([]CommentResponse, error)
	FullList(ctx context.Context, username string, pagination, page int) ([]RelatedCommentResponse, error)
	Delete(ctx context.Context, id string) error
	SetActive(ctx context.Context, id string, isActive bool) error
}

func NewCommentStore(session database.Session, log logger.Logger) CommentStore {
	return &CommentRepository{session, log}
}

func (c *CommentRepository) Create(ctx context.Context, username string, comment *CommentRequest) (err error) {
	const createCommentQuery = `INSERT INTO user_s.comments(username, parent, children_amount, referrer, body)
		VALUES (@Username, NULLIF(@Parent, '')::UUID, 0, @Referrer, @Body)`
	const upadteChildrenCount = `UPDATE user_s.comments
		SET children_amount = children_amount + 1
		WHERE id = $1::UUID`

	args := pgx.NamedArgs{
		"Username": username,
		"Parent":   comment.Parent,
		"Referrer": comment.Referrer,
		"Body":     comment.Body,
	}
	err = pgx.BeginFunc(ctx, c.session, func(tx pgx.Tx) error {
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
	if err != nil {
		c.log.Debug("CommentRepository.Create", "error", err)
	}
	return
}

func (c *CommentRepository) Get(ctx context.Context, id string) (item RelatedCommentResponse, err error) {
	const q = `SELECT
			id, COALESCE(username, 'deleted') AS username, pub_date,
			COALESCE(parent::TEXT, '') AS parent, children_amount, referrer,
			body, is_active
		FROM user_s.comments
		WHERE id = $1::UUID`
	item, err = get[RelatedCommentResponse](ctx, c.session, q, id)
	if err != nil {
		c.log.Debug("CommentRepository.Get", "error", err)
	}
	return
}

func (c *CommentRepository) List(ctx context.Context, parent string, referrer string, pagination, page int) (items []CommentResponse, err error) {
	const q = `SELECT
			id, COALESCE(username, 'deleted') AS username, pub_date, children_amount, body
		FROM user_s.comments
		WHERE COALESCE(parent::TEXT, '') = @Parent
			AND referrer = @Referrer
			AND is_active = true
		ORDER BY pub_date DESC
		LIMIT @Limit
		OFFSET @Offset`
	args := pgx.NamedArgs{
		"Parent":   parent,
		"Referrer": referrer,
		"Limit":    pagination,
		"Offset":   getOffsetFromPageNum(pagination, page),
	}
	items, err = list[CommentResponse](ctx, c.session, q, args)
	if err != nil {
		c.log.Debug("CommentRepository.List", "error", err)
	}
	return
}

func (c *CommentRepository) FullList(ctx context.Context, username string, pagination, page int) (items []RelatedCommentResponse, err error) {
	const baseQuery = `SELECT
			id, COALESCE(username, 'deleted') AS username, pub_date,
			COALESCE(parent::TEXT, '') AS parent, children_amount, referrer,
			body, is_active
		FROM user_s.comments`
	const limitOffset = ` LIMIT @Limit OFFSET @Offset`
	args := pgx.NamedArgs{
		"Limit":  pagination,
		"Offset": getOffsetFromPageNum(pagination, page),
	}

	var q string
	if username == "" {
		q = baseQuery + " ORDER BY pub_date DESC, is_active" + limitOffset
	} else {
		args["Username"] = username
		q = baseQuery + " WHERE username = NULLIF(@Username, 'deleted') ORDER BY pub_date DESC" + limitOffset
	}
	items, err = list[RelatedCommentResponse](ctx, c.session, q, args)
	if err != nil {
		c.log.Debug("CommentRepository.FullList", "error", err)
	}
	return
}

func (c *CommentRepository) Delete(ctx context.Context, id string) (err error) {
	const q = `DELETE FROM user_s.comments WHERE id = $1::UUID OR parent = $1::UUID`
	if _, err = c.session.Exec(ctx, q, id); err != nil {
		c.log.Debug("CommentRepository.Delete", "error", err)
	}
	return
}

func (c *CommentRepository) SetActive(ctx context.Context, id string, isActive bool) (err error) {
	const q = `UPDATE user_s.comments SET is_active = $1 WHERE id = $2::UUID`
	if err = execOne(ctx, c.session, q, isActive, id); err != nil {
		c.log.Debug("CommentRepository.SetActive", "error", err)
	}
	return
}
