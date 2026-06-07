package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/farildzaky/user-service/internal/db"
	emailpkg "github.com/farildzaky/user-service/internal/email"
	"github.com/farildzaky/user-service/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// -----------------------------------------------------------------------------
// POST /api/v1/auth/register
// -----------------------------------------------------------------------------

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_BIND", Message: "Request tidak valid", Detail: err.Error()})
	}

	var fieldErrors []FieldError

	firstName, ve := utils.ValidateQueryString(req.FirstName, 100, "first_name")
	if ve != nil || strings.TrimSpace(req.FirstName) == "" {
		fieldErrors = append(fieldErrors, FieldError{Field: "first_name", Message: "nama depan wajib diisi"})
	}

	lastName, ve := utils.ValidateQueryString(req.LastName, 100, "last_name")
	if ve != nil || strings.TrimSpace(req.LastName) == "" {
		fieldErrors = append(fieldErrors, FieldError{Field: "last_name", Message: "nama belakang wajib diisi"})
	}

	phone, ve := utils.ValidatePhone(req.Phone)
	if ve != nil {
		fieldErrors = append(fieldErrors, FieldError{Field: ve.Field, Message: ve.Message})
	}

	email, ve := utils.ValidateEmail(req.Email)
	if ve != nil {
		fieldErrors = append(fieldErrors, FieldError{Field: ve.Field, Message: ve.Message})
	}

	nik, ve := utils.ValidateNIK(req.NIK)
	if ve != nil {
		fieldErrors = append(fieldErrors, FieldError{Field: ve.Field, Message: ve.Message})
	}

	if ve := utils.ValidatePassword(req.Password); ve != nil {
		fieldErrors = append(fieldErrors, FieldError{Field: ve.Field, Message: ve.Message})
	}
	if req.Password != req.ConfirmPassword {
		fieldErrors = append(fieldErrors, FieldError{Field: "confirm_password", Message: "konfirmasi password tidak cocok"})
	}

	birthTime, ve := utils.ValidateBirthDate(req.BirthDate)
	if ve != nil {
		fieldErrors = append(fieldErrors, FieldError{Field: ve.Field, Message: ve.Message})
	}

	if len(fieldErrors) > 0 {
		return validationErrors(c, fieldErrors)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	// Uniqueness checks (parallel)
	if _, err := h.Queries.CheckEmailExists(ctx, email); err == nil {
		return c.Status(409).JSON(ErrorResponse{Code: "ERR_EMAIL_TAKEN", Message: "Email sudah digunakan"})
	}
	if _, err := h.Queries.CheckPhoneExists(ctx, phone); err == nil {
		return c.Status(409).JSON(ErrorResponse{Code: "ERR_PHONE_TAKEN", Message: "Nomor HP sudah digunakan"})
	}
	if _, err := h.Queries.CheckNIKExists(ctx, nik); err == nil {
		return c.Status(409).JSON(ErrorResponse{Code: "ERR_NIK_TAKEN", Message: "NIK sudah terdaftar"})
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		slog.Error("auth.register.hash_error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal memproses password"})
	}

	params := db.CreateUserParams{
		FirstName:    firstName,
		LastName:     lastName,
		Phone:        phone,
		Email:        email,
		Nik:          nik,
		PasswordHash: hash,
	}
	if req.Address != "" {
		params.Address = pgtype.Text{String: req.Address, Valid: true}
	}
	if !birthTime.IsZero() {
		params.BirthDate = pgtype.Date{Time: birthTime, Valid: true, InfinityModifier: pgtype.Finite}
	}
	if req.Gender == string(db.GenderTypeLAKILAKI) || req.Gender == string(db.GenderTypePEREMPUAN) {
		params.Gender = db.NullGenderType{GenderType: db.GenderType(req.Gender), Valid: true}
	}

	row, err := h.Queries.CreateUser(ctx, params)
	if err != nil {
		slog.Error("auth.register.create_user_error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal membuat akun"})
	}

	// Generate and store email verification token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err == nil {
		token := hex.EncodeToString(tokenBytes)
		expiry := pgtype.Timestamptz{Time: time.Now().UTC().Add(24 * time.Hour), Valid: true}
		_ = h.Queries.SetEmailVerificationToken(ctx, db.SetEmailVerificationTokenParams{
			ID:                   row.ID,
			EmailVerifyToken:     pgtype.Text{String: token, Valid: true},
			EmailVerifyExpiresAt: expiry,
		})

		baseURL := os.Getenv("APP_BASE_URL")
		if baseURL == "" {
			port := os.Getenv("PORT")
			if port == "" {
				port = "3000"
			}
			baseURL = "http://localhost:" + port
		}
		verifyURL := baseURL + "/api/v1/user/auth/verify-email?token=" + token
		emailpkg.SendVerificationEmailAsync(row.Email, row.FirstName, verifyURL)
	}

	return c.Status(201).JSON(SuccessResponse{
		Pesan: "Registrasi berhasil. Silakan cek email Anda untuk verifikasi akun sebelum login.",
		Data:  fiber.Map{"email": row.Email},
	})
}

