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

type BankAccount struct {
	ID            int `json:"id"`
	UserID        string `json:"user_id"`
	AcctNo 		  int `json:"acct_no"`
	AcctName      string `json:"acct_name"`
	Balance       int `json:"balance"`
	Amount       int `json:"amount"`
	From		 int `json:"from"`
	To			 int `json:"to"`
}

type UserService interface {
	All() ([]User, error)
	Insert(user *User) error
	GetByID(id int) (*User, error)
	Update(id int,user *User) (*User, error)
	DeleteByID(id int) error
	
}

type AccountService interface {
	InsertAcct(userid int,bacct *BankAccount) error
	GetAcctByID(userid int) ([]BankAccount, error)
	DeleteAcctByID(id int) error
	UpdateBalance(id int,balance int, acct *BankAccount) (*BankAccount, error)
	GetAcctByAcctID(id int) (*BankAccount, error)
	GetAcctByAcctNo(id int) (*BankAccount, error)
}

type UserServiceImp struct {
	db *sql.DB
}

type AccountServiceImp struct {
	db *sql.DB
}

type Server struct {
	db            *sql.DB
	userService   UserService
	acctService	  AccountService			
}

func (s *Server) All(c *gin.Context) {
	users, err := s.userService.All()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"object":  "error",
			"message": fmt.Sprintf("db: query error: %s", err),
		})
		return
	}
	c.JSON(http.StatusOK, users)
}
func (s *UserServiceImp) All() ([]User, error) {
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

	if err := s.userService.Insert(&user); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (s *UserServiceImp) Insert(user *User) error {
	row := s.db.QueryRow("INSERT INTO users (first_name, last_name) values (?, ?)", user.FirstName,user.LastName)
	if err := row.Scan(&user.ID); err != nil {
		return err
	}
	return nil
}

func (s *Server) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user, err := s.userService.GetByID(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (s *UserServiceImp) GetByID(id int) (*User, error) {
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
	userupdated, err := s.userService.Update(id,&user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, userupdated)
}

func (s *UserServiceImp) Update(id int,user *User) (*User, error) {
	stmt := "UPDATE users SET first_name = ?, last_name = ? WHERE id = ?"
	_, err := s.db.Exec(stmt, user.FirstName, user.LastName, id)
	if err != nil {
		return nil, err
	}
	return s.GetByID(id)
}

func (s *Server) DeleteByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := s.userService.DeleteByID(id); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
}

func (s *UserServiceImp) DeleteByID(id int) error {
	stmt := "DELETE FROM users WHERE id = ?"
	_, err := s.db.Exec(stmt, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) InsertAcct(c *gin.Context) {
	var acct BankAccount
	err := c.ShouldBindJSON(&acct)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"object":  "error",
			"message": fmt.Sprintf("json: wrong params: %s", err),
		})
		return
	}
	userid, _ := strconv.Atoi(c.Param("id"))
	if err := s.acctService.InsertAcct(userid,&acct); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, acct)
}

func (s *AccountServiceImp) InsertAcct(userid int,acct *BankAccount) error {
	row := s.db.QueryRow("INSERT INTO accounts (user_id, acct_no, acct_name) values (?, ?, ?)", userid, acct.AcctNo,acct.AcctName)
	if err := row.Scan(&acct.ID); err != nil {
		return err
	}
	return nil
}

func (s *Server) GetAcctByID(c *gin.Context) {
	userid, _ := strconv.Atoi(c.Param("id"))
	accts, err := s.acctService.GetAcctByID(userid)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"object":  "error",
			"message": fmt.Sprintf("db: query error: %s", err),
		})
		return
	}
	c.JSON(http.StatusOK, accts)
}
func (s *AccountServiceImp) GetAcctByID(userid int) ([]BankAccount, error) {
	stmt := "SELECT acct_no,acct_name FROM accounts WHERE user_id = ? ORDER BY acct_no DESC"
	rows, err := s.db.Query(stmt,userid)
	if err != nil {
		return nil, err
	}
	bankAccts := []BankAccount{} // set empty slice without nil
	for rows.Next() {
		var acct BankAccount
		err := rows.Scan(&acct.AcctNo, &acct.AcctName)
		if err != nil {
			return nil, err
		}
		bankAccts = append(bankAccts, acct)
	}
	return bankAccts, nil
}
func (s *Server) DeleteAcctByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := s.acctService.DeleteAcctByID(id); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
}

