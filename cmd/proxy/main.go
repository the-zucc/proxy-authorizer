package main

import (
	"io"
	"log"
	"net"
	"net/http"
)

func handleConnect(res http.ResponseWriter, req *http.Request) {
	// Dial the target server
	destConn, err := net.Dial("tcp", req.URL.Host)
	if err != nil {
		http.Error(res, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Send 200 Connection established to the client
	res.WriteHeader(http.StatusOK)

	// Hijack the connection to the client
	hijacker, ok := res.(http.Hijacker)
	if !ok {
		http.Error(res, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// Transfer data between client and destination server
	go transfer(destConn, clientConn)
	go transfer(clientConn, destConn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	// Parse the destination server
	url := req.URL.Scheme + "://" + req.URL.Host + req.URL.Path
	if req.URL.RawQuery != "" {
		url += "?" + req.URL.RawQuery
	}

	// Create a new request based on the original
	proxyReq, err := http.NewRequest(req.Method, url, req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy the headers from the original request
	for header, values := range req.Header {
		for _, value := range values {
			proxyReq.Header.Add(header, value)
		}
	}

	// Send the request to the destination server
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy headers from the response
	for header, values := range resp.Header {
		for _, value := range values {
			res.Header().Add(header, value)
		}
	}

	// Set the status code from the response
	res.WriteHeader(resp.StatusCode)

	// Copy the response body to the client
	io.Copy(res, resp.Body)
}

func main() {
	log.Fatal(http.ListenAndServe(":8080", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodConnect {
			handleConnect(res, req)
		} else {
			handleRequestAndRedirect(res, req)
		}
	})))
}
