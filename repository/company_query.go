package repository

var (
	// SQLInsertCompany ...
	SQLInsertCompany = `
	INSERT INTO companies (name)
	VALUES ($1)
	RETURNING id;
	`

	// SQLSelectCompany ...
	SQLSelectCompany = `
	SELECT
		id, name, address, logo_url, website, description, phone, email
	FROM companies
	WHERE companies.id = $1
	LIMIT 1;
	`
)
