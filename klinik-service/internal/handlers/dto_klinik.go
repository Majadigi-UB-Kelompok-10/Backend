package handlers

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
	Slug         string `json:"slug"`
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

type AdminNewsItem struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	CategoryName string `json:"category_name"`
	PublishedAt  string `json:"published_at"`
	CreatedAt    string `json:"created_at"`
	TicketNumber string `json:"ticket_number,omitempty"`
}

type CategoryItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	IconUrl  string `json:"icon_url"`
}