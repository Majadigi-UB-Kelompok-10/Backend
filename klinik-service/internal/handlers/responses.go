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


type TicketCreatedResponse struct {
	TicketNumber string `json:"ticket_number"`
}

type ReportTrackingResponse struct {
	ReportID     string `json:"report_id"`
	TicketNumber string `json:"ticket_number"`
	ReporterName string `json:"reporter_name"`
	ReportStatus string `json:"report_status"`
	ReportedAt   string `json:"reported_at"`
	NewsID       string `json:"news_id,omitempty"`
	NewsTitle    string `json:"news_title,omitempty"`
	NewsImage    string `json:"news_image,omitempty"`
	CategoryName string `json:"category_name,omitempty"`
}

type DashboardStatItem struct {
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	CategorySlug string `json:"category_slug"`
	TotalNews    int64  `json:"total_news"`
}

type PublicNewsItem struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Slug		 string `json:"slug"`
	ImageUrl     string `json:"image_url"`
	CategoryName string `json:"category_name"`
	CategorySlug string `json:"category_slug"`
	PublishedAt  string `json:"published_at"`
}

type PublicNewsDetail struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	ReferenceLink string `json:"reference_link"`
	ImageUrl      string `json:"image_url"`
	CategoryName  string `json:"category_name"`
	CategorySlug  string `json:"category_slug"`
	PublishedAt   string `json:"published_at"`
}

type AdminReportItem struct {
	ID            string `json:"id"`
	TicketNumber  string `json:"ticket_number"`
	ReporterName  string `json:"reporter_name"`
	ReporterEmail string `json:"reporter_email"`
	Content       string `json:"content"`
	ProofLink     string `json:"proof_link"`
	ProofImageUrl string `json:"proof_image_url"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}