// -----------------------------------------------------------------------------
// POST /api/v1/auth/login
// -----------------------------------------------------------------------------

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_BIND", Message: "Request tidak valid"})
	}

	email, ve := utils.ValidateEmail(req.Email)
	if ve != nil {
		return validationError(c, ve)
	}
	if req.Password == "" {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION_FAILED", Message: "Password wajib diisi"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	user, err := h.Queries.GetUserByEmailForAuth(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(401).JSON(ErrorResponse{Code: "ERR_INVALID_CREDENTIALS", Message: "Email atau password salah"})
		}
		slog.Error("auth.login.db_error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal memproses login"})
	}

	if !user.IsActive {
		return c.Status(403).JSON(ErrorResponse{Code: "ERR_ACCOUNT_INACTIVE", Message: "Akun tidak aktif, hubungi admin"})
	}

	if !user.EmailVerified {
		return c.Status(403).JSON(ErrorResponse{Code: "ERR_EMAIL_NOT_VERIFIED", Message: "Email belum diverifikasi. Silakan cek inbox Anda dan klik link verifikasi."})
	}

	if !utils.VerifyPassword(req.Password, user.PasswordHash) {
		return c.Status(401).JSON(ErrorResponse{Code: "ERR_INVALID_CREDENTIALS", Message: "Email atau password salah"})
	}

	userIDStr := uuidToString(user.ID)

	accessToken, refreshToken, err := h.issueTokenPair(ctx, userIDStr, string(user.Role))
	if err != nil {
		slog.Error("auth.login.token_error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal membuat token"})
	}

	slog.Info("auth.login.success", slog.String("user_id", userIDStr), slog.String("role", string(user.Role)))

	return c.JSON(SuccessResponse{
		Pesan: "Login berhasil",
		Data: AuthResponse{
			User:   toUserProfileFromAuthRow(user),
			Tokens: TokenPair{AccessToken: accessToken, RefreshToken: refreshToken, TokenType: "Bearer"},
		},
	})
}

// -----------------------------------------------------------------------------
// POST /api/v1/auth/refresh
// -----------------------------------------------------------------------------

