package routes

import (
	"strings"
	"time"

	"github.com/farildzaky/user-service/internal/handlers"
	"github.com/farildzaky/user-service/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func SetupRoutes(app *fiber.App, h *handlers.AuthHandler, jwtSecret string) {
	api := app.Group("/api/v1")

	authLimiter := limiter.New(limiter.Config{
		Max:        20,
		Expiration: 1 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(handlers.ErrorResponse{
				Code:    "ERR_RATE_LIMIT",
				Message: "Terlalu banyak percobaan, silakan tunggu 1 menit",
			})
		},
	})

	api.Get("/", func(c fiber.Ctx) error {
		return c.SendString("User Service API v1 Active")
	})

	// ---------------------------------------------------------------------------
	// Public auth routes
	// ---------------------------------------------------------------------------
	auth := api.Group("/auth")
	auth.Post("/register", authLimiter, h.Register)
	auth.Post("/login", authLimiter, h.Login)
	auth.Post("/refresh", h.RefreshToken)
	auth.Post("/logout", h.Logout)
	auth.Get("/verify-email", h.VerifyEmail)

	// Protected auth routes (require valid access token)
	me := auth.Group("/", requireAuth(jwtSecret))
	me.Get("/me", h.GetMe)
	me.Put("/me", h.UpdateMe)
	me.Get("/preferences", h.GetPreferences)
	me.Post("/preferences", h.SetPreferences)
	me.Get("/favorites", h.GetFavorites)
	me.Post("/favorites/:service_id", h.AddFavorite)
	me.Delete("/favorites/:service_id", h.RemoveFavorite)

	// ---------------------------------------------------------------------------
	// Admin routes (require admin or superadmin role)
	// ---------------------------------------------------------------------------
	admin := api.Group("/admin", requireAuth(jwtSecret), requireAdmin)
	adminUsers := admin.Group("/users")
	adminUsers.Get("/", h.ListUsersAdmin)
	adminUsers.Get("/:id", h.GetUserByIDAdmin)
	adminUsers.Delete("/:id", h.DeactivateUserAdmin)

	// Role update — superadmin only
	adminUsers.Put("/:id/role", requireSuperadmin, h.UpdateUserRoleAdmin)
}

// =============================================================================
// JWT Middleware
// =============================================================================

func requireAuth(jwtSecret string) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(handlers.ErrorResponse{
				Code:    "ERR_UNAUTHORIZED",
				Message: "Authorization header diperlukan (Bearer <token>)",
			})
		}

		tokenStr := header[7:]
		claims, err := utils.ValidateToken(tokenStr, jwtSecret)
		if err != nil || claims.Type != utils.TokenTypeAccess {
			return c.Status(fiber.StatusUnauthorized).JSON(handlers.ErrorResponse{
				Code:    "ERR_INVALID_TOKEN",
				Message: "Token tidak valid atau sudah kadaluarsa",
			})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("user_role", claims.Role)
		return c.Next()
	}
}

func requireAdmin(c fiber.Ctx) error {
	role, _ := c.Locals("user_role").(string)
	if role != "admin" && role != "superadmin" {
		return c.Status(fiber.StatusForbidden).JSON(handlers.ErrorResponse{
			Code:    "ERR_FORBIDDEN",
			Message: "Akses hanya untuk admin",
		})
	}
	return c.Next()
}

func requireSuperadmin(c fiber.Ctx) error {
	role, _ := c.Locals("user_role").(string)
	if role != "superadmin" {
		return c.Status(fiber.StatusForbidden).JSON(handlers.ErrorResponse{
			Code:    "ERR_FORBIDDEN",
			Message: "Akses hanya untuk superadmin",
		})
	}
	return c.Next()
}
