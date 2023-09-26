package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var VideoQue = Queue{}

// Create a pool of clients
var clients = make(map[string]*websocket.Conn)
var mutex = &sync.Mutex{}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	// Initialize MinIO client
	minioClient, err := minio.New("CHANGE-ME:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("user", "password", ""),
		Secure: false,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create a new bucket
	bucketName := "videos-upload"
	location := "us-east-1"
	err = minioClient.MakeBucket(bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		exists, err := minioClient.BucketExists(bucketName)
		if err == nil && exists {
			log.Printf("Already own %s\n", bucketName)
		} else {
			log.Fatal(err)
		}
	}

	// Create a new folder (MinIO treats folders as zero-byte objects)
	// folderName := "newfolder/"
	// _, err = minioClient.PutObject(bucketName, folderName, bytes.NewReader([]byte("")), 0, minio.PutObjectOptions{ContentType: "application/x-directory"})
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// Write video chunk data to MinIO
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		videoChunk := buf.Bytes()

		objectName := folderName + "video-chunk-1" // Use unique names for chunks
		_, err := minioClient.PutObject(bucketName, objectName, bytes.NewReader(videoChunk), int64(len(videoChunk)), minio.PutObjectOptions{ContentType: "video/mp4"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte("Successfully uploaded video chunk"))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

//Function that will add a user's video upload to a queue
func UploadRequest(w http.ResponseWriter, r *http.Request){
	//Decode token
	var requestObj VideoRequest
	err := json.NewDecoder(r.Body).Decode(&requestObj)
	if err != nil {
		http.Error(w, "Could not decode video request body", http.StatusBadRequest)
		return
	}

	//create new queue item
	var queueItem QueueItem
	queueItem.Name = requestObj.Name
	queueItem.VideoName = requestObj.VideoName

	//create new unique token for queue item
	token, err := generateToken()
	if err != nil{
		http.Error(w, "There was en error in generating the queue token", http.StatusInternalServerError)
		return
	}

	queueItem.QueueToken = token

	//add item to queue
	VideoQue.items = append(VideoQue.items, queueItem)

	//return Queue token to user
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"queueToken": token})
}

func EstablishPersistenConnection(w http.ResponseWriter, r *http.Request){
	//decode token
	var tokenObj TokenBody
	err := json.NewDecoder(r.Body).Decode(&tokenObj)
	if err != nil {
		http.Error(w, "Could not decode persistent connection request body", http.StatusBadRequest)
		return
	}
	token := tokenObj.Token

	//verify token is in queue
	if !SearchQueue(token){
		http.Error(w, "Token not found in queue", http.StatusBadRequest)
		return
	}

	//if token is in queue, establish websocket connection with user
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}

	// Assume some unique client ID is obtained from request headers or other sources
	clientID := token
	mutex.Lock()
	clients[clientID] = conn
	mutex.Unlock()

	//find clients position in the queue and send that back
}

func SendClientQueuePosition(clientID string, number int){
	mutex.Lock()
	conn, ok := clients[clientID]
	mutex.Unlock()

	if ok {
		numberStr := strconv.Itoa(number) // Convert integer to string
		if err := conn.WriteMessage(websocket.TextMessage, []byte(numberStr)); err != nil {
			// Handle error: maybe remove the client from the map or log the error
			fmt.Println("There as an error in sharing the clients queue position")
		}
	}
}

func NotifyClient(clientID string){

}

//Function that will take a video upload from the user and upload it to MinIO

//Search queue for token
func SearchQueue(token string) bool{
	
	for i := 0; i < len(VideoQue.items); i++ {
		if VideoQue.items[i].QueueToken == token {
			return true
		}
	}
	return false
}

//search queue via token and get clients position in queue
func ReturnClientsPosition(token string) int{
	for i := 0; i < len(VideoQue.items); i++ {
		if VideoQue.items[i].QueueToken == token {
			return VideoQue.items[i].Index
		}
	}
	return -1
}

//Queue Item
func (q *Queue) Enqueue(item QueueItem) {
	q.lock.Lock()
	q.items = append(q.items, item)
	q.lock.Unlock()
}

//Dequeue Item
func (q *Queue) Dequeue() *QueueItem {
	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	item := q.items[0]
	q.items = q.items[1:]
	return &item
}

func generateToken() (string, error) {
	byteSlice := make([]byte, 16) // 16 bytes will result in 32 character string
	_, err := rand.Read(byteSlice)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(byteSlice), nil
}

var tokenSignKey = "not-very-secret-key"

//Queue Item
type QueueItem struct {
	Index      int
	Name       string
	VideoName  string
	QueueToken string
}

//Queue
type Queue struct {
	items []QueueItem
	lock  sync.Mutex
}

// Define client type
type Client struct {
	ID   string
	Conn *websocket.Conn
}

type TokenBody struct {
	Token string `json:"token"`
}

type VideoRequest struct {
	Name string `json:"name"`
	VideoName string `json:"videoName"`
}
