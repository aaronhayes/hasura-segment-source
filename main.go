package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"io"
	"log"
	"os"
	"github.com/gorilla/mux"
	"gopkg.in/segmentio/analytics-go.v3"
)

type HasuraEvent struct {
	ID	string `json:"id"`
	Event `json:"event"`
	Table `json:"table"`
	Trigger `json:"trigger"`
}

type Event struct {
	Op string `json:"op"`
	Data `json:"data"`
}

type Data struct {
	Old map[string]interface{} `json:"old"`
	New map[string]interface{} `json:"new"`
}

type Table struct {
	Name string `json:"name"`
	Schema string `json:"schema"`
}

type Trigger struct {
	ID string `json:"id"`
	Name string `json:"name"`
}

type TriggerResponse struct {
	Message string `json:"message"`
	OldData map[string]interface{} `json:"oldData"`
	NewData map[string]interface{} `json:"newData"`
}

func FetchServerHealth(w http.ResponseWriter, r *http.Request) {
    // A very simple health check.
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)

    // In the future we could report back on the status of our DB, or our cache
    // (e.g. Redis) by performing a simple PING, and include them in the response.
    io.WriteString(w, `{"alive": true}`)
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	segmentWriteKey, segmentWriteKeyExists := os.LookupEnv("SEGMENT_WRITE_API_KEY")
	userIdKey, userIdKeyExists := os.LookupEnv("USER_ID_FIELD")

	if !segmentWriteKeyExists {
		log.Println("Poorly Configured Server - Missing Segment Write Key")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if !userIdKeyExists {
		userIdKey = "user_id"
	}

	decoder := json.NewDecoder(r.Body)

	var event HasuraEvent
	err := decoder.Decode(&event)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	response := TriggerResponse{
		Message: fmt.Sprintf(
			"got '%s' for '%s' operation on '%s' table in '%s' schema from '%s' trigger",
			event.ID,
			event.Event.Op,
			event.Table.Name,
			event.Table.Schema,
			event.Trigger.Name,
		),
		OldData: event.Data.Old,
		NewData: event.Data.New,
	}

	client, _ := analytics.NewWithConfig(segmentWriteKey, analytics.Config{
        Verbose: true,
        Logger:  analytics.StdLogger(log.New(os.Stderr, "segment ", log.LstdFlags)),
    })

	user_id, ok := event.Data.New[userIdKey]
	if !ok || user_id == nil {
		user_id = "anonymous"
	}

	event_name := fmt.Sprintf("%s_%s", event.Event.Op, event.Table.Name)
	event_properties := analytics.NewProperties().
		Set("event_id", event.ID)

	for key, value := range event.Data.New {
		event_properties.Set(key, value)
	} 

	client.Enqueue(analytics.Track{
		Event: event_name,
		UserId: user_id.(string),
		Properties: event_properties,
	})

	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Println(r.RequestURI)
        next.ServeHTTP(w, r)
    })
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/webhook", WebhookHandler).Methods("POST")
	r.HandleFunc("/health", FetchServerHealth).Methods("GET")
	r.Use(loggingMiddleware)

	log.Println("Starting Hasura Segment Source Server. Listening on port 4004 ...")
	err := http.ListenAndServe(":4004", r)

	if err != nil {
		log.Println(err)
	}
}