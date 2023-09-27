package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
)


func main(){
	fmt.Println("Starting Upload Controller...")

	//Handle Upload Request
	http.HandleFunc("/request", func(w http.ResponseWriter, r *http.Request) {
		RequestUpload(w, r)
	})

	//Start server and host on port 8086
	log.Fatal(http.ListenAndServe("0.0.0.0:8086", nil))
}

//Method to receive metadata for video upload
func RequestUpload(w http.ResponseWriter, r *http.Request) {
	//Verify JWT Token
	token := r.Header.Get("Authorization")

	//validate token
	if !ValidateToken(token) {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	//if token is valid, send the request to the upload service
	client := &http.Client{}
	
	// Read the body of the incoming request
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	// Create a new request to send to the next service
	newReq, err := http.NewRequest("POST", "http://video-storage-controller-service/request:8010", bytes.NewBuffer(bodyBytes))
	if err != nil {
		http.Error(w, "Error creating new request", http.StatusInternalServerError)
		return
	}

	// Set headers for the new request
	newReq.Header.Set("Content-Type", "application/json")

	// Send the request to the next service
	resp, err := client.Do(newReq)
	if err != nil {
		http.Error(w, "Error sending request to upload service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Forward response back to the original client
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response body", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
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

// Not so Secret key used to sign JWT tokens
var tokenSignKey = []byte("not-very-secret-key")
