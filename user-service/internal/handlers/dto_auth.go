package handlers

// =============================================================================
// Request DTOs
// =============================================================================

type RegisterRequest struct {
	// Step 1 — Informasi Dasar & Kontak
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	// Step 2 — Data Kependudukan & Keamanan
	NIK             string   `json:"nik"`
	Address         string   `json:"address"`
	BirthDate       string   `json:"birth_date"` // format mm/dd/yyyy
	Gender          string   `json:"gender"`     // LAKI_LAKI | PEREMPUAN
	Password        string   `json:"password"`
	ConfirmPassword string   `json:"confirm_password"`
	CategoryIDs []string `json:"category_ids"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
	BirthDate string `json:"birth_date"` 
	Gender    string `json:"gender"`
}

type SetPreferencesRequest struct {
	CategoryIDs []string `json:"category_ids"`
}

type UpdateRoleRequest struct {
	Role string `json:"role"`
}

// =============================================================================
// Response DTOs
// =============================================================================

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

type AuthResponse struct {
	User   UserProfile `json:"user"`
	Tokens TokenPair   `json:"tokens"`
}

type UserProfile struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	NIK       string `json:"nik"`
	Address   string `json:"address,omitempty"`
	BirthDate string `json:"birth_date,omitempty"`
	Gender    string `json:"gender,omitempty"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type UserListItem struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	NIK       string `json:"nik"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

type UserRoleResponse struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	UpdatedAt string `json:"updated_at"`
}
