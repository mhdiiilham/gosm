package delivery

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mhdiiilham/gosm/entity"
	"github.com/mhdiiilham/gosm/pkg"
)

// JwtGenerator defines an interface for handling JWT operations, including token creation and parsing.
type JwtGenerator interface {
	CreateAccessToken(userID, email string, userRole entity.UserRole, duration time.Duration) (response *entity.AuthResponse, err error)
	ParseToken(accessToken string) (*pkg.TokenClaims, error)
}

// UserRepository defines an interface for user-related database operations.
type UserRepository interface {
	FindByEmail(ct context.Context, email string) (existingUser *entity.User, err error)
}

// Middleware provides authentication-related middleware functions.
type Middleware struct {
	jwtService     JwtGenerator
	userRepository UserRepository
}

// NewMiddleware initializes a new Middleware instance with the provided JWT service and user repository.
func NewMiddleware(jwtService JwtGenerator, userRepository UserRepository) *Middleware {
	return &Middleware{jwtService: jwtService, userRepository: userRepository}
}

// AuthMiddleware is a middleware function that handles authentication and authorization.
// It verifies the JWT token from the Authorization header and checks if the user has the required role.
func (m *Middleware) AuthMiddleware(allowedRoles []entity.UserRole, next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		authHeader = strings.ReplaceAll(authHeader, "Bearer ", "")

		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, Response{StatusCode: http.StatusUnauthorized, Message: "Request could not be authorised"})
		}

		claims, err := m.jwtService.ParseToken(authHeader)
		if err != nil {
			switch parsedErr := err.(type) {
			case entity.GosmError:
				if parsedErr.Code == "AUTH_TOKEN_EXPIRED" {
					return c.JSON(http.StatusUnauthorized, Response{
						StatusCode: http.StatusUnauthorized,
						Message:    parsedErr.Message,
						Data:       parsedErr.Code,
						Error:      nil,
					})
				}
			}
			return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
		}

		ctx := c.Request().Context()
		email := claims.Email
		role := claims.Role

		if _, err := m.userRepository.FindByEmail(ctx, email); err != nil {
			return c.JSON(http.StatusUnauthorized, Response{StatusCode: http.StatusUnauthorized, Message: "Request could not be authorised"})
		}

		if !slices.Contains(allowedRoles, role) {
			return c.JSON(http.StatusUnauthorized, Response{StatusCode: http.StatusUnauthorized, Message: "Request could not be authorised"})
		}

		c.Set("user_id", claims.ID)
		c.Set("user_email", claims.Email)

		return next(c)
	}
}
