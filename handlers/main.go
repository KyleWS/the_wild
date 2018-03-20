package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const POST_FILE_ERROR = "error reading file from post request. Please try again or contact admin"

// CORS struct contains handler that will attach proper Access-Control headers
type CORS struct {
	Handler http.Handler
}

// NewCORS returns cors object with given handler assigned
func NewCORS(handler http.Handler) *CORS {
	return &CORS{
		Handler: handler,
	}
}

func (c *CORS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "GET, PUT, POST, PATCH, DELETE")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Authorization, RequestID")
	// w.Header().Add("Access-Control-Expose-Headers", "Authorization, RequestID")
	w.Header().Add("Access-Control-Max-Age", "600")
	if r.Method != "OPTIONS" {
		c.Handler.ServeHTTP(w, r)
	}
}

func PostFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		_, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("%s : %v", POST_FILE_ERROR, err), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// TODO: Make it so we hash critter files
		// and only override an existing file if there was a change.
		json.NewEncoder(w).Encode("hello")
		//https://astaxie.gitbooks.io/build-web-application-with-golang/en/04.5.html
	}
}

func main() {
	addr := os.Getenv("RUNNER_ADDR")
	if len(addr) == 0 {
		log.Printf("defaulting addr to localhost:4555")
		addr = "localhost:4555"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world\n"))
	})
	mux.HandleFunc("/upload", PostFileHandler)

	corsMux := NewCORS(mux)

	fmt.Printf("listening on %s...\n", addr)
	log.Fatal(http.ListenAndServe(addr, corsMux))
}
