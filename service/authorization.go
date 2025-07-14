package service

import (
	"context"
	"net/mail"

	"github.com/AlekSi/pointer"
	"github.com/mhdiiilham/gosm/entity"
	"github.com/mhdiiilham/gosm/logger"
	"github.com/mhdiiilham/gosm/pkg"
)

// UserRepository defines an interface for user-related database operations.
type UserRepository interface {
	CreateUser(ctx context.Context, newUser entity.User, companyID *int) (createdUser *entity.User, err error)
	FindByEmail(ct context.Context, email string) (existingUser *entity.User, err error)
	GetUserByID(ctx context.Context, userID int) (targetUser *entity.User, err error)
}

// CompanyRepository defines an interface for company-related database operations.
type CompanyRepository interface {
	CreateCompany(ctx context.Context, companyName string) (*entity.Company, error)
	FindByID(ctx context.Context, ID int) (*entity.Company, error)
}

// PasswordHasher defines an interface for handling password hashing and comparison.
type PasswordHasher interface {
	HashPassword(plainPassword string) (hashedPassword string, err error)
	ComparePassword(password, hashedPassword string) (passwordIsValid bool)
}

// JwtGenerator defines an interface for handling JWT operations, including token creation and parsing.
type JwtGenerator interface {
	CreateAccessToken(userID int, companyID int, email string, userRole entity.UserRole) (response *entity.AuthResponse, err error)
	ParseToken(accessToken string) (*pkg.TokenClaims, error)
}

// Authenticator struct provides authentication and authorization-related operations.
type Authenticator struct {
	userRepository    UserRepository
	companyRepository CompanyRepository
	passwordHasher    PasswordHasher
	jwtGenerator      JwtGenerator
}

// NewAuthorizationService initializes and returns an instance of Authenticator.
func NewAuthorizationService(userRepository UserRepository, companyRepository CompanyRepository, passwordHasher PasswordHasher, jwtGenerator JwtGenerator) *Authenticator {
	return &Authenticator{
		userRepository:    userRepository,
		companyRepository: companyRepository,
		passwordHasher:    passwordHasher,
		jwtGenerator:      jwtGenerator,
	}
}

// RegisterNewUser handles the registration of a new user.
func (a *Authenticator) RegisterNewUser(ctx context.Context, user entity.User, companyName string) (createdUser *entity.User, company *entity.Company, err error) {
	const ops = "Authenticator.RegisterNewUser"

	if _, err := mail.ParseAddress(user.Email); err != nil {
		return nil, nil, entity.ErrUserInvalidEmailAddress
	}

	if user.Role == "" {
		return nil, nil, entity.ErrUserRoleIsEmpty
	}

	if user.FirstName == "" {
		return nil, nil, entity.ErrUserFirstNameEmpty
	}

	if user.Password == "" {
		return nil, nil, entity.ErrUserPasswordEmpty
	}

	var companyID *int
	newlyCreatedCompany, err := a.companyRepository.CreateCompany(ctx, companyName)
	if err != nil {
		logger.Errorf(ctx, ops, "failed to hash insert company: %v", err)
		return nil, nil, entity.UnknownError(err)
	}

	hashedPassword, err := a.passwordHasher.HashPassword(user.Password)
	if err != nil {
		logger.Errorf(ctx, ops, "failed to hash plain password: %v", err)
		return nil, nil, entity.UnknownError(err)
	}

	if newlyCreatedCompany != nil {
		companyID = &newlyCreatedCompany.ID
	}

	user.Password = hashedPassword
	createdUser, err = a.userRepository.CreateUser(ctx, user, companyID)
	if err != nil {
		return nil, nil, entity.UnknownError(err)
	}

	return createdUser, newlyCreatedCompany, nil
}

// GenerateAccessToken generates a JWT access token for the given user.
// This function takes a user entity and uses the JWT generator to create a signed access token.
// If the token generation fails, it logs the error and returns a structured application error.
func (a *Authenticator) GenerateAccessToken(ctx context.Context, userID int, companyID int, userEmail string, userRole entity.UserRole) (authResponse *entity.AuthResponse, err error) {
	const ops = "Authenticator.GenerateAccessToken"

	// Generate an access token using the JWT generator
	authResponse, err = a.jwtGenerator.CreateAccessToken(userID, companyID, userEmail, userRole)
	if err != nil {
		logger.Errorf(ctx, ops, "failed to generate user access token: %v", err)
		return nil, entity.UnknownError(err)
	}

	return authResponse, nil
}

// UserSignIn handles user authentication by validating the provided email and password.
// It returns an access token upon successful authentication.
func (a *Authenticator) UserSignIn(ctx context.Context, email, password string, remember bool) (user *entity.User, company *entity.Company, accessToken string, err error) {
	const ops = "Authenticator.UserSignIn"

	if email == "" || password == "" {
		return nil, nil, "", entity.ErrInvalidSignInPayload
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return nil, nil, "", entity.ErrUserInvalidEmailAddress
	}

	user, err = a.userRepository.FindByEmail(ctx, email)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil, "", entity.ErrInvalidSignInPayload
		}

		logger.Errorf(ctx, ops, "failed to retrieve user: %v", err)
		return nil, nil, "", entity.UnknownError(err)
	}

	if !a.passwordHasher.ComparePassword(password, user.Password) {
		return nil, nil, "", entity.ErrInvalidSignInPayload
	}

	if user.CompanyID != nil {
		company, err = a.companyRepository.FindByID(ctx, user.GetCompanyID())
		if err != nil {
			logger.Errorf(ctx, ops, "failed to retrieve user: %v", err)
			return nil, nil, "", entity.UnknownError(err)
		}
	}

	authResponse, err := a.GenerateAccessToken(ctx, user.ID, pointer.GetInt(user.CompanyID), user.Email, user.Role)
	if err != nil {
		logger.Errorf(ctx, ops, "failed to generate accessToken: %v", err)
		return nil, nil, "", entity.UnknownError(err)
	}

	return user, company, authResponse.AccessToken, nil
}

// GetUserByID ...
func (a *Authenticator) GetUserByID(ctx context.Context, userID int) (targetUser *entity.User, err error) {
	return a.userRepository.GetUserByID(ctx, userID)
}

// GetCompanyByID ...
func (a *Authenticator) GetCompanyByID(ctx context.Context, ID int) (company *entity.Company, err error) {
	return a.companyRepository.FindByID(ctx, ID)
}
