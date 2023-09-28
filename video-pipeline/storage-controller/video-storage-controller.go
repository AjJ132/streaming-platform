package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/websocket"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var VideoQue = Queue{lock: sync.Mutex{}, modified: false}


// Create a pool of clients
var clients = make(map[string]*websocket.Conn)
var mutex = &sync.Mutex{}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

//Client PassKeys 
var clientPasskeys = make(map[string]string)
var currentVideoUploadName = ""

// Create a new bucket
var bucketName string
var location string

// Initialize minio client object.
var minioClient *minio.Client


func main() {
	bucketName = "videos-upload"
	location = "us-east-1"
	

	// Initialize MinIO client
	var err error
	minioClient, err = minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("minioadmin", "minioadmin", ""),
		Secure: false,
	})



	
	err = minioClient.MakeBucket(context.TODO(), bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		exists, err := minioClient.BucketExists(context.TODO(), bucketName)
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

	go watchQueue(&VideoQue)

	// Write video chunk data to MinIO
	http.HandleFunc("/request", func(w http.ResponseWriter, r *http.Request) {
		UploadRequest(w, r)
	})

	http.HandleFunc("/ws/connect", func(w http.ResponseWriter, r *http.Request) {
		EstablishPersistenConnection(w, r)
	})

	http.HandleFunc("/ws/disconnect", func(w http.ResponseWriter, r *http.Request) {
		
	})

	http.HandleFunc("/handle-upload", func(w http.ResponseWriter, r *http.Request) {
		HandleVideoUpload(w, r)
	})

	corsObj := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:5173"}),  // specify your allowed origins
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}), // include OPTIONS for preflight
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "Chunk-Number", "Video-Name", "Client-ID"}), // add any other headers your client might send
		
	)
	

	fmt.Println("Video Storage Controller is now running on port 8010")

	//start server
	log.Fatal(http.ListenAndServe("0.0.0.0:8010" , corsObj(http.DefaultServeMux)))
}

//Function that will add a user's video upload to a queue
func UploadRequest(w http.ResponseWriter, r *http.Request){
	fmt.Println("Received request for video upload")

	//print the response of the body
	fmt.Println("")
	fmt.Println("")
	fmt.Println(r.Body)
	fmt.Println("")
	fmt.Println("")


	//Decode body
	var requestObj VideoRequest
	err := json.NewDecoder(r.Body).Decode(&requestObj)
	fmt.Println("Decoding request body")
	if err != nil {
		fmt.Println("There was an error in decoding the request body")
		http.Error(w, "Could not decode video request body", http.StatusBadRequest)
		return
	}

	fmt.Println("Video name: " + requestObj.VideoName)
	fmt.Println("User name: " + requestObj.Name)

	//create new queue item
	var queueItem QueueItem
	queueItem.Name = requestObj.Name
	queueItem.VideoName = requestObj.VideoName

	//create new unique token for queue item
	fmt.Println("Generating queue token")
	token, err := generateToken()
	if err != nil{
		http.Error(w, "There was en error in generating the queue token", http.StatusInternalServerError)
		return
	}

	queueItem.QueueToken = token
	fmt.Println("Generated token: " + token)

	//add item to queue
	VideoQue.Enqueue(queueItem)
	fmt.Println("Added item to queue")

	//return Queue token to user
	fmt.Println("Returning queue token to user")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"queueToken": token})
}

