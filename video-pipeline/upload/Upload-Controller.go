package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	_ "github.com/lib/pq"
)


func main(){
	fmt.Println("Starting Upload Controller...")

	//Handle Upload Request
	http.HandleFunc("/request", func(w http.ResponseWriter, r *http.Request) {
		RequestUpload(w, r)
	})

	// Define CORS settings
	corsObj := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}), // replace "*" with specific origin when deploying to production
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}),
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
	)

	//Start server and host on port 8086
	log.Fatal(http.ListenAndServe("0.0.0.0:8086", corsObj(http.DefaultServeMux)))
}

//Method to receive metadata for video upload
func RequestUpload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request for video upload")
	//print body
	fmt.Println("Body: ", r.Body)

	// Uncomment this block for production
	// token := r.Header.Get("Authorization")
	// if !ValidateToken(token) {
	// 	http.Error(w, "Invalid token", http.StatusBadRequest)
	// 	return
	// }     

	client := &http.Client{}
	

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		fmt.Println("Error reading request body:", err)
		return
	}

	newReq, err := http.NewRequest("POST", "http://localhost:8010/request", bytes.NewBuffer(bodyBytes))
	if err != nil {
		http.Error(w, "Error creating new request", http.StatusInternalServerError)
		fmt.Println("Error creating new request:", err)
		return
	}

	newReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(newReq)
	if err != nil {
		http.Error(w, "Error sending request to upload service", http.StatusInternalServerError)
		fmt.Println("Error sending request to upload service:", err)
		return
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response body", http.StatusInternalServerError)
		fmt.Println("Error reading response body:", err)
		return
	}

	fmt.Println("Message sent and received from upload service")
	fmt.Println(string(respBody))
	fmt.Println("Status code:", resp.StatusCode)

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
