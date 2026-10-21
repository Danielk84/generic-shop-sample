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
	Body     string `json:"body" binding:"required,max=5000"`
}

type CommentResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	PubDate        time.Time `json:"pub_date"`
	ChildrenAmount int32     `json:"children_amount"`
	Body           string    `json:"body"`
}

type RelatedCommentResponse struct {
	CommentResponse
	UserID   string `json:"user_id"`
	Parent   string `json:"parent"`
	Referrer string `json:"referrer"`
	IsActive bool   `json:"is_active"`
}

type commentRepository struct {
	session database.Session
	log     logger.Logger
}

type CommentStore interface {
	Create(ctx context.Context, userID, name string, comment *CommentRequest) error
	Get(ctx context.Context, id string) (RelatedCommentResponse, error)
	List(ctx context.Context, parent string, referrer string, pagination, page int) ([]CommentResponse, error)
	MaxListPage(parent, referrer string) MaxPageType
	FullList(ctx context.Context, userID string, pagination, page int) ([]RelatedCommentResponse, error)
	MaxFullListPage(userID string) MaxPageType
	Find(ctx context.Context, id string, pagination, page int) ([]RelatedCommentResponse, error)
	MaxFindPage(id string) MaxPageType
	Delete(ctx context.Context, id string) (string, error)
	SetActive(ctx context.Context, id string, isActive bool) error
}

func NewCommentStore(session database.Session, log logger.Logger) CommentStore {
	return &commentRepository{session, log}
}

func (c *commentRepository) Create(ctx context.Context, userID, name string, comment *CommentRequest) (err error) {
	const createCommentQuery = `INSERT INTO user_s.comments(user_id, name, parent, children_amount, referrer, body)
		VALUES (@UserID, @Name, NULLIF(@Parent, '')::UUID, 0, @Referrer, @Body)`
	const updateChildrenCount = `UPDATE user_s.comments
		SET children_amount = children_amount + 1
		WHERE id = $1::UUID AND referrer = $2::UUID AND is_active = TRUE`

	args := pgx.NamedArgs{
		"UserID":   userID,
		"Name":     name,
		"Parent":   comment.Parent,
		"Referrer": comment.Referrer,
		"Body":     comment.Body,
	}
	err = pgx.BeginFunc(ctx, c.session, func(tx pgx.Tx) error {
		if comment.Parent != "" {
			tag, err := tx.Exec(ctx, updateChildrenCount, comment.Parent, comment.Referrer)
			if err != nil {
				return err
			}
			if tag.RowsAffected() != 1 {
				return ErrNotFound
			}
		}

		cTag, err := tx.Exec(ctx, createCommentQuery, args)
		if err != nil {
			return err
		}
		if cTag.RowsAffected() != 1 {
			return pgx.ErrNoRows
		}
		return nil
	})
	if err != nil {
		c.log.Debug("CommentRepository.Create", "error", err)
	}
	return
}

func (c *commentRepository) Get(ctx context.Context, id string) (item RelatedCommentResponse, err error) {
	const q = `SELECT
			id, user_id, COALESCE(name, 'deleted') AS name,
			pub_date, COALESCE(parent::TEXT, '') AS parent, children_amount,
			referrer, body, is_active
		FROM user_s.comments
		WHERE id = $1::UUID`
	item, err = get[RelatedCommentResponse](ctx, c.session, q, id)
	if err != nil {
		c.log.Debug("CommentRepository.Get", "error", err)
	}
	return
}

