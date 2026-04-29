package handlers

type CreateReportRequest struct {
	ReporterName  string `json:"reporter_name"`
	ReporterEmail string `json:"reporter_email"`
	ReporterPhone string `json:"reporter_phone"`
	Content       string `json:"content"`
	ProofLink     string `json:"proof_link,omitempty"`
	ProofImageUrl string `json:"proof_image_url,omitempty"`
}


type CreateNewsRequest struct {
	ReportID      string `json:"report_id,omitempty"`
	CategoryID    string `json:"category_id"`
	Title         string `json:"title"`
	Slug          string `json:"slug"`
	Description   string `json:"description"`
	ReferenceLink string `json:"reference_link,omitempty"`
	ImageUrl      string `json:"image_url"`
}


type UpdateReportStatusRequest struct {
	Status string `json:"status"`
}

type CreateCategoryRequest struct {
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	IconUrl string `json:"icon_url,omitempty"`
}


type TicketCreatedResponse struct {
	TicketNumber string `json:"ticket_number"`
	CreatedAt    string `json:"created_at"`
}


type ReportTrackingResponse struct {
	ReportID     string `json:"report_id"`
	TicketNumber string `json:"ticket_number"`
	ReporterName string `json:"reporter_name"`
	ReportStatus string `json:"report_status"`
	ReportedAt   string `json:"reported_at"`
	NewsID       string `json:"news_id,omitempty"`
	NewsTitle    string `json:"news_title,omitempty"`
	NewsSlug     string `json:"news_slug,omitempty"` 
	NewsImage    string `json:"news_image,omitempty"`
	CategoryName string `json:"category_name,omitempty"`
}


type DashboardStatItem struct {
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	CategorySlug string `json:"category_slug"`
	IconUrl      string `json:"icon_url,omitempty"`
	TotalNews    int64  `json:"total_news"`
}

type PublicNewsItem struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Slug         string `json:"slug"`
	ImageUrl     string `json:"image_url"`
	CategoryName string `json:"category_name"`
	CategorySlug string `json:"category_slug"`
	PublishedAt  string `json:"published_at"`
}

type PublicNewsDetail struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Slug          string `json:"slug"`
	Description   string `json:"description"`
	ReferenceLink string `json:"reference_link,omitempty"`
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
	ProofLink     string `json:"proof_link,omitempty"`
	ProofImageUrl string `json:"proof_image_url,omitempty"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}


type AdminNewsItem struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Slug         string `json:"slug"`
	CategoryName string `json:"category_name"`
	PublishedAt  string `json:"published_at"`
	CreatedAt    string `json:"created_at"`
	TicketNumber string `json:"ticket_number,omitempty"`
}

type CategoryItem struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	IconUrl string `json:"icon_url,omitempty"`
}