package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// Contact represents a contact entity
type Contact struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	// MySQL connection parameters
	db, err := sql.Open("mysql", "DBUserName:DBPassword@tcp(localhost:3306)/databaseName")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize Gin-Gonic router
	router := gin.Default()

	// Endpoint to list all contacts
	router.GET("/contacts", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name, email FROM contacts")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var contacts []Contact

		for rows.Next() {
			var contact Contact
			err := rows.Scan(&contact.ID, &contact.Name, &contact.Email)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			contacts = append(contacts, contact)
		}

		c.JSON(http.StatusOK, contacts)
	})

	// Endpoint to create a new contact
	router.POST("/contacts", func(c *gin.Context) {
		var contact Contact
		if err := c.ShouldBindJSON(&contact); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := db.Exec("INSERT INTO contacts (name, email) VALUES (?, ?)", contact.Name, contact.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		contactID, _ := result.LastInsertId()
		contact.ID = int(contactID)

		c.JSON(http.StatusCreated, contact)
	})

	// Endpoint to update a contact
	router.PUT("/contacts/:id", func(c *gin.Context) {
		contactID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID"})
			return
		}

		var contact Contact
		if err := c.ShouldBindJSON(&contact); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err = db.Exec("UPDATE contacts SET name = ?, email = ? WHERE id = ?", contact.Name, contact.Email, contactID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Contact updated successfully"})
	})

	// Endpoint to delete a contact
	router.DELETE("/contacts/:id", func(c *gin.Context) {
		contactID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID"})
			return
		}

		_, err = db.Exec("DELETE FROM contacts WHERE id = ?", contactID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Contact deleted successfully"})
	})

	// Endpoint to fetch contact info based on contact ID
	router.GET("/contacts/:id", func(c *gin.Context) {
		contactID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID"})
			return
		}

		var contact Contact
		err = db.QueryRow("SELECT id, name, email FROM contacts WHERE id = ?", contactID).Scan(&contact.ID, &contact.Name, &contact.Email)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, contact)
	})

	// Run the server
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
