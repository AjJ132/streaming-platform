package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

//Read Database Connection
var db *sql.DB

func main(){
	fmt.Println("Starting Backend API Controller...")
	initDB()

	//create handler for user login and signup
	handler := &LoginHandler{
		DB:     db,
		Hasher: BcryptHasher{},
	}

	//create handler for user information
	userHandler := &UserHandler{
		DB: db,
	}

	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		handler.Signup(w, r)
	})

	http.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) {
		handler.Signin(w, r)
	})

	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		userHandler.UpdateUserInfo(w, r)
	})

	//Start server and host on port 8086
	log.Fatal(http.ListenAndServe("0.0.0.0:8086", nil))
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
	connString := fmt.Sprintf("postgres://%s:%s@localhost:5432/user_information_db?sslmode=disable", user, password)
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

var tokenSignKey = []byte("not-very-secret-key")

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

type BcryptHasher struct{}

type Credentials struct {
	user_id string
	user_password string `json:"password"`
	user_username string `json:"username"`
}

type User_Write struct {
	email string `json:"email"`
	first_name string `json:"first_name"`
	last_name string `json:"last_name"`
	date_joined time.Time `json:"date_joined"`
	channel_id int `json:"channel_id"`
}

type CombinedCreds struct {
    Credentials Credentials `json:"credentials"`
    UserInfo    User_Write  `json:"user_info"`
}

	

func (h *LoginHandler) Signup(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Signup request received")
	//decode request body into struct
	combinedCreds := &CombinedCreds{}
	err := json.NewDecoder(r.Body).Decode(combinedCreds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}


	//hash password
	hashedPassword, err := h.Hasher.GenerateFromPassword([]byte(combinedCreds.Credentials.user_password), 10)

	//Check for errors from hashed password
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//generate UUID
	newUUID := uuid.New().String()
	combinedCreds.Credentials.user_id = newUUID
	
	//set date joined
	combinedCreds.UserInfo.date_joined = time.Now()

	//set channel id to zero
	combinedCreds.UserInfo.channel_id = 0


	//insert user into login table
	if _, err = h.DB.Exec(`INSERT INTO User_Login (user_id,user_userName, user_password) VALUES ($1, $2, $3)`, combinedCreds.Credentials.user_id, combinedCreds.Credentials.user_username, string(hashedPassword)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//insert user information into user info table
	if _, err = h.DB.Exec(`INSERT INTO User_Info (user_id, user_email, user_firstName, user_lastName, user_dateJoined, user_channelID) VALUES ($1, $2, $3, $4, $5, $6)`, combinedCreds.Credentials.user_id, combinedCreds.UserInfo.email, combinedCreds.UserInfo.first_name, combinedCreds.UserInfo.last_name, combinedCreds.UserInfo.date_joined, combinedCreds.UserInfo.channel_id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	
	//generate JWT token
	token, err := generateToken(combinedCreds.Credentials.user_id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//return token
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
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

//Update User Information
func (h *UserHandler) UpdateUserInfo(w http.ResponseWriter, r *http.Request) {
	//verify JWT token
	tokenString := r.Header.Get("Authorization")
	if !ValidateToken(tokenString) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	//decode request body into struct
	userInfo := &User_Write{}
	err := json.NewDecoder(r.Body).Decode(userInfo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//get user id from token
	token, err := jwt.Parse(tokenString, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	claims, _ := token.Claims.(jwt.MapClaims)
	user_id := claims["id"].(string)

	//TEMP print user id
	fmt.Println(user_id)

	//if user id is empty return error
	if user_id == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//update user information
	if _, err = h.DB.Exec(`UPDATE User_Info SET user_email=$1, user_firstName=$2, user_lastName=$3 WHERE user_id=$4`, userInfo.email, userInfo.first_name, userInfo.last_name, user_id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//return success
	w.WriteHeader(http.StatusOK)
}

// Generate JWT Token
func generateToken(id string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": id,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	// return token
	return token.SignedString(tokenSignKey)
}

// Validate JWT Token
func ValidateToken(tokenString string) bool {
	// Verify JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return tokenSignKey, nil
	})

	if err != nil {
		return false
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return true
	} else {
		return false
	}
}