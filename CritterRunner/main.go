package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const POST_FILE_ERROR = "error reading file from post request. Please try again or contact admin"

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

		//https://astaxie.gitbooks.io/build-web-application-with-golang/en/04.5.html
	}
}

func RunCompile(filepath string) error {
	if err := exec.Command("javac", filepath).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error compiling java program: %v\n", err)
		return err
	}
	return nil
}

func RunTest(filepath string, parameters []string) (string, error) {
	if err := RunCompile(filepath); err != nil {
		return "", err
	}
	indexOfSlash := strings.LastIndex(filepath, "/") + 1
	directoryPath := filepath[:indexOfSlash]
	programName := filepath[indexOfSlash : len(filepath)-5]
	classpathArgs := []string{"-classpath", directoryPath, programName}
	args := append(classpathArgs, parameters...)
	cmd := exec.Command("java", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("error getting stdd pipe when running java: %v", err)
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error running java program: %v\n", err)
		return "", err
	}
	// Get the output
	buf := new(bytes.Buffer)
	go buf.ReadFrom(stdout)
	// Make channel to handle timing our function;
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case <-time.After(3 * time.Second):
		if err := cmd.Process.Kill(); err != nil {
			return "", fmt.Errorf("error killing process \"%s %s\": %v\n", filepath, args, err)
		} else {
			return "", fmt.Errorf("error running \"%s %s\" took too long. killed process\n", filepath, args)
		}
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("process finished with error: %v\n", err)
		}
	}
	// This may be rerunning the command :(
	return buf.String(), nil
}

func main() {
	//addr := os.Getenv("RUNNER_ADDR")

	//mux := http.NewServeMux()
	//mux.HandleFunc("/test", RunnerHandler)
	result, err := RunTest("./CritterTest.java", []string{"W"})
	if err != nil {
		fmt.Printf("error running test %v\n", err)
	} else {
		fmt.Printf("Result of Critter Test [W] %v\n", result)
	}

	_, err = RunTest("./CritterTest.java", []string{"T"})
	if err != nil {
		fmt.Printf("error running test %v\n", err)
	}

}
