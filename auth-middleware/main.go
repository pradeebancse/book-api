// auth-middleware/main.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	// Connect to PostgreSQL
	connStr := "user=postgres dbname=postgres sslmode=disable password=mysecretpassword host=db port=5432"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Database connection error:", err)
	}

	// Create users table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'user'
		)
	`)
	if err != nil {
		log.Fatal("Failed to create users table:", err)
	}

	// Insert admin user for testing
	_, err = db.Exec(`
		INSERT INTO users (email, role) VALUES ('admin@example.com', 'admin')
		ON CONFLICT (email) DO UPDATE SET role = 'admin'
	`)
	if err != nil {
		log.Fatal("Failed to create admin user:", err)
	}

	// Set up the HTTP server
	http.HandleFunc("/auth", handleAuth)
	log.Println("Starting auth middleware service on port 4000...")
	if err := http.ListenAndServe(":4000", nil); err != nil {
		log.Fatal(err)
	}
}

func handleAuth(w http.ResponseWriter, r *http.Request) {
	// Get user email from the Google OAuth middleware
	email := r.Header.Get("X-Forwarded-Email")
	if email == "" {
		http.Error(w, "Unauthorized - No email provided", http.StatusUnauthorized)
		return
	}

	// Get required role from headers
	requiredRole := r.Header.Get("X-Required-Role")

	// If no specific role is required, allow access
	if requiredRole == "" || requiredRole == "any" {
		// Look up user in database anyway to get their role
		setUserHeaders(w, email)
		return
	}

	// Look up the user's role
	var role string
	err := db.QueryRow("SELECT role FROM users WHERE email = $1", email).Scan(&role)
	if err == sql.ErrNoRows {
		// User not found, create with default role
		_, err = db.Exec("INSERT INTO users (email, role) VALUES ($1, 'user')", email)
		if err != nil {
			log.Printf("Error creating user: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		role = "user"
	} else if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set role in response headers
	w.Header().Set("X-User-Email", email)
	w.Header().Set("X-User-Role", role)

	// Check if user has required role
	if requiredRole != "any" && role != requiredRole {
		http.Error(w, fmt.Sprintf("Forbidden - Required role: %s", requiredRole), http.StatusForbidden)
		return
	}

	// All checks passed, allow the request
	w.WriteHeader(http.StatusOK)
}

func setUserHeaders(w http.ResponseWriter, email string) {
	// Look up user's role and set headers
	var role string
	err := db.QueryRow("SELECT role FROM users WHERE email = $1", email).Scan(&role)
	if err == sql.ErrNoRows {
		// Create new user with default role
		_, err = db.Exec("INSERT INTO users (email, role) VALUES ($1, 'user')", email)
		if err != nil {
			log.Printf("Error creating user: %v", err)
		}
		role = "user"
	} else if err != nil {
		log.Printf("Database error: %v", err)
		role = "user" // Default if error
	}

	// Set headers for downstream services
	w.Header().Set("X-User-Email", email)
	w.Header().Set("X-User-Role", role)
	w.WriteHeader(http.StatusOK)
}
