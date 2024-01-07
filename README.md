# Proxy Authorizer

## Overview
The HTTPS Proxy Authorizer is a tool designed to provide an easy way to use a proxy that requires authorization, especially beneficial for HTTP clients that do not support using an authorized proxy. This application handles both HTTP and HTTPS traffic, forwarding requests to an upstream proxy server with the necessary credentials, and supports proxied HTTPS/TLS traffic.

## Installation
### Prerequisites
- Go programming language installed on your system.
- `GOPATH` environment variable should be set and added to your system's `PATH` environment variable.

### Installing the Application
The application can be installed using the Go package manager. Run the following command:

```sh
go install github.com/the-zucc/proxy-authorizer
```

This command downloads and installs the `proxy-authorizer` binary in your `GOPATH/bin` directory.

### Running the Proxy Server
After installation, you can run the application using the `proxy-authorizer` command:

```sh
proxy-authorizer
```

Ensure that `GOPATH/bin` is in your system's `PATH` so that the command can be found and executed.

## Configurable Elements
Configure the application using the following environment variables:

- `UPSTREAM_PROXY_ADDRESS`: Address of the upstream proxy server (e.g., "upstream-proxy:12312").
- `UPSTREAM_PROXY_CREDENTIALS`: Credentials for the upstream proxy, formatted as "username:password".
- `PROXY_LISTEN_PORT`: Port on which the proxy server listens for requests (e.g., ":8080").

### Setting Environment Variables
Set these variables before running the application:

For Unix-like systems:

```sh
export UPSTREAM_PROXY_ADDRESS="upstream-proxy:12312"
export UPSTREAM_PROXY_CREDENTIALS="username:password"
export PROXY_LISTEN_PORT=":8080"
```

Alternatively, set them directly in the command line:

```sh
UPSTREAM_PROXY_ADDRESS="upstream-proxy:12312" UPSTREAM_PROXY_CREDENTIALS="username:password" PROXY_LISTEN_PORT=":8080" https-proxy-authorizer
```

## Features
- **HTTP and HTTPS/TLS Support**: Handles both HTTP and proxied HTTPS/TLS requests.
- **Proxy Authorization**: Adds authorization headers automatically when forwarding requests to an upstream proxy that requires authentication.
- **Transparent Handling**: Forwards requests and responses transparently, ensuring data integrity.
- **Easy Configuration**: Configurable via environment variables for flexibility.
