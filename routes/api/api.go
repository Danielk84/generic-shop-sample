package api

import (
	"generic-shop-sample/internal"
	"strconv"
)

type SetFlag struct {
	Accepted bool `json:"accepted"`
}

var defaultPagination = internal.NewConfig().Pagination

func getOffsetFromPageNum(p string) int {
	page, err := strconv.Atoi(p)
	if err != nil || page < 1 {
		return 1
	}
	return (page - 1) * defaultPagination
}
