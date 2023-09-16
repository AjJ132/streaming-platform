//Backend API controller for user frontend

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"

	"net/http/httputil"
	"net/url"
)




func main(){
	fmt.Println("Starting Backend API Controller...")
	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
	})

	http.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) {
	})

	//Start server and host on port 8081
	log.Fatal(http.ListenAndServe("0.0.0.0:8081", nil))
	
}

//Token Signing Key for JWT
var tokenSignKey = []byte("not-very-secret-key")
	
//signup
func Signup(w http.ResponseWriter, r *http.Request) {

	//Validate JWT token
	tokenString := r.Header.Get("Authorization")
	if !ValidateToken(tokenString) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

    url, _ := url.Parse("http://user-info-write-service:8086/signup")
    proxy := httputil.NewSingleHostReverseProxy(url)
    proxy.ServeHTTP(w, r)
}

//signin
func Signin(w http.ResponseWriter, r *http.Request) {
	
	//Validate JWT token
	tokenString := r.Header.Get("Authorization")
	if !ValidateToken(tokenString) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	url, _ := url.Parse("http://user-info-read-service:8086/signin")
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ServeHTTP(w, r)
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