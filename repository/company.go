package repository

import (
	"context"
	"database/sql"

	"github.com/mhdiiilham/gosm/entity"
	"github.com/mhdiiilham/gosm/logger"
)

// CompanyRepository provides an abstraction layer for database operations related to users.
// It encapsulates the database connection and offers methods to interact with the user data.
type CompanyRepository struct {
	db *sql.DB
}

// NewCompanyRepository initializes a new instance of UserRepository.
// It accepts a database connection and returns a pointer to the UserRepository struct.
func NewCompanyRepository(db *sql.DB) *CompanyRepository {
	return &CompanyRepository{db: db}
}

// CreateCompany inserts a new company record into the database.
// It returns the created user with a generated ID or an error if the operation fails.
func (r *CompanyRepository) CreateCompany(ctx context.Context, companyName string) (*entity.Company, error) {
	const ops = "CompanyRepository.CreateCompany"
	row := r.db.QueryRowContext(
		ctx,
		SQLInsertCompany,
		companyName,
	)

	var createdID int
	if err := row.Scan(&createdID); err != nil {
		logger.Errorf(ctx, ops, "failed to insert new company: %v", err)
		return nil, err
	}

	return &entity.Company{
		ID:   createdID,
		Name: companyName,
	}, nil
}

// FindByID ...
func (r *CompanyRepository) FindByID(ctx context.Context, companyID int) (*entity.Company, error) {
	const ops = "CompanyRepository.FindByID"

	var company entity.Company
	row := r.db.QueryRowContext(
		ctx,
		SQLSelectCompany,
		companyID,
	)

	if err := row.Scan(
		&company.ID,
		&company.Name,
		&company.Address,
		&company.LogoURL,
		&company.Website,
		&company.Description,
		&company.Phone,
		&company.Email,
	); err != nil {
		logger.Errorf(ctx, ops, "failed to select company: %v", err)
		return nil, err
	}

	return &company, nil
}
