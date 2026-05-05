// Package paging holds the standard list-pagination DTO.
package paging

// Query is a generic page/limit query.
type Query struct {
	Page  int `form:"page"  json:"page"  validate:"min=0"`
	Limit int `form:"limit" json:"limit" validate:"min=0,max=200"`
}

// Normalize clamps Page/Limit to safe defaults.
func (q *Query) Normalize() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Limit <= 0 || q.Limit > 200 {
		q.Limit = 20
	}
}

// Offset returns the SQL OFFSET for the current page.
func (q Query) Offset() int { return (q.Page - 1) * q.Limit }
