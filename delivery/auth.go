package delivery

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	"github.com/mhdiiilham/gosm/entity"
	"github.com/mhdiiilham/gosm/logger"
)

// AuthService defines authentication-related operations.
type AuthService interface {
	RegisterNewUser(ctx context.Context, user entity.User) (createdUser *entity.User, err error)
	GenerateAccessToken(ctx context.Context, user entity.User) (accessToken string, err error)
	UserSignIn(ctx context.Context, email, password string) (accessToken string, err error)
}

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authService AuthService
}

// NewAuthHandler creates a new instance of AuthHandler.
func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterAuthRoutes registers authentication-related endpoints to the provided echo group.
func (h *AuthHandler) RegisterAuthRoutes(e *echo.Group) {
	e.POST("/", h.handleSignIn)
	e.POST("/signup", h.handleSignUp)
}

// handleSignUp processes user registration requests.
func (h *AuthHandler) handleSignUp(c echo.Context) error {
	ctx := c.Request().Context()
	const ops = "AuthHandler.handleSignUp"
	var requestBody SignUpRequest

	if err := c.Bind(&requestBody); err != nil {
		logger.Warn(ctx, ops, "failed to parse request body")
		return c.JSON(http.StatusInternalServerError, Response{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal Server Error",
			Data:       nil,
			Error:      err,
		})
	}

	toCreateUser := entity.User{
		FirstName:   requestBody.FirstName,
		LastName:    &requestBody.LastName,
		Email:       requestBody.Email,
		CountryCode: pointer.To("+62"), // Hardcoded this cause it's only Indo, lol
		PhoneNumber: &requestBody.PhoneNumber,
		Password:    requestBody.Password,
		Role:        entity.UserRole(requestBody.Role),
	}

	newlyCreatedUser, serviceErr := h.authService.RegisterNewUser(ctx, toCreateUser)
	if serviceErr != nil {
		switch err := serviceErr.(type) {
		case entity.GosmError:
			if err.Type == entity.GosmErrorTypeBadRequest {
				return c.JSON(http.StatusBadRequest, Response{
					StatusCode: http.StatusBadRequest,
					Message:    err.Message,
					Data:       nil,
					Error:      err.Source,
				})
			}
		}

		return c.JSON(http.StatusInternalServerError, throwInternalServerError(serviceErr))
	}

	accessToken, err := h.authService.GenerateAccessToken(ctx, *newlyCreatedUser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	return c.JSON(http.StatusCreated, Response{
		StatusCode: http.StatusCreated,
		Message:    fmt.Sprintf("user %s created", requestBody.FirstName),
		Data: AccessTokenResponse{
			Email:       requestBody.Email,
			AccessToken: accessToken,
		},
	})
}

// handleSignIn processes user signIn requests.
func (h *AuthHandler) handleSignIn(c echo.Context) error {
	ctx := c.Request().Context()
	const ops = "AuthHandler.handleSignIn"
	var requestBody SignInRequest

	if err := c.Bind(&requestBody); err != nil {
		logger.Warn(ctx, ops, "failed to parse request body")
		return c.JSON(http.StatusInternalServerError, Response{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal Server Error",
			Data:       nil,
			Error:      err,
		})
	}

	accessToken, serviceErr := h.authService.UserSignIn(ctx, requestBody.Email, requestBody.Password)
	if serviceErr != nil {
		logger.Errorf(ctx, ops, "user sign in fails: %v", serviceErr)
		switch err := serviceErr.(type) {
		case entity.GosmError:
			if err.Type == entity.GosmErrorTypeBadRequest {
				return c.JSON(http.StatusBadRequest, Response{
					StatusCode: http.StatusBadRequest,
					Message:    err.Message,
					Data:       nil,
					Error:      err.Source,
				})
			}
		}

		return c.JSON(http.StatusInternalServerError, throwInternalServerError(serviceErr))
	}

	return c.JSON(http.StatusOK, Response{
		StatusCode: http.StatusOK,
		Message:    "success",
		Data: AccessTokenResponse{
			Email:       requestBody.Email,
			AccessToken: accessToken,
		},
	})
}
