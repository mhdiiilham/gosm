package repository

var (
	// SQLStatementInsertUser is an SQL query for inserting a new user into the "users" table.
	// Returns the newly created user's ID.
	SQLStatementInsertUser = `
		INSERT INTO "users" (first_name, last_name, role, email, password, phone_number, country_code)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING "id";
	`

	// SQLStatementSelectUserByEmail is an SQL query to retrieve a user by email.
	// It selects the user's ID, name, role, email, and hashed password.
	// The query excludes soft-deleted users by checking if deleted_at IS NULL.
	// Only one user is returned due to the LIMIT 1 clause.
	SQLStatementSelectUserByEmail = `
		SELECT
			id,
			first_name,
			last_name,
			role,
			email,
			password,
			phone_number,
			country_code
		FROM "users"
		WHERE email = $1 AND deleted_at IS NULL
		LIMIT 1;
	`

	// SQLStatementSelectUserByID is an SQL query to retrieve a user by ID.
	// It selects the user's ID, name, role, email, and hashed password.
	// The query excludes soft-deleted users by checking if deleted_at IS NULL.
	// Only one user is returned due to the LIMIT 1 clause.
	SQLStatementSelectUserByID = `
		SELECT
			id,
			first_name,
			last_name,
			role,
			email,
			password,
			phone_number,
			country_code
		FROM "users"
		WHERE id = $1 AND deleted_at IS NULL
		LIMIT 1;
	`
)
