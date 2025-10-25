package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

type Response struct {
	Status  int
	Message string
}

func main() {
	var db *sql.DB

	if os.Getenv("DB_HOST") != "" {
		db = dbConnect()
		defer db.Close()
	}

	http.HandleFunc("/backend", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[/backend] Received request - Method: %s, RemoteAddr: %s, Referer: %s", r.Method, r.RemoteAddr, r.Referer())
		username := "backend"
		body := Response{http.StatusOK, "Hello World, " + username + "!"}
		res, err := json.Marshal(body)
		if err != nil {
			log.Printf("[/backend] ERROR: Failed to marshal response: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(res)
		log.Printf("[/backend] Response sent successfully: %s", string(res))
	})

	http.HandleFunc("/notification", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[/notification] Received request - Method: %s, RemoteAddr: %s, Referer: %s", r.Method, r.RemoteAddr, r.Referer())
		id := r.FormValue("id")
		log.Printf("[/notification] Request parameter - id: %s", id)
		msg := "no message"

		if db == nil {
			log.Printf("[/notification] ERROR: Database connection not available")
			body := Response{http.StatusServiceUnavailable, "Database connection not available"}
			res, _ := json.Marshal(body)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write(res)
			return
		}

		if id != "" {
			log.Printf("[/notification] Fetching notification for id: %s", id)
			n, err := getNotification(db, id)
			if err != nil {
				log.Printf("[/notification] ERROR: Error getting notification: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			byteMsg, err := json.Marshal(n)
			if err != nil {
				log.Printf("[/notification] ERROR: Error marshaling notification: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			msg = string(byteMsg)
			log.Printf("[/notification] Successfully fetched notification: %s", msg)
		} else {
			log.Printf("[/notification] No id parameter provided, returning default message")
		}

		body := Response{http.StatusOK, msg}
		res, err := json.Marshal(body)
		if err != nil {
			log.Printf("[/notification] ERROR: Failed to marshal response: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(res)
		log.Printf("[/notification] Response sent successfully: %s", string(res))
	})

	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[/healthcheck] Received request - Method: %s, RemoteAddr: %s", r.Method, r.RemoteAddr)
		fmt.Fprintf(w, "healthcheck OK")
		log.Printf("[/healthcheck] Response sent: healthcheck OK")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func dbConnect() *sql.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=require TimeZone=Asia/Tokyo connect_timeout=10",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Printf("WARNING: failed to init database: %v", err)
		return nil
	}

	err = db.Ping()
	if err != nil {
		log.Printf("WARNING: failed to connect database: %v", err)
		return nil
	}

	log.Default().Println("success to connect db!!")

	return db
}

type Notification struct {
	ID           string
	CreatedAt    string
	UpdatedAt    string
	IsRead       bool
	IsDeleted    bool
	Verification bool
	Email        string
	Body         string
}

func getNotification(db *sql.DB, id string) (*Notification, error) {
	n := &Notification{}
	err := db.QueryRow("select * from notification where id=$1", id).Scan(&n.ID, &n.CreatedAt, &n.UpdatedAt, &n.IsRead, &n.IsDeleted, &n.Verification, &n.Email, &n.Body)
	if err != nil {
		n.ID = id
		n.Body = "No message found"
	}

	return n, nil
}
