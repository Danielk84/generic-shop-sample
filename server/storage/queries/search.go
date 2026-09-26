package queries

import (
	"context"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/database"

	"github.com/jackc/pgx/v5"
)

type searchRepository struct {
	session database.Session
	log     logger.Logger
}

type SearchStore interface {
	Reindex(ctx context.Context, product_id string) error
	Search(ctx context.Context, queryStr string, pagination, page int) ([]ProductSummaryResponse, error)
	SearchAll(ctx context.Context, queryStr string, pagination, page int) ([]ProductSummaryResponse, error)
	DeleteAll(ctx context.Context) error
}

func NewSearchStore(session database.Session, log logger.Logger) SearchStore {
	return &searchRepository{session, log}
}

func (s *searchRepository) Reindex(ctx context.Context, product_id string) (err error) {
	const q = `INSERT INTO full_text_search_s.products_changes(product_id)
		VALUES ($1::UUID)
		ON CONFLICT DO NOTHING`
	if err = execOne(ctx, s.session, q, product_id); err != nil {
		s.log.Debug("searchRepository.Reindex", "error", err)
	}
	return
}

func (s *searchRepository) Search(
	ctx context.Context,
	queryStr string,
	pagination, page int,
) (items []ProductSummaryResponse, err error) {
	const q = `SELECT p.id, p.name, p.price, p.pub_date,
			COALESCE(i.img_path, '') AS img_path
		FROM product_s.products AS p
		LEFT JOIN product_s.product_images AS i
			ON p.id = i.product_id AND i.pos = 0
		WHERE p.is_active = true AND
			p.__search @@ websearch_to_tsquery('simple', @QueryStr)
		ORDER BY
			p.pub_date DESC,
			p.available_quantity DESC,
			p.price,
			p.is_available DESC
		LIMIT @Limit
		OFFSET @Offset`
	args := pgx.NamedArgs{
		"QueryStr": queryStr,
		"Limit":    pagination,
		"Offset":   getOffsetFromPageNum(pagination, page),
	}
	items, err = list[ProductSummaryResponse](ctx, s.session, q, args)
	if err != nil {
		s.log.Debug("searchRepository.Search", "error", err)
	}
	return
}

func (s *searchRepository) SearchAll(
	ctx context.Context,
	queryStr string,
	pagination, page int,
) (items []ProductSummaryResponse, err error) {
	const q = `SELECT p.id, p.name, p.price, p.pub_date,
			COALESCE(i.img_path, '') AS img_path
		FROM product_s.products AS p
		LEFT JOIN product_s.product_images AS i
			ON p.id = i.product_id AND i.pos = 0
		WHERE p.__search @@ websearch_to_tsquery('simple', @QueryStr)
		ORDER BY
			p.pub_date DESC,
			p.available_quantity DESC,
			p.price,
			p.is_available DESC
		LIMIT @Limit
		OFFSET @Offset`
	args := pgx.NamedArgs{
		"QueryStr": queryStr,
		"Limit":    pagination,
		"Offset":   getOffsetFromPageNum(pagination, page),
	}
	items, err = list[ProductSummaryResponse](ctx, s.session, q, args)
	if err != nil {
		s.log.Debug("searchRepository.SearchAll", "error", err)
	}
	return
}

// on deleting rows in full_text_search_s.products_changes,
// TRIGGER _01_set_products_search_on_delete_t will execute,
// that process product_s.products.__search.
func (s *searchRepository) DeleteAll(ctx context.Context) (err error) {
	const q = `DELETE FROM full_text_search_s.products_changes`
	if err = execOne(ctx, s.session, q); err != nil {
		s.log.Debug("searchRepository.DeleteAll", "error", err)
	}
	return
}
