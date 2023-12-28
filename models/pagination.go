package models

const defaultPageSize = 2

type Page struct {
	Page     int
	PageSize int
}

func NewPage(page, pageSize int) *Page {
	if pageSize == 0 {
		pageSize = defaultPageSize
	}

	if page < 1 {
		page = 1
	}

	return &Page{
		Page:     page,
		PageSize: pageSize,
	}
}

func (p *Page) Offset() int {
	return (p.Page - 1) * p.PageSize
}

func (p *Page) Limit() int {
	return p.PageSize
}

type PaginatedResponse[T any] struct {
	Count int
	Data  []T
}
