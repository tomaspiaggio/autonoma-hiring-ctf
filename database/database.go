package database

import (
	"context"
	"database/sql"
	"log"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
)

// DB represents the database connection pool.
type DB struct {
	pool *sql.DB
	ctx  context.Context
}

// New connects to the database using the provided DSN and returns a DB instance.
func New(dsn string) (*DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// Ping the database to verify the connection.
	if err := db.Ping(); err != nil {
		db.Close() // Close the connection if ping fails
		return nil, err
	}

	log.Println("Successfully connected to the database")
	result := DB{pool: db, ctx: context.Background()}

	err = result.InitSchema()
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// Close closes the database connection pool.
func (db *DB) Close() error {
	return db.pool.Close()
}

// InitSchema creates the necessary tables if they do not exist.
func (db *DB) InitSchema() error {
	// Create users table
	usersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);`

	_, err := db.pool.ExecContext(db.ctx, usersTableSQL)
	if err != nil {
		log.Printf("Error creating users table: %v\n", err)
		return err
	}
	log.Println("Users table checked/created successfully.")

	// Create attempts table
	attemptsTableSQL := `
	CREATE TABLE IF NOT EXISTS attempts (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		failed BOOLEAN DEFAULT TRUE,
		details JSONB,
		submitted_at TIMESTAMPTZ DEFAULT NOW()
	);`

	_, err = db.pool.ExecContext(db.ctx, attemptsTableSQL)
	if err != nil {
		log.Printf("Error creating attempts table: %v\n", err)
		return err
	}
	log.Println("Attempts table checked/created successfully.")

	return nil
}

func (db *DB) DoesUserHaveFailedAttemptsToday(email string) (bool, error) {
	var exists int // Dummy variable to scan into
	// Select 1 to just check for existence, also ensure we check for failed attempts
	query := `
		SELECT 1 
		FROM attempts a 
		JOIN users u ON a.user_id = u.id 
		WHERE u.email = $1 
		  AND a.failed = TRUE 
		  AND a.submitted_at > NOW() - INTERVAL '24 hours' 
		LIMIT 1`
	lowerEmail := strings.ToLower(email)

	// QueryRowContext returns sql.ErrNoRows if no row is found
	err := db.pool.QueryRowContext(db.ctx, query, lowerEmail).Scan(&exists)

	if err != nil {
		if err == sql.ErrNoRows {
			// No failed attempts found in the last 24 hours
			return false, nil
		}
		// Some other database error occurred
		log.Printf("Error checking for failed attempts for %s: %v\n", email, err)
		return false, err
	}

	// A row was found, meaning a failed attempt exists
	return true, nil
}

func (db *DB) HasUserWon(email string) (bool, error) {
	var exists int // Use dummy variable
	// Adjust query to select 1 and join with users table
	query := `
		SELECT 1 
		FROM attempts a
		JOIN users u ON a.user_id = u.id
		WHERE u.email = $1 
		  AND a.failed = FALSE 
		LIMIT 1`
	lowerEmail := strings.ToLower(email)
	err := db.pool.QueryRowContext(db.ctx, query, lowerEmail).Scan(&exists)

	if err != nil {
		if err == sql.ErrNoRows {
			// No winning attempt found
			return false, nil
		}
		// Other database error
		log.Printf("Error checking if user %s has won: %v\n", email, err)
		return false, err
	}

	// A winning attempt was found
	return true, nil
}

func (db *DB) CreateUser(email string) (int, error) {
	var id int
	query := "INSERT INTO users (email) VALUES ($1) RETURNING id"
	err := db.pool.QueryRowContext(db.ctx, query, strings.ToLower(email)).Scan(&id)
	return id, err
}

func (db *DB) getUserIdByEmail(email string) (int, error) {
	var id int
	query := "SELECT id FROM users WHERE email = $1"
	err := db.pool.QueryRowContext(db.ctx, query, strings.ToLower(email)).Scan(&id)
	return id, err
}

func (db *DB) CreateAttempt(email string, failed bool, details map[string]interface{}) (int, error) {
	userId, err := db.getUserIdByEmail(email)
	if err != nil {
		return -1, err
	}
	
	var id int
	query := "INSERT INTO attempts (user_id, failed, details) VALUES ($1, $2, $3) RETURNING id"
	err = db.pool.QueryRowContext(db.ctx, query, userId, failed, details).Scan(&id)

	return id, err
}
