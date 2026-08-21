package pagination

type Request struct {
	Limit  int
	Cursor string
}
type Response[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
}

func Normalize(r Request) Request {
	if r.Limit < 1 {
		r.Limit = 50
	}
	if r.Limit > 200 {
		r.Limit = 200
	}
	return r
}
