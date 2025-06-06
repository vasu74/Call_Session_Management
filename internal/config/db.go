package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() *sql.DB {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require", host, port, user, password, dbname)
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	fmt.Println("Connected to database")
	enableExtensions()
	createTables()
	return DB
}

func enableExtensions() {
	extensions := []string{
		"CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"",
		"CREATE EXTENSION IF NOT EXISTS \"pgcrypto\"",
		"CREATE EXTENSION IF NOT EXISTS \"pg_stat_statements\"",
	}

	for _, ext := range extensions {
		_, err := DB.Exec(ext)
		if err != nil {
			log.Printf("Warning: Error enabling extension %s: %v", ext, err)
		}
	}
}

func createTables() {
	// Create enum types
	createStatusEnum := `
	DO $$ BEGIN
		CREATE TYPE session_status AS ENUM ('ongoing', 'completed', 'failed');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;`

	createRoleEnum := `
	DO $$ BEGIN
		CREATE TYPE user_role AS ENUM ('user', 'admin');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;`

	// users table
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		role user_role NOT NULL DEFAULT 'user',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	// session table
	sessionTable := `
	CREATE TABLE IF NOT EXISTS sessions (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		started_at TIMESTAMP NOT NULL,
		ended_at TIMESTAMP,
		caller_id TEXT NOT NULL,
		callee_id TEXT NOT NULL,
		status session_status NOT NULL DEFAULT 'ongoing',
		initial_metadata JSONB,
		disposition TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT valid_session_times CHECK (ended_at IS NULL OR ended_at >= started_at)
	);`

	// session_events table
	sessionEventsTable := `
	CREATE TABLE IF NOT EXISTS session_events (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
		event_type TEXT NOT NULL,
		event_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		metadata JSONB,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT valid_event_time CHECK (event_time >= CURRENT_TIMESTAMP - INTERVAL '1 year')
	);`

	// Create indexes
	createIndexes := `
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
	CREATE INDEX IF NOT EXISTS idx_sessions_caller_id ON sessions(caller_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_callee_id ON sessions(callee_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
	CREATE INDEX IF NOT EXISTS idx_sessions_created_at ON sessions(created_at);
	CREATE INDEX IF NOT EXISTS idx_session_events_session_id ON session_events(session_id);
	CREATE INDEX IF NOT EXISTS idx_session_events_event_time ON session_events(event_time);
	CREATE INDEX IF NOT EXISTS idx_session_events_event_type ON session_events(event_type);
	CREATE INDEX IF NOT EXISTS idx_sessions_initial_metadata ON sessions USING GIN (initial_metadata);
	CREATE INDEX IF NOT EXISTS idx_session_events_metadata ON session_events USING GIN (metadata);`

	// Create updated_at trigger function
	createUpdatedAtTrigger := `
	CREATE OR REPLACE FUNCTION update_updated_at_column()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
	$$ language 'plpgsql';

	DROP TRIGGER IF EXISTS update_users_updated_at ON users;
	CREATE TRIGGER update_users_updated_at
		BEFORE UPDATE ON users
		FOR EACH ROW
		EXECUTE FUNCTION update_updated_at_column();

	DROP TRIGGER IF EXISTS update_sessions_updated_at ON sessions;
	CREATE TRIGGER update_sessions_updated_at
		BEFORE UPDATE ON sessions
		FOR EACH ROW
		EXECUTE FUNCTION update_updated_at_column();`

	// Execute all statements
	statements := []string{
		createStatusEnum,
		createRoleEnum,
		usersTable,
		sessionTable,
		sessionEventsTable,
		createIndexes,
		createUpdatedAtTrigger,
	}

	for _, stmt := range statements {
		_, err := DB.Exec(stmt)
		if err != nil {
			log.Printf("Error executing statement: %v\nStatement: %s", err, stmt)
		}
	}
}
