package service

import (
	"context"
	"net/mail"
	"time"

	"github.com/mhdiiilham/gosm/entity"
	"github.com/mhdiiilham/gosm/logger"
	"github.com/mhdiiilham/gosm/pkg"
)

// UserRepository defines an interface for user-related database operations.
type UserRepository interface {
	CreateUser(ctx context.Context, newUser entity.User) (createdUser *entity.User, err error)
	FindByEmail(ct context.Context, email string) (existingUser *entity.User, err error)
}

// PasswordHasher defines an interface for handling password hashing and comparison.
type PasswordHasher interface {
	HashPassword(plainPassword string) (hashedPassword string, err error)
	ComparePassword(password, hashedPassword string) (passwordIsValid bool)
}

// JwtGenerator defines an interface for handling JWT operations, including token creation and parsing.
type JwtGenerator interface {
	CreateAccessToken(userID, email string, userRole entity.UserRole, duration time.Duration) (accessToken, expireAt string, err error)
	ParseToken(accessToken string) (*pkg.TokenClaims, error)
}

// Authenticator struct provides authentication and authorization-related operations.
type Authenticator struct {
	userRepository UserRepository
	passwordHasher PasswordHasher
	jwtGenerator   JwtGenerator
}

// NewAuthorizationService initializes and returns an instance of Authenticator.
func NewAuthorizationService(userRepository UserRepository, passwordHasher PasswordHasher, jwtGenerator JwtGenerator) *Authenticator {
	return &Authenticator{
		userRepository: userRepository,
		passwordHasher: passwordHasher,
		jwtGenerator:   jwtGenerator,
	}
}

// RegisterNewUser handles the registration of a new user.
func (a *Authenticator) RegisterNewUser(ctx context.Context, user entity.User) (createdUser *entity.User, err error) {
	const ops = "Authenticator.RegisterNewUser"

	if _, err := mail.ParseAddress(user.Email); err != nil {
		return nil, entity.ErrUserInvalidEmailAddress
	}

	if user.Role == "" {
		return nil, entity.ErrUserRoleIsEmpty
	}

	if user.FirstName == "" {
		return nil, entity.ErrUserFirstNameEmpty
	}

	if user.Password == "" {
		return nil, entity.ErrUserPasswordEmpty
	}

	// Other than guest, we have to check whether users' email exist or not.
	if user.Role != entity.UserRoleGuest {
		if _, err := a.userRepository.FindByEmail(ctx, user.Email); err == nil {
			return nil, entity.ErrUserExisted
		}
	}

	hashedPassword, err := a.passwordHasher.HashPassword(user.Password)
	if err != nil {
		logger.Errorf(ctx, ops, "failed to hash plain password: %v", err)
		return nil, entity.UnknownError(err)
	}

	user.Password = hashedPassword
	createdUser, err = a.userRepository.CreateUser(ctx, user)
	if err != nil {
		return nil, entity.UnknownError(err)
	}

	return createdUser, nil
}

// GenerateAccessToken generates a JWT access token for the given user.
// This function takes a user entity and uses the JWT generator to create a signed access token.
// If the token generation fails, it logs the error and returns a structured application error.
func (a *Authenticator) GenerateAccessToken(ctx context.Context, user entity.User, duration time.Duration) (accessToken, expireAt string, err error) {
	const ops = "Authenticator.GenerateAccessToken"

	// Generate an access token using the JWT generator
	accessToken, expireAt, err = a.jwtGenerator.CreateAccessToken(user.ID, user.Email, user.Role, duration)
	if err != nil {
		logger.Errorf(ctx, ops, "failed to generate user access token: %v", err)
		return accessToken, expireAt, entity.UnknownError(err)
	}

	return accessToken, expireAt, nil
}

// UserSignIn handles user authentication by validating the provided email and password.
// It returns an access token upon successful authentication.
func (a *Authenticator) UserSignIn(ctx context.Context, email, password string, remember bool) (accessToken, expireAt string, err error) {
	const ops = "Authenticator.UserSignIn"

	if email == "" || password == "" {
		return "", "", entity.ErrInvalidSignInPayload
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return "", "", entity.ErrUserInvalidEmailAddress
	}

	user, err := a.userRepository.FindByEmail(ctx, email)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return "", "", entity.ErrInvalidSignInPayload
		}

		logger.Errorf(ctx, ops, "failed to retrieve user: %v", err)
		return "", "", entity.UnknownError(err)
	}

	if !a.passwordHasher.ComparePassword(password, user.Password) {
		return "", "", entity.ErrInvalidSignInPayload
	}

	duration := 12 * time.Hour
	if remember {
		duration = 168 * time.Hour // one week
	}

	return a.GenerateAccessToken(ctx, *user, duration)
}
