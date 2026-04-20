package auth

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

const UserIDKey = "userID"

func JWTMiddleware(jwtManager *JWTManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}
			claims, err := jwtManager.Verify(parts[1])
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}
			c.Set(UserIDKey, claims.UserID)
			return next(c)
		}
	}
}

func GetUserID(c echo.Context) string {
	return c.Get(UserIDKey).(string)
}
