- Install Docker if you haven't.

- Build the Docker image: `docker build -t myredis .`

- Run the Docker container: `docker run -p 6379:6379 myredis`

- Install Go if you haven't.

- Install the Go dependencies: `go get`

- Run the Go server: `go run server.go --origin https://jsonplaceholder.typicode.com --port 3000`

- Make a GET request with curl or Postman: `curl http://localhost:3000/posts`

- If you want to clear the cache: `go run server.go --clear-cache`
