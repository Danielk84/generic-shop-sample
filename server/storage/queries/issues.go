package queries

import (
	"context"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/database"

	"github.com/jackc/pgx/v5"
)

type IssuesIDRequest struct {
	ID string `json:"id" binding:"required,uuid"`
}

type IssuesUserIDRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
}

type GetIssuesRequest struct {
	IssuesIDRequest
	IssuesUserIDRequest
}

type CreateIssuesRequest struct {
	IssuesUserIDRequest
	Req string `json:"req" binding:"required,max=5000,min=4"`
}

type UpdateIssuesRequest struct {
	IssuesIDRequest
	CreateIssuesRequest
}

type SetIssuesResRequest struct {
	ID     string `json:"id" binding:"required,uuid"`
	Res    string `json:"res" binding:"required,max=5000,min=1"`
	IsDone bool   `json:"is_done" binding:"required"`
}

type IssuesSummaryResponse struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	IsDone bool   `json:"is_done"`
}

type IssuesResponse struct {
	IssuesSummaryResponse
	Req string `json:"req"`
	Res string `json:"res"`
}

type issuesRepository struct {
	session database.Session
	log     logger.Logger
}

type IssuesStore interface {
	Create(ctx context.Context, issues CreateIssuesRequest) (string, error)

	UserList(ctx context.Context, userID string, pagination, page int) ([]IssuesSummaryResponse, error)
	UserMaxPage(userID string) MaxPageType

	AdminList(ctx context.Context, pagination, page int) ([]IssuesSummaryResponse, error)
	AdminMaxPage(ctx context.Context, pagination int) (int, error)

	Get(ctx context.Context, id GetIssuesRequest) (IssuesResponse, error)
	Update(ctx context.Context, issues UpdateIssuesRequest) error
	SetResponse(ctx context.Context, issues SetIssuesResRequest) error
}

func NewIssuesStore(session database.Session, log logger.Logger) IssuesStore {
	return &issuesRepository{session, log}
}

func (i *issuesRepository) Create(ctx context.Context, issues CreateIssuesRequest) (item string, err error) {
	const q = `INSERT INTO user_s.issues(user_id, req)
		VALUES (@UserID::UUID, @Req)
		RETURNING id`
	args := pgx.NamedArgs{
		"UserID": issues.UserID,
		"Req":    issues.Req,
	}
	if err := i.session.QueryRow(ctx, q, args).Scan(&item); err != nil {
		i.log.Debug("issuesRepository.Create", "error", err)
	}
	return
}

func (i *issuesRepository) UserList(
	ctx context.Context,
	userID string,
	pagination, page int,
) (items []IssuesSummaryResponse, err error) {
	const q = `SELECT id, user_id, is_done
		FROM user_s.issues
		WHERE user_id = @UserID::UUID
		LIMIT @Limit
		OFFSET @Offset`
	args := pgx.NamedArgs{
		"UserID": userID,
		"Limit":  pagination,
		"Offset": getOffsetFromPageNum(pagination, page),
	}
	items, err = list[IssuesSummaryResponse](ctx, i.session, q, args)
	if err != nil {
		i.log.Debug("issuesRepository.List", "error", err)
	}
	return
}

func (i *issuesRepository) UserMaxPage(userID string) MaxPageType {
	const q = `SELECT COUNT(*)
		FROM user_s.issues
		WHERE user_id = $1::UUID`
	return func(ctx context.Context, pagination int) (count int, err error) {
		if err = i.session.QueryRow(ctx, q, userID).Scan(&count); err != nil {
			i.log.Debug("issuesRepository.UserMaxPage", "error", err)
			return
		}
		count = getPageSize(count, pagination)
		return
	}
}

func (i *issuesRepository) AdminList(
	ctx context.Context,
	pagination, page int,
) (items []IssuesSummaryResponse, err error) {
	const q = `SELECT id, user_id, is_done
		FROM user_s.issues
		ORDER BY id
		LIMIT @Limit
		OFFSET @Offset`
	args := pgx.NamedArgs{
		"Limit":  pagination,
		"Offset": getOffsetFromPageNum(pagination, page),
	}
	items, err = list[IssuesSummaryResponse](ctx, i.session, q, args)
	if err != nil {
		i.log.Debug("issuesRepository.List", "error", err)
	}
	return
}

func (i *issuesRepository) AdminMaxPage(ctx context.Context, pagination int) (int, error) {
	return getMaxPage(ctx, i.session, "user_s.issues", pagination)
}

func (i *issuesRepository) Get(ctx context.Context, id GetIssuesRequest) (item IssuesResponse, err error) {
	const q = `SELECT id, user_id, req, res, is_done
		FROM user_s.issues
		WHERE id = @ID::UUID AND user_id = @UserID::UUID
		LIMIT 1`
	args := pgx.NamedArgs{
		"ID":     id.ID,
		"UserID": id.UserID,
	}
	item, err = get[IssuesResponse](ctx, i.session, q, args)
	if err != nil {
		i.log.Debug("issuesRepository.Get", "error", err)
	}
	return
}

func (i *issuesRepository) Update(ctx context.Context, issues UpdateIssuesRequest) (err error) {
	const q = `UPDATE user_s.issues
		SET req = @Req
		WHERE
			is_done = FALSE AND
			id = @ID::UUID AND
			user_id = @UserID::UUID`
	args := pgx.NamedArgs{
		"Req":    issues.Req,
		"ID":     issues.ID,
		"UserID": issues.UserID,
	}
	if err := execOne(ctx, i.session, q, args); err != nil {
		i.log.Debug("issuesRepository.Update", "error", err)
	}
	return
}

func (i *issuesRepository) SetResponse(ctx context.Context, issues SetIssuesResRequest) (err error) {
	const q = `UPDATE user_s.issues
		SET
			res = @Res,
			is_done = @IsDone
		WHERE id = @ID::UUID`
	args := pgx.NamedArgs{
		"Res":    issues.Res,
		"IsDone": issues.IsDone,
		"ID":     issues.ID,
	}
	if err = execOne(ctx, i.session, q, args); err != nil {
		i.log.Debug("issuesRepository.SetResponse", "error", err)
	}
	return
}