func (h *AuthHandler) RefreshToken(c fiber.Ctx) error {
	var req RefreshTokenRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_BIND", Message: "Request tidak valid"})
	}
	if req.RefreshToken == "" {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION_FAILED", Message: "refresh_token wajib diisi"})
	}

	claims, err := utils.ValidateToken(req.RefreshToken, h.JWTSecret)
	if err != nil || claims.Type != utils.TokenTypeRefresh {
		return c.Status(401).JSON(ErrorResponse{Code: "ERR_INVALID_TOKEN", Message: "Refresh token tidak valid atau sudah kadaluarsa"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	tokenHash := utils.HashToken(req.RefreshToken)
	stored, err := h.Queries.GetRefreshToken(ctx, tokenHash)
	if err != nil || stored.IsRevoked || stored.ExpiresAt.Time.Before(time.Now()) {
		return c.Status(401).JSON(ErrorResponse{Code: "ERR_TOKEN_REVOKED", Message: "Refresh token sudah tidak valid"})
	}

	// Rotate: revoke old token
	if err := h.Queries.RevokeRefreshToken(ctx, tokenHash); err != nil {
		slog.Warn("auth.refresh.revoke_error", slog.String("err", err.Error()))
	}

	// Get fresh user data
	userID, err := stringToUUID(claims.UserID)
	if err != nil {
		return c.Status(401).JSON(ErrorResponse{Code: "ERR_INVALID_TOKEN", Message: "Token berisi user ID tidak valid"})
	}
	user, err := h.Queries.GetUserByID(ctx, userID)
	if err != nil || !user.IsActive {
		return c.Status(401).JSON(ErrorResponse{Code: "ERR_ACCOUNT_INACTIVE", Message: "Akun tidak aktif"})
	}

	accessToken, refreshToken, err := h.issueTokenPair(ctx, claims.UserID, string(user.Role))
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal membuat token baru"})
	}

	return c.JSON(SuccessResponse{
		Pesan: "Token diperbarui",
		Data:  TokenPair{AccessToken: accessToken, RefreshToken: refreshToken, TokenType: "Bearer"},
	})
}

// -----------------------------------------------------------------------------
// POST /api/v1/auth/logout
// -----------------------------------------------------------------------------

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	var req LogoutRequest
	if err := c.Bind().Body(&req); err != nil || req.RefreshToken == "" {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION_FAILED", Message: "refresh_token wajib diisi"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	tokenHash := utils.HashToken(req.RefreshToken)
	if err := h.Queries.RevokeRefreshToken(ctx, tokenHash); err != nil {
		slog.Warn("auth.logout.revoke_error", slog.String("err", err.Error()))
	}

	return c.JSON(SuccessResponse{Pesan: "Logout berhasil"})
}

// -----------------------------------------------------------------------------
// GET /api/v1/auth/me  (requires JWT middleware)
// -----------------------------------------------------------------------------

func (h *AuthHandler) GetMe(c fiber.Ctx) error {
	userIDStr, _ := c.Locals("user_id").(string)

	cacheKey := "user:me:" + userIDStr
	if respondCached(c, cacheKey) {
		return nil
	}

	userID, err := stringToUUID(userIDStr)
	if err != nil {
		return c.Status(401).JSON(ErrorResponse{Code: "ERR_UNAUTHORIZED", Message: "Token tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	user, err := h.Queries.GetUserByID(ctx, userID)
	if err != nil {
		return c.Status(404).JSON(ErrorResponse{Code: "ERR_NOT_FOUND", Message: "Pengguna tidak ditemukan"})
	}

	profile := toUserProfileFromIDRow(user)
	res := SuccessResponse{Pesan: "Sukses", Data: profile}
	return cacheJSON(c, cacheKey, CacheTTLDetail, res)
}

// -----------------------------------------------------------------------------
// PUT /api/v1/auth/me  (requires JWT middleware)
// -----------------------------------------------------------------------------

func (h *AuthHandler) UpdateMe(c fiber.Ctx) error {
	userIDStr, _ := c.Locals("user_id").(string)

	var req UpdateProfileRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_BIND", Message: "Request tidak valid"})
	}

	var fieldErrors []FieldError
	firstName, ve := utils.ValidateQueryString(req.FirstName, 100, "first_name")
	if ve != nil || strings.TrimSpace(req.FirstName) == "" {
		fieldErrors = append(fieldErrors, FieldError{Field: "first_name", Message: "nama depan wajib diisi"})
	}
	lastName, ve := utils.ValidateQueryString(req.LastName, 100, "last_name")
	if ve != nil || strings.TrimSpace(req.LastName) == "" {
		fieldErrors = append(fieldErrors, FieldError{Field: "last_name", Message: "nama belakang wajib diisi"})
	}
	phone, ve := utils.ValidatePhone(req.Phone)
	if ve != nil {
		fieldErrors = append(fieldErrors, FieldError{Field: ve.Field, Message: ve.Message})
	}
	birthTime, ve := utils.ValidateBirthDate(req.BirthDate)
	if ve != nil {
		fieldErrors = append(fieldErrors, FieldError{Field: ve.Field, Message: ve.Message})
	}
	if len(fieldErrors) > 0 {
		return validationErrors(c, fieldErrors)
	}

	userID, err := stringToUUID(userIDStr)
	if err != nil {
		return c.Status(401).JSON(ErrorResponse{Code: "ERR_UNAUTHORIZED", Message: "Token tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	params := db.UpdateUserProfileParams{
		ID:        userID,
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
	}
	if req.Address != "" {
		params.Address = pgtype.Text{String: req.Address, Valid: true}
	}
	if !birthTime.IsZero() {
		params.BirthDate = pgtype.Date{Time: birthTime, Valid: true, InfinityModifier: pgtype.Finite}
	}
	if req.Gender == string(db.GenderTypeLAKILAKI) || req.Gender == string(db.GenderTypePEREMPUAN) {
		params.Gender = db.NullGenderType{GenderType: db.GenderType(req.Gender), Valid: true}
	}

	row, err := h.Queries.UpdateUserProfile(ctx, params)
	if err != nil {
		slog.Error("auth.update_me.error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal memperbarui profil"})
	}

	invalidateUserCache(userIDStr)

	return c.JSON(SuccessResponse{Pesan: "Profil berhasil diperbarui", Data: toUserProfile(toCreateRowFromUpdateRow(row))})
}

// -----------------------------------------------------------------------------
// GET /api/v1/auth/preferences  (requires JWT middleware)
// -----------------------------------------------------------------------------

func (h *AuthHandler) GetPreferences(c fiber.Ctx) error {
	userIDStr, _ := c.Locals("user_id").(string)

	cacheKey := "user:prefs:" + userIDStr
	if respondCached(c, cacheKey) {
		return nil
	}

	userID, err := stringToUUID(userIDStr)
	if err != nil {
		return c.Status(401).JSON(ErrorResponse{Code: "ERR_UNAUTHORIZED", Message: "Token tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	prefs, err := h.Queries.GetUserPreferences(ctx, userID)
	if err != nil {
		slog.Error("auth.get_prefs.error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal memuat preferensi"})
	}

	ids := make([]string, 0, len(prefs))
	for _, p := range prefs {
		ids = append(ids, uuidToString(p))
	}

	res := SuccessResponse{Pesan: "Sukses", Data: fiber.Map{"category_ids": ids}}
	return cacheJSON(c, cacheKey, CacheTTLDetail, res)
}

// -----------------------------------------------------------------------------
// POST /api/v1/auth/preferences  (requires JWT middleware)
// -----------------------------------------------------------------------------

func (h *AuthHandler) SetPreferences(c fiber.Ctx) error {
	userIDStr, _ := c.Locals("user_id").(string)

	var req SetPreferencesRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_BIND", Message: "Request tidak valid"})
	}

	userID, err := stringToUUID(userIDStr)
	if err != nil {
		return c.Status(401).JSON(ErrorResponse{Code: "ERR_UNAUTHORIZED", Message: "Token tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.savePreferences(ctx, userID, req.CategoryIDs); err != nil {
		slog.Error("auth.set_prefs.error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal menyimpan preferensi"})
	}

	// Auto-seed favorites from selected categories (fire-and-forget)
	go h.seedFavoritesFromCategories(userID, req.CategoryIDs)

	invalidateUserCache(userIDStr)

	return c.JSON(SuccessResponse{Pesan: "Preferensi berhasil disimpan"})
}

// -----------------------------------------------------------------------------
// GET /api/v1/auth/verify-email?token=xxx  (public)
// -----------------------------------------------------------------------------

func (h *AuthHandler) VerifyEmail(c fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Token verifikasi tidak ditemukan"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	user, err := h.Queries.GetUserByEmailVerifyToken(ctx, pgtype.Text{String: token, Valid: true})
	if err != nil {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_INVALID_TOKEN", Message: "Token tidak valid atau sudah kadaluarsa. Silakan daftar ulang."})
	}

	if err := h.Queries.MarkEmailVerified(ctx, user.ID); err != nil {
		slog.Error("auth.verify_email.error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal memverifikasi email"})
	}

	return c.JSON(SuccessResponse{Pesan: "Email berhasil diverifikasi. Silakan login."})
}

// -----------------------------------------------------------------------------
// POST /api/v1/auth/resend-verification
// -----------------------------------------------------------------------------

func (h *AuthHandler) ResendVerification(c fiber.Ctx) error {
	var req ResendVerificationRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_BIND", Message: "Request tidak valid"})
	}

	email, ve := utils.ValidateEmail(req.Email)
	if ve != nil {
		return validationError(c, ve)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	user, err := h.Queries.GetUserByEmailBasic(ctx, email)
	if err != nil {
		return c.Status(404).JSON(ErrorResponse{Code: "ERR_EMAIL_NOT_FOUND", Message: "Email tidak terdaftar"})
	}

	if !user.IsActive {
		return c.Status(403).JSON(ErrorResponse{Code: "ERR_ACCOUNT_INACTIVE", Message: "Akun tidak aktif, hubungi admin"})
	}

	if user.EmailVerified {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_ALREADY_VERIFIED", Message: "Email sudah diverifikasi, silakan login"})
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		slog.Error("auth.resend_verification.token_error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal membuat token"})
	}
	token := hex.EncodeToString(tokenBytes)
	expiry := pgtype.Timestamptz{Time: time.Now().UTC().Add(24 * time.Hour), Valid: true}

	if err := h.Queries.SetEmailVerificationToken(ctx, db.SetEmailVerificationTokenParams{
		ID:                   user.ID,
		EmailVerifyToken:     pgtype.Text{String: token, Valid: true},
		EmailVerifyExpiresAt: expiry,
	}); err != nil {
		slog.Error("auth.resend_verification.db_error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal menyimpan token"})
	}

	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "3000"
		}
		baseURL = "http://localhost:" + port
	}
	verifyURL := baseURL + "/api/v1/user/auth/verify-email?token=" + token
	emailpkg.SendVerificationEmailAsync(user.Email, user.FirstName, verifyURL)

	return c.JSON(SuccessResponse{Pesan: "Link verifikasi telah dikirim ulang ke email Anda."})
}

// -----------------------------------------------------------------------------
// POST /api/v1/auth/forgot-password
// -----------------------------------------------------------------------------

func (h *AuthHandler) ForgotPassword(c fiber.Ctx) error {
	var req ForgotPasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_BIND", Message: "Request tidak valid"})
	}

	email, ve := utils.ValidateEmail(req.Email)
	if ve != nil {
		return validationError(c, ve)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	user, err := h.Queries.GetUserByEmailBasic(ctx, email)
	if err != nil {
		return c.Status(404).JSON(ErrorResponse{Code: "ERR_EMAIL_NOT_FOUND", Message: "Email tidak terdaftar"})
	}

	if !user.IsActive {
		return c.Status(403).JSON(ErrorResponse{Code: "ERR_ACCOUNT_INACTIVE", Message: "Akun tidak aktif, hubungi admin"})
	}

	if !user.EmailVerified {
		return c.Status(403).JSON(ErrorResponse{Code: "ERR_EMAIL_NOT_VERIFIED", Message: "Email belum diverifikasi. Silakan verifikasi email Anda terlebih dahulu."})
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		slog.Error("auth.forgot_password.token_error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal membuat token"})
	}
	token := hex.EncodeToString(tokenBytes)
	expiry := pgtype.Timestamptz{Time: time.Now().UTC().Add(15 * time.Minute), Valid: true}

	if err := h.Queries.SetPasswordResetToken(ctx, db.SetPasswordResetTokenParams{
		PasswordResetToken:     pgtype.Text{String: token, Valid: true},
		PasswordResetExpiresAt: expiry,
		ID:                     user.ID,
	}); err != nil {
		slog.Error("auth.forgot_password.db_error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal menyimpan token reset"})
	}

	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "3000"
		}
		baseURL = "http://localhost:" + port
	}
	resetURL := baseURL + "/api/v1/user/auth/reset-password-redirect?token=" + token
	emailpkg.SendPasswordResetEmailAsync(user.Email, user.FirstName, resetURL)

	return c.JSON(SuccessResponse{Pesan: "Link reset password telah dikirim ke email Anda. Link berlaku selama 15 menit."})
}

