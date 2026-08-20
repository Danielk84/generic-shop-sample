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
	const q = `SELECT id, name, price, pub_date
		FROM product_s.products
		WHERE is_active = true
		ORDER BY
			pub_date DESC,
			available_quantity DESC,
			price,
			is_available DESC
		WHERE __search @@ websearch_to_tsquery('simple', @QueryStr)
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
