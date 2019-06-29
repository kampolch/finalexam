package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Customer struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

type Msg struct {
	Message string `json:"message"`
}

func createTable() {
	url := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Fatal("faltal", err.Error())
		return
	}
	defer db.Close()

	createTb := `
	CREATE TABLE IF NOT EXISTS customers(
		id SERIAL PRIMARY KEY,
		name TEXT,
		email TEXT,
		status TEXT
	);
`

	_, err = db.Exec(createTb)
	if err != nil {
		log.Fatal("faltal", err.Error())
		return
	}

}

func postCustomersHandler(c *gin.Context) {
	fmt.Println("postCustomersHandler")
	cus := Customer{}
	if err := c.ShouldBindJSON(&cus); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	url := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()

	name := cus.Name
	email := cus.Email
	status := cus.Status
	query := `
		INSERT INTO customers (name, email, status) VALUES ($1, $2, $3) RETURNING id
	`
	var id int
	row := db.QueryRow(query, name, email, status)
	err = row.Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cus.ID = id
	c.JSON(http.StatusCreated, cus)
}

func getCustomerByIdHandler(c *gin.Context) {
	fmt.Println("getCustomerByIdHandler")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	url := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()

	stmt, err := db.Prepare("SELECT id, name, email, status FROM customers WHERE id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	row := stmt.QueryRow(id)
	cus := Customer{}

	err2 := row.Scan(&cus.ID, &cus.Name, &cus.Email, &cus.Status)
	if err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cus)
}

func getCustomersHandler(c *gin.Context) {
	fmt.Println("getCustomersHandler")
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stmt, err := db.Prepare("SELECT id, name, email, status FROM customers")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := stmt.Query()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()

	customers := []Customer{}

	for rows.Next() {
		cus := Customer{}
		err := rows.Scan(&cus.ID, &cus.Name, &cus.Email, &cus.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		customers = append(customers, cus)
	}

	c.JSON(http.StatusOK, customers)
}

func putCustomersHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cus := Customer{}
	if err := c.ShouldBindJSON(&cus); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	url := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()

	stmt, err := db.Prepare("UPDATE customers SET id=$1, name=$2, email=$3, status =$4 WHERE id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := stmt.Exec(id, cus.Name, cus.Email, cus.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cus.ID = id
	c.JSON(http.StatusOK, cus)
}

func deleteCustomersByIdHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	url := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()

	sqlStatement := "DELETE FROM customers WHERE id=$1"
	_, err = db.Exec(sqlStatement, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, Msg{"customer deleted"})
}

func authMiddleware(c *gin.Context) {
	token := c.GetHeader("Authorization")

	if token != "token2019" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
		c.Abort()
		return
	}

	c.Next()
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(authMiddleware)

	r.POST("/customers", postCustomersHandler)
	r.GET("/customers/:id", getCustomerByIdHandler)
	r.GET("/customers", getCustomersHandler)
	r.PUT("/customers/:id", putCustomersHandler)
	r.DELETE("/customers/:id", deleteCustomersByIdHandler)

	return r
}

func main() {
	createTable()
	r := setupRouter()
	r.Run(":2019")
	fmt.Println("Okey")
}
