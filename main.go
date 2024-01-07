package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
)

// Configurable elements
// Global variables initialized with values from environment variables
var (
	upstreamProxyServerAddress     = os.Getenv("UPSTREAM_PROXY_ADDRESS")     // e.g., "upstream-proxy:12312"
	upstreamProxyServerCredentials = os.Getenv("UPSTREAM_PROXY_CREDENTIALS") // e.g., "username:password"
	upstreamProxyURL               = fmt.Sprintf("http://%s@%s", upstreamProxyServerCredentials, upstreamProxyServerAddress)
	listenPort                     = os.Getenv("PROXY_LISTEN_PORT") // e.g., ":8080"
)

// handleConnect deals with CONNECT method requests
func handleConnect(w http.ResponseWriter, r *http.Request) {
	// Dial the upstream proxy
	upstreamConn, err := net.Dial("tcp", upstreamProxyServerAddress)
	if err != nil {
		http.Error(w, "Error connecting to upstream proxy: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Encode credentials for the Proxy-Authorization header
	proxyAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(upstreamProxyServerCredentials))
	// Send CONNECT request to upstream proxy
	fmt.Fprintf(upstreamConn, "CONNECT %s HTTP/1.1\r\nHost: %s\r\nProxy-Authorization: %s\r\n\r\n", r.URL.Host, r.URL.Host, proxyAuth)

	// Wait for a response from the upstream proxy
	br := bufio.NewReader(upstreamConn)
	resp, err := http.ReadResponse(br, r)
	if err != nil {
		upstreamConn.Close()
		http.Error(w, "Error reading response from upstream proxy: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	if resp.StatusCode != http.StatusOK {
		upstreamConn.Close()
		http.Error(w, "Non-OK response from upstream proxy: "+resp.Status, resp.StatusCode)
		return
	}

	// Hijack the client connection to tunnel data
	hj, ok := w.(http.Hijacker)
	if !ok {
		upstreamConn.Close()
		http.Error(w, "Server does not support hijacking", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hj.Hijack()
	if err != nil {
		upstreamConn.Close()
		http.Error(w, "Error hijacking client connection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Inform the client that the connection has been established
	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	// Start tunneling data between client and upstream proxy
	go transfer(clientConn, upstreamConn)
	go transfer(upstreamConn, clientConn)
}

// transfer handles the bidirectional data transfer
func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

// handleRequestAndRedirect deals with non-CONNECT method requests and redirects them
func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	// Parse the upstream proxy URL
	proxyURL, err := url.Parse(upstreamProxyURL)
	if err != nil {
		http.Error(res, "Error parsing proxy URL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a transport that uses the upstream proxy
	transport := &http.Transport{Proxy: http.ProxyURL(proxyURL)}

	// Create a client with the transport
	client := &http.Client{Transport: transport}

	// Modify the request to be sent to the actual destination
	req.URL.Scheme = "http"
	req.URL.Host = req.Host
	req.RequestURI = ""

	// Forward the request to the upstream proxy
	response, err := client.Do(req)
	if err != nil {
		http.Error(res, "Error forwarding request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	// Copy headers and status code from the response to the client
	for key, values := range response.Header {
		for _, value := range values {
			res.Header().Add(key, value)
		}
	}
	res.WriteHeader(response.StatusCode)

	// Copy the response body to the client
	io.Copy(res, response.Body)
}

// main sets up the HTTP server and routes requests
func main() {
	// Validate that necessary environment variables are set
	if upstreamProxyServerAddress == "" || upstreamProxyServerCredentials == "" || listenPort == "" {
		log.Fatal("Environment variables UPSTREAM_PROXY_ADDRESS, UPSTREAM_PROXY_CREDENTIALS, and PROXY_LISTEN_PORT must be set")
	}
	// Start an HTTP server and listen on the specified port
	log.Fatal(http.ListenAndServe(":"+listenPort, http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodConnect {
			handleConnect(res, req) // Handle CONNECT method
		} else {
			handleRequestAndRedirect(res, req) // Handle all other methods
		}
	})))
}
