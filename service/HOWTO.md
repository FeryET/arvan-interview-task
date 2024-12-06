# Service Setup and Testing

## Setup the Environment

To properly build and test the service, you need to ensure your environment is correctly set up with Go and Docker.

### Development Environment

#### Install Go

1. **Prerequisites**: Make sure you have Go installed. This project requires Go version specified in the `go.mod` file.
   - You can verify the Go installation using:

     ```sh
     go version
     # go version is 1.21.5 for this project
     ```

2. **Build the Go Application**:
   - Navigate to the Go source directory in the `go` folder:

     ```sh
     cd ./go
     ```

   - Build the application using:

     ```sh
     go build -o /tmp/server
     ```

   - This will generate an executable binary for the application in `/tmp/server`. Now you can run it via: `$ /tmp/server/`

### Local Test Environment

#### Install Docker

Please setup docker at first, and have docker compose ready. [[Setup Guide]](https://docs.docker.com/engine/install/)

#### Running Tests with Docker Compose

To create the test environment run: `docker compose up -d --build`. This will build the image. You can then use commands like this to test the app:

```sh
curl -X GET -H "Content-type: application/json" \
                      -H "Accept: application/json" \
                      -H "X-Real-IP: 92.102.246.46" \
                      "http://localhost:3333/"
```