// -----------------------------------------------------------------------------
// GET /api/v1/auth/reset-password-redirect?token=xxx  (public)
// -----------------------------------------------------------------------------

func (h *AuthHandler) ResetPasswordRedirect(c fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION_FAILED", Message: "Token tidak ditemukan"})
	}

	deepLinkBase := os.Getenv("MOBILE_DEEP_LINK_BASE")
	if deepLinkBase == "" {
		return c.Status(200).Type("html").SendString(`<!DOCTYPE html>
<html lang="id">
<head><meta charset="UTF-8"><title>Reset Password</title>
<meta name="viewport" content="width=device-width,initial-scale=1"></head>
<body style="font-family:sans-serif;text-align:center;padding:40px">
<h2>Reset Password</h2>
<p>Silakan buka aplikasi <b>Majadigi</b> untuk melanjutkan reset password.</p>
<p style="color:#6b7280;font-size:14px">Jika Anda belum memiliki aplikasi, unduh terlebih dahulu.</p>
</body></html>`)
	}

	base := strings.TrimSuffix(deepLinkBase, "/")
	redirectURL := base + "/reset-password?token=" + token
	return c.Redirect().To(redirectURL)
}

// -----------------------------------------------------------------------------
// POST /api/v1/auth/reset-password
// -----------------------------------------------------------------------------