func (c *commentRepository) List(ctx context.Context, parent string, referrer string, pagination, page int) (items []CommentResponse, err error) {
	const q = `SELECT
			id, COALESCE(name, 'deleted') AS name, pub_date, children_amount, body
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

func (c *commentRepository) MaxListPage(parent, referrer string) MaxPageType {
	const q = `SELECT COUNT(*)
		FROM user_s.comments
		WHERE COALESCE(parent::TEXT, '') = $1
			AND referrer = $2
			AND is_active = TRUE`
	return func(ctx context.Context, pagination int) (count int, err error) {
		if err = c.session.QueryRow(ctx, q, parent, referrer).Scan(&count); err != nil {
			c.log.Debug("commentRepository.MaxListPage", "error", err)
			return
		}
		count = getPageSize(count, pagination)
		return
	}
}

func (c *commentRepository) FullList(
	ctx context.Context,
	userID string,
	pagination, page int,
) (items []RelatedCommentResponse, err error) {
	const baseQuery = `SELECT
			id, user_id, COALESCE(name, 'deleted') AS name,
			pub_date, COALESCE(parent::TEXT, '') AS parent, children_amount,
			referrer, body, is_active
		FROM user_s.comments`
	const limitOffset = ` LIMIT @Limit OFFSET @Offset`
	args := pgx.NamedArgs{
		"Limit":  pagination,
		"Offset": getOffsetFromPageNum(pagination, page),
	}

	var q string
	if userID == "" {
		q = baseQuery + " ORDER BY pub_date DESC, is_active" + limitOffset
	} else {
		args["UserID"] = userID
		q = baseQuery + " WHERE user_id = @UserID::UUID ORDER BY pub_date DESC" + limitOffset
	}
	items, err = list[RelatedCommentResponse](ctx, c.session, q, args)
	if err != nil {
		c.log.Debug("CommentRepository.FullList", "error", err)
	}
	return
}

func (c *commentRepository) MaxFullListPage(userID string) MaxPageType {
	if userID == "" {
		return func(ctx context.Context, pagination int) (int, error) {
			return getMaxPage(ctx, c.session, "user_s.comments", pagination)
		}
	}
	const q = `SELECT COUNT(*) FROM user_s.comments WHERE user_id = $1::UUID`
	return func(ctx context.Context, pagination int) (count int, err error) {
		if err = c.session.QueryRow(ctx, q, userID).Scan(&count); err != nil {
			c.log.Debug("commentRepository.MaxFullListPage", "error", err)
			return
		}
		count = getPageSize(count, pagination)
		return
	}
}

func (c *commentRepository) Find(
	ctx context.Context,
	id string,
	pagination, page int,
) (items []RelatedCommentResponse, err error) {
	const q = `SELECT
			id, user_id, COALESCE(name, 'deleted') AS name,
			pub_date, COALESCE(parent::TEXT, '') AS parent, children_amount,
			referrer, body, is_active
		FROM user_s.comments
		WHERE
			id = @ID::UUID OR
			parent = @ID::UUID
		ORDER BY parent DESC
		LIMIT @Limit OFFSET @Offset`
	args := pgx.NamedArgs{
		"ID":     id,
		"Limit":  pagination,
		"Offset": getOffsetFromPageNum(pagination, page),
	}
	items, err = list[RelatedCommentResponse](ctx, c.session, q, args)
	if err != nil {
		c.log.Debug("CommentRepository.FullList", "error", err)
	}
	return
}

func (c *commentRepository) MaxFindPage(id string) MaxPageType {
	const q = `SELECT COUNT(*)
		FROM user_s.comments
		WHERE id = $1::UUID OR parent = $1::UUID`
	return func(ctx context.Context, pagination int) (count int, err error) {
		if err = c.session.QueryRow(ctx, q, id).Scan(&count); err != nil {
			c.log.Debug("commentRepository.MaxFindPage", "error", err)
			return
		}
		count = getPageSize(count, pagination)
		return
	}
}

func (c *commentRepository) Delete(ctx context.Context, id string) (item string, err error) {
	const selectQuery = `SELECT referrer, COALESCE(parent::TEXT, '')
		FROM user_s.comments
		WHERE id = $1::UUID`
	const deleteQuery = `DELETE FROM user_s.comments
		WHERE id = $1::UUID OR parent = $1::UUID`
	const decrementParentQuery = `UPDATE user_s.comments
		SET children_amount = GREATEST(children_amount - 1, 0)
		WHERE id = $1::UUID`

	err = pgx.BeginFunc(ctx, c.session, func(tx pgx.Tx) error {
		var parent string
		if err := tx.QueryRow(ctx, selectQuery, id).Scan(&item, &parent); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, deleteQuery, id); err != nil {
			return err
		}
		if parent != "" {
			if _, err := tx.Exec(ctx, decrementParentQuery, parent); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		c.log.Debug("CommentRepository.Delete", "error", err)
	}
	return
}

func (c *commentRepository) SetActive(ctx context.Context, id string, isActive bool) (err error) {
	const q = `UPDATE user_s.comments SET is_active = $1 WHERE id = $2::UUID`
	if err = execOne(ctx, c.session, q, isActive, id); err != nil {
		c.log.Debug("CommentRepository.SetActive", "error", err)
	}
	return
}
