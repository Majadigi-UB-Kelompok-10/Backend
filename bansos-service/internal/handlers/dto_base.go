package handlers

type SuccessResponse struct {
	Pesan      string          `json:"pesan"`
	Data       interface{}     `json:"data,omitempty"`
	Pagination *PaginationMeta `json:"pagination,omitempty"`
}

type ErrorResponse struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Action  string       `json:"action,omitempty"`
	Errors  []FieldError `json:"errors,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type PaginationMeta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type PaginatedResponse struct {
	Pesan      string          `json:"pesan"`
	Data       interface{}     `json:"data"`
	Pagination *PaginationMeta `json:"pagination"`
}