func (h *AuthHandler) ResetPassword(c fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_BIND", Message: "Request tidak valid"})
	}

	if req.Token == "" {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION_FAILED", Message: "Token wajib diisi"})
	}

	if ve := utils.ValidatePassword(req.NewPassword); ve != nil {
		return validationError(c, ve)
	}

	if req.NewPassword != req.ConfirmNewPassword {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION_FAILED", Message: "Konfirmasi password tidak cocok"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	user, err := h.Queries.GetUserByPasswordResetToken(ctx, pgtype.Text{String: req.Token, Valid: true})
	if err != nil {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_INVALID_TOKEN", Message: "Token tidak valid atau sudah kadaluarsa"})
	}

	hash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		slog.Error("auth.reset_password.hash_error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal memproses password"})
	}

	if err := h.Queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		PasswordHash: hash,
		ID:           user.ID,
	}); err != nil {
		slog.Error("auth.reset_password.db_error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal memperbarui password"})
	}

	_ = h.Queries.RevokeAllUserTokens(ctx, user.ID)
	invalidateUserCache(uuidToString(user.ID))

	return c.JSON(SuccessResponse{Pesan: "Password berhasil diubah. Silakan login dengan password baru."})
}

// -----------------------------------------------------------------------------
// GET /api/v1/auth/favorites  (requires JWT middleware)
// -----------------------------------------------------------------------------

