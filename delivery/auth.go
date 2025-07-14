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
	RegisterNewUser(ctx context.Context, user entity.User, companyName string) (createdUser *entity.User, company *entity.Company, err error)
	GenerateAccessToken(ctx context.Context, userID int, companyID int, userEmail string, userRole entity.UserRole) (authResponse *entity.AuthResponse, err error)
	UserSignIn(ctx context.Context, email, password string, remember bool) (user *entity.User, company *entity.Company, accessToken string, err error)
	GetUserByID(ctx context.Context, userID int) (targetUser *entity.User, err error)
	GetCompanyByID(ctx context.Context, ID int) (company *entity.Company, err error)
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
	e.GET("/companies", middleware.AuthMiddleware(AllowedAuthenticatedOnly, h.handleGetCompany))
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
	userID := c.Get("user_id").(int)
	ctx := c.Request().Context()

	targetUser, err := h.authService.GetUserByID(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	return c.JSON(http.StatusOK, Response{
		StatusCode: http.StatusOK,
		Message:    fmt.Sprintf("success get profile of %s", targetUser.GetName()),
		Data:       ProfileResponseFromEntity(targetUser),
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

	toCreateUser := entity.User{
		FirstName: requestBody.FirstName,
		LastName:  &requestBody.LastName,
		Email:     requestBody.Email,
		Password:  requestBody.Password,
		Role:      entity.UserRole(requestBody.Role),
	}

	newlyCreatedUser, company, serviceErr := h.authService.RegisterNewUser(ctx, toCreateUser, requestBody.CompanyName)
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

	var companyResponse CompanyResponse
	if requestBody.CompanyName != "" {
		companyResponse = CompanyResponseFromEntity(pointer.Get(company))
	}

	authResponse, err := h.authService.GenerateAccessToken(ctx, newlyCreatedUser.ID, pointer.GetInt(newlyCreatedUser.CompanyID), newlyCreatedUser.Email, newlyCreatedUser.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	return c.JSON(http.StatusCreated, Response{
		StatusCode: http.StatusCreated,
		Message:    fmt.Sprintf("user %s created", requestBody.Email),
		Data: AccessTokenResponse{
			AccessToken: authResponse.AccessToken,
			User: UserResponse{
				ID:       newlyCreatedUser.ID,
				Name:     newlyCreatedUser.GetName(),
				Email:    newlyCreatedUser.Email,
				Phone:    newlyCreatedUser.PhoneNumber,
				JobTitle: newlyCreatedUser.JobTitle,
				Role:     newlyCreatedUser.Role,
			},
			Company: &companyResponse,
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

	user, company, accessToken, serviceErr := h.authService.UserSignIn(ctx, requestBody.Email, requestBody.Password, requestBody.Remember)
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

	var companyResponse CompanyResponse
	if company != nil {
		companyResponse = CompanyResponseFromEntity(pointer.Get(company))
	}

	return c.JSON(http.StatusOK, Response{
		StatusCode: http.StatusOK,
		Message:    "success",
		Data: AccessTokenResponse{
			AccessToken: accessToken,
			User: UserResponse{
				ID:       user.ID,
				Name:     user.GetName(),
				Email:    user.Email,
				Phone:    user.PhoneNumber,
				JobTitle: user.JobTitle,
				Role:     user.Role,
			},
			Company: &companyResponse,
		},
	})
}

func (h *AuthHandler) handleGetCompany(c echo.Context) error {
	ctx := c.Request().Context()

	companyID := c.Get("company_id").(int)

	company, err := h.authService.GetCompanyByID(ctx, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	return c.JSON(http.StatusOK, Response{
		StatusCode: http.StatusOK,
		Message:    "ok",
		Data:       CompanyResponseFromEntity(pointer.Get(company)),
		Error:      nil,
	})

}
