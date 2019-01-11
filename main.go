package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type (
	// Options contains all of the application
	// configuration passed by the user in of cli
	Options struct {
		Method  string
		URL     string
		Body    *bytes.Buffer
		Headers Headers
	}
	// Headers is a slice of string containing
	// formated header pairs. Example:
	// 		- "Content-Type=application/json"
	//		- User-Agent=Mozilla/5.0
	Headers []string
)

// NewHeaders returns a headers array
func NewHeaders(input string) Headers {
	return Headers(strings.Split(input, ","))
}

// Parse header list into a http.Header map
// that can be used in the request
func (h Headers) Parse() http.Header {
	parsed := http.Header(make(map[string][]string))
	for _, line := range h {
		line = strings.Replace(line, "\"", "", -1)
		parts := strings.Split(line, "=")
		if len(parts) >= 2 {
			parsed.Set(parts[0], parts[1])
		}
	}
	return parsed
}

// ToRequest takes the request configuration and
// build a http request from the input
func (o Options) ToRequest() *http.Request {
	request, err := http.NewRequest(o.Method, o.URL, o.Body)
	check(err, "Could not create request from options")
	request.Header = o.Headers.Parse()
	return request
}

func check(err error, message string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %v", message, err))
	}
}

func main() {
	// Declare and parse flags
	method := flag.String("method", http.MethodGet, "Sets the http method to use")
	headers := flag.String("headers", "\"Content-Type=application/json\"", "Sets a list of headers to use. Every pair must be quoted and comma separated")
	body := flag.String("body", "", "Data to include in request body")
	stdin := flag.Bool("stdin", false, "Set to true if body should be read from stdin instead of body flag")
	flag.Parse()

	// Make sure that the user passed a url
	if len(flag.Args()) == 0 {
		fmt.Println("You must pass a url to request.")
		fmt.Printf("\t - Example: gcurl [OPTIONS] URL")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Build options object
	options := new(Options)
	options.Method = *method
	options.Headers = NewHeaders(*headers)
	options.URL = flag.Args()[0]
	if *stdin {
		buffer := new(bytes.Buffer)
		_, err := io.Copy(buffer, os.Stdin)
		check(err, "Error reading body from stdin")
		options.Body = buffer
	} else {
		buffer := new(bytes.Buffer)
		_, err := buffer.WriteString(*body)
		check(err, "Error writing to body buffer")
		options.Body = buffer
	}

	// Make request and print response
	request := options.ToRequest()
	response, err := http.DefaultClient.Do(request)
	check(err, "Error in response")
	defer response.Body.Close()
	_, err = io.Copy(os.Stdout, response.Body)
	check(err, "Could copy response body to stdout")
}