func (h *AuthHandler) GetFavorites(c fiber.Ctx) error {
	userIDStr, _ := c.Locals("user_id").(string)

	cacheKey := "user:favorites:" + userIDStr
	if respondCached(c, cacheKey) {
		return nil
	}

	userID, err := stringToUUID(userIDStr)
	if err != nil {
		return c.Status(401).JSON(ErrorResponse{Code: "ERR_UNAUTHORIZED", Message: "Token tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	updatedAt, err := h.Queries.GetFavoritesUpdatedAt(ctx, userID)
	if err != nil {
		slog.Error("auth.get_favorites.ts_error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal memuat favorit"})
	}

	rows, err := h.Queries.GetUserFavorites(ctx, userID)
	if err != nil {
		slog.Error("auth.get_favorites.error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal memuat favorit"})
	}

	type favoriteItem struct {
		ServiceID string `json:"service_id"`
		UpdatedAt string `json:"updated_at"`
	}
	items := make([]favoriteItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, favoriteItem{
			ServiceID: uuidToString(r.ServiceID),
			UpdatedAt: r.UpdatedAt.Time.UTC().Format(time.RFC3339),
		})
	}

	res := SuccessResponse{Pesan: "Sukses", Data: fiber.Map{
		"updated_at": updatedAt.Time.UTC().Format(time.RFC3339),
		"favorites":  items,
	}}
	return cacheJSON(c, cacheKey, CacheTTLDetail, res)
}

// -----------------------------------------------------------------------------
// POST /api/v1/auth/favorites  (requires JWT middleware)
// Body: {"service_ids": ["uuid1", "uuid2", ...]}
// -----------------------------------------------------------------------------

func (h *AuthHandler) AddFavorite(c fiber.Ctx) error {
	userIDStr, _ := c.Locals("user_id").(string)

	userID, err := stringToUUID(userIDStr)
	if err != nil {
		return c.Status(401).JSON(ErrorResponse{Code: "ERR_UNAUTHORIZED", Message: "Token tidak valid"})
	}

	var req struct {
		ServiceIDs []string `json:"service_ids"`
	}
	if err := c.Bind().JSON(&req); err != nil || len(req.ServiceIDs) == 0 {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "service_ids tidak boleh kosong"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	for _, idStr := range req.ServiceIDs {
		svcID, err := stringToUUID(idStr)
		if err != nil {
			return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "service_id tidak valid: " + idStr})
		}
		if err := h.Queries.AddUserFavorite(ctx, db.AddUserFavoriteParams{
			UserID:    userID,
			ServiceID: svcID,
		}); err != nil {
			slog.Error("auth.add_favorite.error", slog.String("id", idStr), slog.String("err", err.Error()))
			return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal menambah favorit"})
		}
	}

	_ = h.Queries.BumpFavoritesTimestamp(ctx, userID)
	invalidateUserCache(userIDStr)

	return c.Status(201).JSON(SuccessResponse{Pesan: "Layanan berhasil ditambahkan ke favorit"})
}

