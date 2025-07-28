package model

type PaginationMetadata struct {
	PageSize      int    `json:"page_size"`
	NextPageToken string `json:"next_page_token,omitempty"`
	HasMore       bool   `json:"has_more"`
}

type RecordsResponse struct {
	Data       []Record           `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
}

type APagingResponse struct {
	Data       []ARecord          `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
}

type AAAAPagingResponse struct {
	Data       []AAAARecord       `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
}
