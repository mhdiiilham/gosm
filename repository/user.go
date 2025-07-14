package repository

import (
	"context"
	"database/sql"

	"github.com/mhdiiilham/gosm/entity"
	"github.com/mhdiiilham/gosm/logger"
)

// UserRepository provides an abstraction layer for database operations related to users.
// It encapsulates the database connection and offers methods to interact with the user data.
type UserRepository struct {
	db *sql.DB // Database connection instance
}

// NewUserRepository initializes a new instance of UserRepository.
// It accepts a database connection and returns a pointer to the UserRepository struct.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser inserts a new user record into the database.
// This function takes a `newUser` entity containing user details and inserts it into the database.
// It returns the created user with a generated ID or an error if the operation fails.
func (r *UserRepository) CreateUser(ctx context.Context, newUser entity.User, companyID *int) (createdUser *entity.User, err error) {
	const ops = "UserRepository.CreateUser"

	row := r.db.QueryRowContext(
		ctx, SQLStatementInsertUser,
		newUser.FirstName,
		newUser.LastName,
		newUser.Role,
		newUser.Email,
		newUser.Password,
		newUser.PhoneNumber,
		companyID,
	)

	if err := row.Scan(&newUser.ID); err != nil {
		logger.Errorf(ctx, ops, "failed to insert new user: %v", err)
		return nil, err
	}

	return &newUser, nil
}

// FindByEmail retrieves a user from the database based on their email address.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	const ops = "UserRepository.FindByEmail"
	var existingUser entity.User

	row := r.db.QueryRowContext(ctx, SQLStatementSelectUserByEmail, email)
	if err := row.Scan(
		&existingUser.ID,
		&existingUser.FirstName,
		&existingUser.LastName,
		&existingUser.Role,
		&existingUser.Email,
		&existingUser.Password,
		&existingUser.PhoneNumber,
		&existingUser.CompanyID,
	); err != nil {
		return nil, err
	}

	return &existingUser, nil
}

// GetUserByID retrieves a user from the database based on their id.
func (r *UserRepository) GetUserByID(ctx context.Context, userID int) (targetUser *entity.User, err error) {
	targetUser = &entity.User{}

	logger.Infof(ctx, "UserRepository.GetUserByID", "query user with id=%d", userID)

	row := r.db.QueryRowContext(ctx, SQLStatementSelectUserByID, userID)
	if err := row.Scan(
		&targetUser.ID,
		&targetUser.FirstName,
		&targetUser.LastName,
		&targetUser.Role,
		&targetUser.Email,
		&targetUser.Password,
		&targetUser.PhoneNumber,
		&targetUser.CompanyID,
	); err != nil {
		return nil, err
	}

	return targetUser, nil
}