// -----------------------------------------------------------------------------
// DELETE /api/v1/auth/favorites  (requires JWT middleware)
// -----------------------------------------------------------------------------

func (h *AuthHandler) RemoveFavorite(c fiber.Ctx) error {
	userIDStr, _ := c.Locals("user_id").(string)

	userID, err := stringToUUID(userIDStr)
	if err != nil {
		return c.Status(401).JSON(ErrorResponse{Code: "ERR_UNAUTHORIZED", Message: "Token tidak valid"})
	}

	var req struct {
		ServiceIDs []string `json:"service_ids"`
	}
	if err := c.Bind().JSON(&req); err != nil || len(req.ServiceIDs) == 0 {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "service_ids tidak boleh kosong"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	for _, idStr := range req.ServiceIDs {
		svcID, err := stringToUUID(idStr)
		if err != nil {
			return c.Status(400).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "service_id tidak valid: " + idStr})
		}
		if err := h.Queries.RemoveUserFavorite(ctx, db.RemoveUserFavoriteParams{
			UserID:    userID,
			ServiceID: svcID,
		}); err != nil {
			slog.Error("auth.remove_favorite.error", slog.String("id", idStr), slog.String("err", err.Error()))
			return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal menghapus favorit"})
		}
	}

	_ = h.Queries.BumpFavoritesTimestamp(ctx, userID)
	invalidateUserCache(userIDStr)

	return c.JSON(SuccessResponse{Pesan: "Layanan berhasil dihapus dari favorit"})
}

// =============================================================================
// Private helpers
// =============================================================================

func (h *AuthHandler) issueTokenPair(ctx context.Context, userID, role string) (access, refresh string, err error) {
	access, err = utils.GenerateAccessToken(userID, role, h.JWTSecret, h.AccessTTL)
	if err != nil {
		return
	}
	refresh, err = utils.GenerateRefreshToken(userID, role, h.JWTSecret, h.RefreshTTL)
	if err != nil {
		return
	}

	refreshHash := utils.HashToken(refresh)
	expiresAt := pgtype.Timestamptz{Time: time.Now().Add(h.RefreshTTL), Valid: true}
	uid, uerr := stringToUUID(userID)
	if uerr != nil {
		err = uerr
		return
	}

	if dbErr := h.Queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    uid,
		TokenHash: refreshHash,
		ExpiresAt: expiresAt,
	}); dbErr != nil {
		slog.Warn("auth.issue_tokens.store_refresh_error", slog.String("err", dbErr.Error()))
	}
	return
}

// seedFavoritesFromCategories queries the gateway DB and adds all services
// belonging to the selected categories into the user's favorites.
func (h *AuthHandler) seedFavoritesFromCategories(userID pgtype.UUID, categoryIDs []string) {
	if len(categoryIDs) == 0 {
		return
	}

	gatewayURL := os.Getenv("GATEWAY_DATABASE_URL")
	if gatewayURL == "" {
		slog.Warn("seed_favorites.no_gateway_url")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, gatewayURL)
	if err != nil {
		slog.Error("seed_favorites.gateway_connect", "err", err.Error())
		return
	}
	defer pool.Close()

	rows, err := pool.Query(ctx, "SELECT DISTINCT service_list_id FROM services_has_categories WHERE category_list_id::text = ANY($1)", categoryIDs)
	if err != nil {
		slog.Error("seed_favorites.query", "err", err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var svcID pgtype.UUID
		if err := rows.Scan(&svcID); err != nil {
			continue
		}
		_ = h.Queries.AddUserFavorite(ctx, db.AddUserFavoriteParams{
			UserID:    userID,
			ServiceID: svcID,
		})
	}
	slog.Info("seed_favorites.done", "user", uuidToString(userID))
}

func (h *AuthHandler) savePreferences(ctx context.Context, userID pgtype.UUID, categoryIDs []string) error {
	if err := h.Queries.DeleteUserPreferences(ctx, userID); err != nil {
		return err
	}
	for _, catIDStr := range categoryIDs {
		catID, err := stringToUUID(catIDStr)
		if err != nil {
			continue
		}
		_ = h.Queries.InsertUserPreference(ctx, db.InsertUserPreferenceParams{
			UserID:     userID,
			CategoryID: catID,
		})
	}
	return nil
}
