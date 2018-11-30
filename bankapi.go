package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"flag"
	"github.com/gin-gonic/gin"
	_"github.com/go-sql-driver/mysql"
	"strconv"
)

type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type BankService interface {
	All() ([]User, error)
	Insert(user *User) error
	GetByID(id int) (*User, error)
	Update(id int,user *User) (*User, error)
	DeleteByID(id int) error
	
}

type BankServiceImp struct {
	db *sql.DB
}


type Server struct {
	db            *sql.DB
	bankService   BankService
}

func (s *Server) All(c *gin.Context) {
	users, err := s.bankService.All()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"object":  "error",
			"message": fmt.Sprintf("db: query error: %s", err),
		})
		return
	}
	c.JSON(http.StatusOK, users)
}
func (s *BankServiceImp) All() ([]User, error) {
	stmt := "SELECT id, first_name, last_name FROM users ORDER BY id DESC"
	rows, err := s.db.Query(stmt)
	if err != nil {
		return nil, err
	}
	users := []User{} // set empty slice without nil
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.FirstName, &user.LastName)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (s *Server) Create(c *gin.Context) {
	var user User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"object":  "error",
			"message": fmt.Sprintf("json: wrong params: %s", err),
		})
		return
	}

	if err := s.bankService.Insert(&user); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (s *BankServiceImp) Insert(user *User) error {
	row := s.db.QueryRow("INSERT INTO users (first_name, last_name) values (?, ?)", user.FirstName,user.LastName)
	if err := row.Scan(&user.ID); err != nil {
		return err
	}
	return nil
}

func (s *Server) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user, err := s.bankService.GetByID(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (s *BankServiceImp) GetByID(id int) (*User, error) {
	stmt := "SELECT id,first_name, last_name  FROM users WHERE id = ?"
	row := s.db.QueryRow(stmt, id)
	var user User
	err := row.Scan(&user.ID,&user.FirstName, &user.LastName)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Server) Update(c *gin.Context) {
	var user User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"object":  "error",
			"message": fmt.Sprintf("json: wrong params: %s", err),
		})
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	userupdated, err := s.bankService.Update(id,&user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, userupdated)
}

func (s *BankServiceImp) Update(id int,user *User) (*User, error) {
	stmt := "UPDATE users SET first_name = ?, last_name = ? WHERE id = ?"
	_, err := s.db.Exec(stmt, user.FirstName, user.LastName, id)
	if err != nil {
		return nil, err
	}
	return s.GetByID(id)
}

func (s *Server) DeleteByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := s.bankService.DeleteByID(id); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
}

func (s *BankServiceImp) DeleteByID(id int) error {
	stmt := "DELETE FROM users WHERE id = ?"
	_, err := s.db.Exec(stmt, id)
	if err != nil {
		return err
	}
	return nil
}

func setupRoute(s *Server) *gin.Engine {
	r := gin.Default()
	users := r.Group("/users")

	users.Use(gin.BasicAuth(gin.Accounts{
		"admin": "1234",
	}))

	users.GET("/", s.All)
	users.POST("/", s.Create)
	users.GET("/:id", s.GetByID)
	users.PUT("/:id", s.Update)
	users.DELETE("/:id", s.DeleteByID)
	return r
}
func main() {

	host := flag.String("host", "localhost", "Host")
	port := flag.String("port", "8000", "Port")
	dbURL := flag.String("dburl", "root:@tcp(127.0.0.1:3306)/bank", "DB Connection")
	flag.Parse()
	addr := fmt.Sprintf("%s:%s", *host, *port)
	db, err := sql.Open("mysql", *dbURL)
	if err != nil {
		log.Fatal(err)
	}
	
	s := &Server{
		db: db,
		bankService: &BankServiceImp{
			db: db,
		},
	}

	r := setupRoute(s)

	r.Run(addr)
}