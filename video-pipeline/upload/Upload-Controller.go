package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
)

//Read Database Connection
var db *sql.DB

func main(){
	fmt.Println("Starting Upload Controller...")
	initDB()

	// //create handler for user login and signup
	// handler := &LoginHandler{
	// 	DB:     db,
	// 	Hasher: BcryptHasher{},
	// }

	// //create handler for user information
	// userHandler := &UserHandler{
	// 	DB: db,
	// }

	// http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
	// 	handler.Signup(w, r)
	// })

	// http.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) {
	// 	handler.Signin(w, r)
	// })

	// http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
	// 	userHandler.UpdateUserInfo(w, r)
	// })

	//Start server and host on port 8086
	log.Fatal(http.ListenAndServe("0.0.0.0:8086", nil))
}

//Method to receive metadata for video upload
func upload(w http.ResponseWriter, r *http.Request) {
	//Verify JWT Token

	//Generate Token for Video Upload Queue
	// token, err := generateToken("1")

	//Create location for video upload in persisten volume

	//On successful creation, add slot to video upload queue, and return confirmation token to user

}

func CreateNewVideoFolder(folderName string) {
	//create folder in persistent volume
	
}



// type LoginHandler struct {
// 	DB     *sql.DB
// 	Hasher Hasher
// }

// type UserHandler struct {
// 	DB *sql.DB
// }

	

// Generate Video Queue Token
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