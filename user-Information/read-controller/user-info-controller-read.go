package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

//Write database connection
var db *sql.DB

func main(){
	fmt.Println("Starting User Information Read Controller...")

	initDB()	

	//Init handler
	handler := &LoginHandler{
		DB:     db,
		Hasher: BcryptHasher{},
	}
	

	http.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request){
		handler.Signin(w,r)
	})

	//start the server and host on port 8086
	log.Fatal(http.ListenAndServe("0.0.0.0:8085",nil))
}

func initDB(){
	//Initialize connections to write database
	fmt.Println("Attempting to connect to read database...")
	var err error

	//username TODO: Get from environment variable
	// user := os.Getenv("POSTGRES_USER")
	user := "admin"

	//password TODO: Get from environment variable
	// password := os.Getenv("POSTGRES_PASSWORD")
	password := "password"

	//Connection string TODO replace with kubernetes service name
	//connString := fmt.Sprintf("postgres://%s:%s@user-information-write-service:5432/user_information_db?sslmode=disable", user, password)
	connString := fmt.Sprintf("postgres://%s:%s@user-info-database-service:5432/user_information_db?sslmode=disable", user, password)
	db, err = sql.Open("postgres", connString)

	if(err != nil){
		fmt.Println("Error opening database connection")
		panic(err)
	}

	err = db.Ping()
	if(err != nil){
		fmt.Println("Error pinging the database connection")
		panic(err)
	}

	fmt.Println("Database connection successful!")
}

func (h *LoginHandler) Signin(w http.ResponseWriter, r *http.Request) {
	//decode request body into struct
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//query database for user
	result := h.DB.QueryRow(`SELECT user_password FROM Users WHERE user_username =$1`, creds.user_username)
	storedCreds := &Credentials{}
	err = result.Scan(&storedCreds.user_password)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("User not found")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//compare passwords using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(storedCreds.user_password), []byte(creds.user_password))
	if err != nil {
		fmt.Println("Passwords do not match")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	//generate JWT token
	token, err := generateToken(creds.user_id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//return token
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

var tokenSignKey = []byte("not-very-secret-key")

type BcryptHasher struct{}

type Credentials struct {
	user_id string
	user_password string `json:"password"`
	user_username string `json:"username"`
}

type LoginHandler struct {
	DB     *sql.DB
	Hasher Hasher
}

type UserHandler struct {
	DB *sql.DB
}

type Hasher interface {
	GenerateFromPassword(password []byte, cost int) ([]byte, error)
}

func (bh BcryptHasher) GenerateFromPassword(password []byte, cost int) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, cost)
}

func generateToken(id string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": id,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	// return token
	return token.SignedString(tokenSignKey)
}
