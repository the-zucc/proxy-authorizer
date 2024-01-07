package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func main() {
	// Define the URL of the resource you want to fetch
	resourceUrl := "https://www.google.com"

	// Define the proxy URL
	proxyUrl, err := url.Parse("http://localhost:8080")
	if err != nil {
		fmt.Println("Error parsing proxy URL:", err)
		return
	}

	// Create an HTTP client with the proxy
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
	}

	// Create a new request
	request, err := http.NewRequest("GET", resourceUrl, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Send the request via the proxy
	response, err := httpClient.Do(request)
	if err != nil {
		fmt.Println("Error sending request through proxy:", err)
		return
	}
	defer response.Body.Close()

	// Read and print the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	fmt.Println("Response received:")
	fmt.Println(string(body))
}
