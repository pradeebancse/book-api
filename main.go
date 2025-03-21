package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Database connection
var db *sql.DB

// Book struct
type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

// User struct
type User struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// Fetch all books
func getAllBooks(c *gin.Context) {
	// Get user details from headers set by Traefik
	userEmail := c.GetHeader("X-User-Email")
	userRole := c.GetHeader("X-User-Role")

	log.Printf("Request from user: %s with role: %s", userEmail, userRole)

	rows, err := db.Query("SELECT id, title, author FROM books")
	if err != nil {
		log.Println("Failed to fetch books:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch books"})
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Author); err != nil {
			log.Println("Error scanning book:", err)
			continue
		}
		books = append(books, book)
	}

	c.JSON(http.StatusOK, books)
}

// Add a new book (Admin only)
func createBook(c *gin.Context) {
	var book Book
	if err := c.ShouldBindJSON(&book); err != nil {
		log.Println("Invalid book data:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	_, err := db.Exec("INSERT INTO books (title, author) VALUES ($1, $2)", book.Title, book.Author)
	if err != nil {
		log.Println("Failed to insert book:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add book"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Book added successfully"})
}

// Admin API to update user role
func updateUserRole(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Println("Invalid input:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	_, err := db.Exec("UPDATE users SET role = $1 WHERE email = $2", user.Role, user.Email)
	if err != nil {
		log.Println("Failed to update role:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User role updated successfully"})
}

func main() {
	var err error
	connStr := "user=postgres dbname=postgres sslmode=disable password=mysecretpassword host=db port=5432"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Database connection error:", err)
	}

	router := gin.Default()

	// Book API routes - no auth middleware needed anymore
	router.GET("/books", getAllBooks)
	router.POST("/books", createBook)

	// Admin API to update roles - no auth middleware needed anymore
	router.PUT("/users/role", updateUserRole)

	router.Run(":8080")
}
