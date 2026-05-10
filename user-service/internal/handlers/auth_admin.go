package handlers

import (
	"context"
	"errors"
	"log/slog"

	"github.com/farildzaky/user-service/internal/db"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

// -----------------------------------------------------------------------------
// GET /api/v1/admin/users
// -----------------------------------------------------------------------------

func (h *AuthHandler) ListUsersAdmin(c fiber.Ctx) error {
	page, limit, offset := parsePagination(c)

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	rows, err := h.Queries.ListUsers(ctx, db.ListUsersParams{
		LimitData:  int32(limit),
		OffsetData: int32(offset),
	})
	if err != nil {
		slog.Error("admin.list_users.error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal memuat daftar pengguna"})
	}

	total, err := h.Queries.CountUsers(ctx)
	if err != nil {
		slog.Error("admin.count_users.error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal menghitung pengguna"})
	}

	items := make([]UserListItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, UserListItem{
			ID:        uuidToString(r.ID),
			FirstName: r.FirstName,
			LastName:  r.LastName,
			Phone:     r.Phone,
			Email:     r.Email,
			NIK:       r.Nik,
			Role:      string(r.Role),
			IsActive:  r.IsActive,
			CreatedAt: r.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return c.JSON(SuccessResponse{
		Pesan: "Sukses",
		Data:  items,
		Pagination: &PaginationMeta{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

// -----------------------------------------------------------------------------
// GET /api/v1/admin/users/:id
// -----------------------------------------------------------------------------

func (h *AuthHandler) GetUserByIDAdmin(c fiber.Ctx) error {
	userID, err := stringToUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_INVALID_UUID", Message: "UUID tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	user, err := h.Queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(404).JSON(ErrorResponse{Code: "ERR_NOT_FOUND", Message: "Pengguna tidak ditemukan"})
		}
		slog.Error("admin.get_user.error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal memuat pengguna"})
	}

	return c.JSON(SuccessResponse{Pesan: "Sukses", Data: toUserProfileFromIDRow(user)})
}

// -----------------------------------------------------------------------------
// PUT /api/v1/admin/users/:id/role  (superadmin only)
// -----------------------------------------------------------------------------

func (h *AuthHandler) UpdateUserRoleAdmin(c fiber.Ctx) error {
	userID, err := stringToUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_INVALID_UUID", Message: "UUID tidak valid"})
	}

	var req UpdateRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_BIND", Message: "Request tidak valid"})
	}

	validRoles := map[string]bool{"user": true, "admin": true, "superadmin": true}
	if !validRoles[req.Role] {
		return c.Status(400).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION_FAILED",
			Message: "Role tidak valid (user | admin | superadmin)",
		})
	}

	// Prevent downgrading your own superadmin account
	callerID, _ := c.Locals("user_id").(string)
	if callerID == uuidToString(userID) && req.Role != "superadmin" {
		return c.Status(403).JSON(ErrorResponse{
			Code:    "ERR_FORBIDDEN",
			Message: "Tidak dapat mengubah role akun sendiri",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	row, err := h.Queries.UpdateUserRole(ctx, db.UpdateUserRoleParams{
		ID:   userID,
		Role: db.UserRole(req.Role),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(404).JSON(ErrorResponse{Code: "ERR_NOT_FOUND", Message: "Pengguna tidak ditemukan"})
		}
		slog.Error("admin.update_role.error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal memperbarui role"})
	}

	slog.Info("admin.update_role.success",
		slog.String("target_user", uuidToString(userID)),
		slog.String("new_role", req.Role),
		slog.String("by", callerID),
	)

	return c.JSON(SuccessResponse{
		Pesan: "Role berhasil diperbarui",
		Data: UserRoleResponse{
			ID:        uuidToString(row.ID),
			FirstName: row.FirstName,
			LastName:  row.LastName,
			Email:     row.Email,
			Role:      string(row.Role),
			IsActive:  row.IsActive,
			UpdatedAt: row.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}

// -----------------------------------------------------------------------------
// DELETE /api/v1/admin/users/:id  (deactivate, not hard delete)
// -----------------------------------------------------------------------------

func (h *AuthHandler) DeactivateUserAdmin(c fiber.Ctx) error {
	userID, err := stringToUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(ErrorResponse{Code: "ERR_INVALID_UUID", Message: "UUID tidak valid"})
	}

	callerID, _ := c.Locals("user_id").(string)
	if callerID == uuidToString(userID) {
		return c.Status(403).JSON(ErrorResponse{
			Code:    "ERR_FORBIDDEN",
			Message: "Tidak dapat menonaktifkan akun sendiri",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeactivateUser(ctx, userID); err != nil {
		slog.Error("admin.deactivate_user.error", slog.String("err", err.Error()))
		return c.Status(500).JSON(ErrorResponse{Code: "ERR_INTERNAL", Message: "Gagal menonaktifkan pengguna"})
	}

	_ = h.Queries.RevokeAllUserTokens(ctx, userID)
	invalidateUserCache(uuidToString(userID))

	slog.Info("admin.deactivate_user.success",
		slog.String("target_user", uuidToString(userID)),
		slog.String("by", callerID),
	)

	return c.JSON(SuccessResponse{Pesan: "Pengguna berhasil dinonaktifkan"})
}
