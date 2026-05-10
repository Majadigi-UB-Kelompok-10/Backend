package handlers

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Action  string       `json:"action,omitempty"`
	Errors  []FieldError `json:"errors,omitempty"`
}

type SuccessResponse struct {
	Pesan      string      `json:"pesan"`
	Data       interface{} `json:"data,omitempty"`
	Pagination interface{} `json:"pagination,omitempty"`
}

type PaginationMeta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}
