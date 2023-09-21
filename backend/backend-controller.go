//Backend API controller for user frontend

package main

import (
	"encoding/json"
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
	// http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
	// 	Signup(w, r)
	// })

	http.HandleFunc("/api/signup", func(w http.ResponseWriter, r *http.Request) {
		Signup(w, r)
	})

	// http.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) {
	// 	Signin(w, r)
	// })

	http.HandleFunc("/api/signin", func(w http.ResponseWriter, r *http.Request) {
		Signin(w, r)
	})
	
	//Start server and host on port 8081
	log.Fatal(http.ListenAndServe("0.0.0.0:8081", nil))
	
}

//Token Signing Key for JWT
var tokenSignKey = []byte("not-very-secret-key")
	
//signup
func Signup(w http.ResponseWriter, r *http.Request) {
	
	fmt.Println("Received signup request.")

	fmt.Println("Authorized. Passing Request to User Info Write Controller...")

	resp, err := http.Post("http://user-info-write-controller-service:8085/signup", "application/json", r.Body)

	if err != nil {
		fmt.Println("Error in signup request: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Response from User Info Write Controller: ", resp)

	//return response from user info write controller
	var tokenResponse struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		fmt.Println("Non 200 response: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write JSON response to client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tokenResponse); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

//signin
func Signin(w http.ResponseWriter, r *http.Request) {
	
	fmt.Println("Received signin reques.t")

	// //Validate JWT token
	// tokenString := r.Header.Get("Authorization")
	// if !ValidateToken(tokenString) {
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	fmt.Println("Unauthorized")
	// 	return
	// }

	fmt.Println("Authorized. Passing Request to User Info Write Controller...")

	url, _ := url.Parse("http://user-info-read-controller-service:8087/signin")
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