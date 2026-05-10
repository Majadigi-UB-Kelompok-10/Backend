package middleware

import (
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gofiber/fiber/v3"
)

type GatewayClaims struct {
	UserID string `json:"uid"`
	Role   string `json:"role"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

type errResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RequireAuth validates a Bearer JWT and sets user_id + user_role in ctx.Locals.
func RequireAuth() fiber.Handler {
	secret := os.Getenv("JWT_SECRET")
	return func(c fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(errResp{
				Code:    "ERR_UNAUTHORIZED",
				Message: "Authorization header diperlukan (Bearer <token>)",
			})
		}

		tokenStr := header[7:]
		token, err := jwt.ParseWithClaims(tokenStr, &GatewayClaims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(errResp{
				Code:    "ERR_INVALID_TOKEN",
				Message: "Token tidak valid atau sudah kadaluarsa",
			})
		}

		claims, ok := token.Claims.(*GatewayClaims)
		if !ok || claims.Type != "access" {
			return c.Status(fiber.StatusUnauthorized).JSON(errResp{
				Code:    "ERR_INVALID_TOKEN",
				Message: "Tipe token tidak valid",
			})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("user_role", claims.Role)
		return c.Next()
	}
}

// RequireAdmin allows only admin and superadmin roles (must be used after RequireAuth).
func RequireAdmin(c fiber.Ctx) error {
	role, _ := c.Locals("user_role").(string)
	if role != "admin" && role != "superadmin" {
		return c.Status(fiber.StatusForbidden).JSON(errResp{
			Code:    "ERR_FORBIDDEN",
			Message: "Akses hanya untuk admin",
		})
	}
	return c.Next()
}
