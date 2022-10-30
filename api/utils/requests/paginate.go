package requests

type Pagination struct {
	Data any
}

type PageMetadata struct {
	Page       int
	PageSize   int
	TotalItems int
	TotalPages int
}
