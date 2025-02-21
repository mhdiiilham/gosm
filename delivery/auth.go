package delivery

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	"github.com/mhdiiilham/gosm/entity"
	"github.com/mhdiiilham/gosm/logger"
)

// AuthService defines authentication-related operations.
type AuthService interface {
	RegisterNewUser(ctx context.Context, user entity.User) (createdUser *entity.User, err error)
	GenerateAccessToken(ctx context.Context, user entity.User, duration time.Duration) (authResponse *entity.AuthResponse, err error)
	UserSignIn(ctx context.Context, email, password string, remember bool) (authResponse *entity.AuthResponse, err error)
	GetUserByID(ctx context.Context, userID string) (targetUser *entity.User, err error)
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
func (h *AuthHandler) RegisterAuthRoutes(e *echo.Group, middleware *Middleware) {
	e.POST("", h.HandleSignIn)
	e.POST("/signup", h.HandleSignUp)
	e.GET("/profile", middleware.AuthMiddleware(AllowedAuthenticatedOnly, h.HandleProfile))
}

// HandleProfile godoc
//
//	@Summary	get logged user's profile.
//	@Tags		auth
//	@Accept		json
//	@Product	json
//	@Success	200	{object}	Response{data=entity.User}
//	@Failure	500	{object}	Response
//	@Router		/api/v1/auth/profile [get]
func (h *AuthHandler) HandleProfile(c echo.Context) error {
	userID := c.Get("user_id").(string)
	ctx := c.Request().Context()

	targetUser, err := h.authService.GetUserByID(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	return c.JSON(http.StatusOK, Response{
		StatusCode: http.StatusOK,
		Message:    fmt.Sprintf("success get profile of %s", targetUser.GetName()),
		Data:       targetUser,
		Error:      nil,
	})
}

// HandleSignUp godoc
//
//	@Summary	Register a new user
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		request	body		SignInRequest	true	"User sign-up credentials"
//	@Success	201		{object}	Response{data=AccessTokenResponse}
//	@Failure	400		{object}	Response
//	@Failure	500		{object}	Response
//	@Router		/api/v1/auth/signup [post]
func (h *AuthHandler) HandleSignUp(c echo.Context) error {
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

	fmt.Printf("payload: %+v\n", requestBody)
	countryCode, localNumber, _ := entity.ParsePhoneNumber(requestBody.PhoneNumber)

	toCreateUser := entity.User{
		FirstName:   requestBody.FirstName,
		LastName:    &requestBody.LastName,
		Email:       requestBody.Email,
		CountryCode: pointer.To(countryCode), // Hardcoded this cause it's only Indo, lol
		PhoneNumber: pointer.To(localNumber),
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

	authResponse, err := h.authService.GenerateAccessToken(ctx, *newlyCreatedUser, 12*time.Hour)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	return c.JSON(http.StatusCreated, Response{
		StatusCode: http.StatusCreated,
		Message:    fmt.Sprintf("user %s created", requestBody.FirstName),
		Data: AccessTokenResponse{
			AccessToken: authResponse.AccessToken,
			ExpiresAt:   authResponse.ExpiresAt,
			Email:       authResponse.Email,
			Role:        authResponse.Role,
		},
	})
}

// HandleSignIn godoc
//
//	@Summary	Authenticate user and return access token
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		request	body		SignInRequest						true	"User sign-in credentials"
//	@Success	200		{object}	Response{data=AccessTokenResponse}	"User successfully authenticated"
//	@Failure	400		{object}	Response							"Invalid credentials or bad request"
//	@Failure	500		{object}	Response							"Internal server error"
//	@Router		/api/v1/auth [post]
func (h *AuthHandler) HandleSignIn(c echo.Context) error {
	ctx := c.Request().Context()
	const ops = "AuthHandler.HandleSignIn"
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

	authResponse, serviceErr := h.authService.UserSignIn(ctx, requestBody.Email, requestBody.Password, requestBody.Remember)
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
			AccessToken: authResponse.AccessToken,
			ExpiresAt:   authResponse.ExpiresAt,
			Email:       authResponse.Email,
			Role:        authResponse.Role,
		},
	})
}