func EstablishPersistenConnection(w http.ResponseWriter, r *http.Request){
	 fmt.Println("Received request to establish persistent connection")
	 // Check if token is passed as a query parameter
	 token := r.URL.Query().Get("queueToken")

	 if token == "" {
		 // If token is not present in query parameters, check headers
		 fmt.Println("Token not found in query parameters. Checking headers")
		 token = r.Header.Get("queueToken")

		 if token == "" {
			//return error if token is not found
			http.Error(w, "Token not found in query parameters or headers", http.StatusBadRequest)
			return
		 }
	 }

	 fmt.Println("Token found: " + token)
	//verify token is in queue
	if !SearchQueue(token){
		fmt.Println("Token not found in queue")
		http.Error(w, "Token not found in queue", http.StatusBadRequest)
		return
	}

	 // Proceed to WebSocket upgrade if token is valid
	 conn, err := upgrader.Upgrade(w, r, nil)
	 if err != nil {
		fmt.Println("Websocket Upgrade Error:", err)
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	 }

	// Assume some unique client ID is obtained from request headers or other sources
	clientID := token
	mutex.Lock()
	clients[clientID] = conn
	mutex.Unlock()

	 //
	 fmt.Println("Client connected: " + clientID)
	 fmt.Println("Sending client their queue position")

	//find clients position in the queue and send that back
	clientPosition := ReturnClientsPosition(token)

	SendClientQueuePosition(clientID, clientPosition)
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

//Notify client that its their turn in the Queue to upload
func NotifyClient(clientID string) {
	fmt.Println("Notifying client that its their turn in the queue")
    passkey, _ := generateToken()  // Reuse your existing generateToken function

	fmt.Println("Generated passkey: " + passkey)
    
    mutex.Lock()
    clientPasskeys[clientID] = passkey
    conn, ok := clients[clientID]
    mutex.Unlock()

	

    if ok {
		//Create folder with video name on the Minio storage pod
		videoName := ReturnClientVideoName(clientID)
		fmt.Println("Video name: " + videoName)
		if(videoName == ""){
			GeneralErrorHandling("There was an error in retrieving the video name")
		}

		//check if folder already exists
		fmt.Println("Checking if folder already exists on Minio storage pod")
		

		fmt.Println("Attempting to create folder on Minio storage pod")
		// Create a new folder (MinIO treats folders as zero-byte objects)
		_, err := minioClient.PutObject(context.TODO(), bucketName, videoName, bytes.NewReader([]byte("")), 0, minio.PutObjectOptions{ContentType: "application/x-directory"})
		if err != nil {
			fmt.Println("There was an error in creating the folder on the Minio storage pod")
			log.Fatalln(err)
		}
		
		combinedMessage := fmt.Sprintf("CLIENTID:%s;PASSKEY:%s", clientID, passkey)
		fmt.Println("Sending client their client ID and passkey")

		if err := conn.WriteMessage(websocket.TextMessage, []byte(combinedMessage)); err != nil {
			fmt.Println("There was an error when notifying the client of their credentials")
		}


		
    }
}

//Function that will handle and save the videos to storage
func HandleVideoUpload(w http.ResponseWriter, r *http.Request){

	fmt.Println("Received request to handle video upload")
	//Decode
	passkey := r.Header.Get("Authorization")
    clientID := r.Header.Get("Client-ID") // Assuming you send the client ID in a header
	fmt.Println("Client ID: " + clientID)
	fmt.Println("Passkey: " + passkey)
    validated := ValidatePasskey(clientID, passkey)
    if !validated {
		//print the response of the body
		fmt.Println("Invalid passkey")
        http.Error(w, "Invalid passkey", http.StatusUnauthorized)
        return
    }

    // Read video chunk
    videoData, err := ioutil.ReadAll(r.Body)
    if err != nil {
		fmt.Println("Error reading video data")
		fmt.Println("")
		fmt.Println("")
		fmt.Println(r.Body)
		fmt.Println("")
		fmt.Println("")
        http.Error(w, "Failed to read video data", http.StatusInternalServerError)
        return
    }

    // Assuming you send chunk number and video name as headers
    chunkNumber := r.Header.Get("Chunk-Number")
    videoName := r.Header.Get("Video-Name")

	if chunkNumber == "" || videoName == "" {
		fmt.Println("Missing chunk number or video name")
		fmt.Println("")
		fmt.Println("")
		fmt.Println(r.Body)
		fmt.Println("")
		fmt.Println("")
        http.Error(w, "Missing chunk number or video name", http.StatusBadRequest)
        return
    }

	//set object (file) name
	objectName := fmt.Sprintf("%s/%s.mp4", videoName, chunkNumber)
	//objectName := fmt.Sprintf("%s.mp4", chunkNumber)
	fmt.Println("Attempting to upload video chunk to Minio with object name: " + objectName)
	fmt.Println("Bucket name: " + bucketName)
	_, err = minioClient.PutObject(
		context.Background(),
		bucketName,
		objectName,
		bytes.NewReader(videoData),
		int64(len(videoData)),
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)



    if err != nil {
		fmt.Println("")
		fmt.Println("")
		fmt.Println(r.Body)
		fmt.Println("")
		fmt.Println("")
        http.Error(w, "Failed to upload video chunk", http.StatusInternalServerError)
        return
    }

    w.Write([]byte("Successfully uploaded video chunk"))
	w.WriteHeader(http.StatusOK)
	
}

//Validate Client Passkey
func ValidatePasskey(clientID, passkey string) bool {
    mutex.Lock()
    storedPasskey, ok := clientPasskeys[clientID]
    mutex.Unlock()
    return ok && storedPasskey == passkey
}

//GO routine to watch Queue for changes
func watchQueue(q *Queue) {
	fmt.Println("Watching queue for changes")
    for {
        q.lock.Lock()
        if q.modified {
            q.modified = false
            q.lock.Unlock()
            fmt.Println("Queue was modified")
            time.Sleep(10 * time.Second) // Wait for 10 seconds

            // Notify the client
            if len(q.items) > 0 {
				fmt.Println("New Item in Queue. Notifying client")
                firstInQueue := q.items[0]
				currentVideoUploadName = firstInQueue.VideoName
                NotifyClient(firstInQueue.QueueToken)
            }
        } else {
            q.lock.Unlock()
        }

		
        time.Sleep(2 * time.Second) // Check every 2 seconds
    }
}

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

func ReturnClientVideoName(token string) string{
	for i := 0; i < len(VideoQue.items); i++ {
		if VideoQue.items[i].QueueToken == token {
			return VideoQue.items[i].VideoName
		}
	}
	return "";
}

//Queue Item
// Enqueue Item
func (q *Queue) Enqueue(item QueueItem) {
	q.lock.Lock()
	q.items = append(q.items, item)
	q.modified = true  	
	q.lock.Unlock()
	fmt.Println("")
	fmt.Println("Ran Enqueue")
	fmt.Println("")
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
	q.modified = true
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

func GeneralErrorHandling(error string){
	fmt.Println("There was a general unhandled error. Printing...")
	fmt.Println(error)
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
	modified bool
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

type VideoUploadPasskey struct{
	ClientID string `json:"clientID"`
	Passkey string `json:"passkey"`
	Data string `json:"data"`
}
