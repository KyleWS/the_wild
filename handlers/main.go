package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
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

		// TODO: Parse received file, save it, return status of
		// compiling it.
		var buf bytes.Buffer
		username := r.FormValue("username")
		file, header, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Printf("error: %v\n", err)
			http.Error(w, fmt.Sprintf("server had error parsing file from request: %v", err), http.StatusBadRequest)
			return
		}
		defer file.Close()
		name := strings.Split(header.Filename, ".")[0]
		io.Copy(&buf, file)
		dir := "uploads/" + username + "_" + name
		os.Mkdir(dir, 0700)
		f, err := os.OpenFile(dir+header.Filename, os.O_CREATE|os.O_WRONLY, 0700)
		defer f.Close()
		if _, err := f.Write(buf.Bytes()); err != nil {
			http.Error(w, fmt.Sprintf("error saving file to disk %v", err), http.StatusInternalServerError)
			return
		}
		f.Sync()
		statusStruct := struct {
			success  bool
			filename string
			err      string
		}{
			success:  false,
			filename: header.Filename,
			err:      "",
		}
		json.NewEncoder(w).Encode(statusStruct)
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