func (s *AccountServiceImp) DeleteAcctByID(id int) error {
	stmt := "DELETE FROM accounts WHERE id = ?"
	_, err := s.db.Exec(stmt, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) WithDraw(c *gin.Context) {
	var acct BankAccount
	err := c.ShouldBindJSON(&acct)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"object":  "error",
			"message": fmt.Sprintf("json: wrong params: %s", err),
		})
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	amount := acct.Amount
	getacct, err := s.acctService.GetAcctByAcctID(id)
	balance := getacct.Balance-amount
	acctupdated, err := s.acctService.UpdateBalance(id,balance,&acct)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, acctupdated)
}

func (s *Server) Deposit(c *gin.Context) {
	var acct BankAccount
	err := c.ShouldBindJSON(&acct)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"object":  "error",
			"message": fmt.Sprintf("json: wrong params: %s", err),
		})
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	amount := acct.Amount
	getacct, err := s.acctService.GetAcctByAcctID(id)
	balance := getacct.Balance+amount
	acctupdated, err := s.acctService.UpdateBalance(id,balance,&acct)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, acctupdated)
}

func (s *Server) Transfers(c *gin.Context) {
	var acct BankAccount
	err := c.ShouldBindJSON(&acct)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"object":  "error",
			"message": fmt.Sprintf("json: wrong params: %s", err),
		})
		return
	}
	amount := acct.Amount
	getacctFrom, err := s.acctService.GetAcctByAcctNo(acct.From)
	fromBalance := getacctFrom.Balance-amount
	from, err := s.acctService.UpdateBalance(getacctFrom.ID,fromBalance,&acct)

	getacctTo, err := s.acctService.GetAcctByAcctNo(acct.To)
	toBalance := getacctTo.Balance+amount
	_, err = s.acctService.UpdateBalance(getacctTo.ID,toBalance,&acct)
	
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, from)
}

func (s *AccountServiceImp) UpdateBalance(id int,balance int,acct *BankAccount) (*BankAccount, error) {

	stmt := "UPDATE accounts SET balance = ? WHERE id = ?"
	_, err := s.db.Exec(stmt,balance,id)
	if err != nil {
		return nil, err
	}
	return s.GetAcctByAcctID(id)
}

func (s *AccountServiceImp) GetAcctByAcctID(id int) (*BankAccount, error) {
	stmt := "SELECT id, acct_no, acct_name, balance FROM accounts WHERE id = ?"
	row := s.db.QueryRow(stmt, id)
	var acct BankAccount
	err := row.Scan(&acct.ID,&acct.AcctNo, &acct.AcctName, &acct.Balance)
	if err != nil {
		return nil, err
	}
	return &acct, nil
}

func (s *AccountServiceImp) GetAcctByAcctNo(acctNo int) (*BankAccount, error) {
	stmt := "SELECT id, acct_no, acct_name, balance FROM accounts WHERE acct_no = ?"
	row := s.db.QueryRow(stmt, acctNo)
	var acct BankAccount
	err := row.Scan(&acct.ID,&acct.AcctNo, &acct.AcctName, &acct.Balance)
	if err != nil {
		return nil, err
	}
	return &acct, nil
}

func setupRoute(s *Server) *gin.Engine {
	r := gin.Default()
	bau := r.Group("/",gin.BasicAuth(gin.Accounts{
		"admin":"1234",
	}))
	bau.GET("/users", s.All)
	bau.POST("/users", s.Create)
	bau.GET("/users/:id", s.GetByID)
	bau.PUT("users/:id", s.Update)
	bau.DELETE("users/:id", s.DeleteByID)

	bau.POST("/users/:id/bankAccounts", s.InsertAcct)
	bau.GET("/users/:id/bankAccounts", s.GetAcctByID)
	bau.DELETE("/bankAccounts/:id", s.DeleteAcctByID)
	bau.PUT("/bankAccounts/:id/withdraw",s.WithDraw)
	bau.PUT("/bankAccounts/:id/deposit",s.Deposit)
	bau.POST("/transfers",s.Transfers)
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
		userService: &UserServiceImp{
			db: db,
		},
		acctService: &AccountServiceImp{
			db: db,
		},
	}

	r := setupRoute(s)

	r.Run(addr)
}