# GOSM

<p align="center">
  <a href="https://goreportcard.com/report/github.com/mhdiiilham/gosm">
    <img src="https://goreportcard.com/badge/github.com/mhdiiilham/gosm" alt="Go Report Card">
  </a>
  <br>
  <a href="https://github.com/mhdiiilham/gosm/actions/workflows/deplyoment.yml">
    <img src="https://github.com/mhdiiilham/gosm/actions/workflows/deplyoment.yml/badge.svg" alt="Deployment">
  </a>
  <br>
  <a href="https://github.com/mhdiiilham/gosm/actions/workflows/release.yml">
    <img src="https://github.com/mhdiiilham/gosm/actions/workflows/release.yml/badge.svg" alt="Semantic Release">
  </a>
</p>

## Database Migration

### How to Create a New Migration

To create a new database migration file, follow these steps:

1. Ensure `golang-migrate` is installed.
2. Run the following Makefile command:
   ```sh
   make migrate-create
   ```
3. This will generate a new migration file in the `database/migrations/` directory.
4. Edit the generated migration file to define the required schema changes.

After creating the migration, apply it by running:

### How to Migrate the Database Schema

Follow these steps to apply database migrations:

1. Install `golang-migrate` by following the instructions at [golang-migrate/migrate](https://github.com/golang-migrate/migrate).
2. Set the environment variable `DATABASE_URL` with your database connection string.
3. Run the following Makefile command to apply the migrations:
   ```sh
   make migrate-up
   ```
