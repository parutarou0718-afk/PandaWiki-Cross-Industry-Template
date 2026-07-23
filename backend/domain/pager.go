package domain

type Pager struct {
	Page     int `json:"page" query:"page" validate:"required,min=1" message:"page must be greater than 0"`
	PageSize int `json:"per_page" query:"per_page" validate:"required,min=1" message:"per_page must be greater than 0"`
}

type PagerInfo struct {
	Total int64 `json:"total"`
}

func (p *Pager) Offset() int {
	offset := (p.Page - 1) * p.PageSize
	if offset < 0 {
		offset = 0
	}
	return offset
}

func (p *Pager) Limit() int {
	limit := p.PageSize
	if limit < 0 {
		limit = 0
	}
	if limit > 100 {
		limit = 100
	}
	return limit
}

type PaginatedResult[T any] struct {
	Total uint64 `json:"total"`
	Data  T      `json:"data"`
}

func NewPaginatedResult[T any](data T, total uint64) *PaginatedResult[T] {
	return &PaginatedResult[T]{
		Total: total,
		Data:  data,
	}
}
