package repository

var (
	// SQLStatementInsertUser Insert a new user and return the user ID
	SQLStatementInsertUser = `
		INSERT INTO users (first_name, last_name, role, email, password_hash, phone, company_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id;
	`

	// SQLStatementSelectUserByEmail Select a user by email
	SQLStatementSelectUserByEmail = `
		SELECT
			id,
			first_name,
			last_name,
			role,
			email,
			password_hash,
			phone,
			company_id
		FROM users
		WHERE email = $1
		LIMIT 1;
	`

	// SQLStatementSelectUserByID Select a user by ID
	SQLStatementSelectUserByID = `
		SELECT
			id,
			first_name,
			last_name,
			role,
			email,
			password_hash,
			phone,
			company_id
		FROM users
		WHERE id = $1
		LIMIT 1;
	`
)
