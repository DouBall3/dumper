# dumper

A minimal HTTP server that returns `200 OK` with an empty body for every request, regardless of method or path.

Run with Go 1.22 or newer:

```sh
go run . -ip 127.0.0.1 -port 8080
```

Use `-ip 0.0.0.0` to listen on all IPv4 interfaces, or `-ip ::` for IPv6. Defaults are `127.0.0.1:8080`.

Each request is written to standard output, starting with an RFC 3339 timestamp followed by the request line, headers, and full body. Three empty lines follow each request. Concurrent requests are logged as separate, uninterrupted entries. Startup messages and errors go to standard error.

The log uses Go's parsed HTTP request representation; it is not a byte-for-byte capture of the network traffic. Requests must be valid HTTP for Go's HTTP server to handle them.

CORS permits browser requests from `*`, using GET, POST, PUT, PATCH, DELETE, HEAD, or OPTIONS and the `Content-Type` and `Authorization` headers. Preflight OPTIONS requests are logged and return `200 OK` like other requests.

Build a standalone executable:

```sh
go build -o dumper .
./dumper -ip 127.0.0.1 -port 8080
```

## Docker Hub publishing

The `Publish Docker image` GitHub Actions workflow builds the root `Dockerfile`
and publishes `docker.io/douball/dumper:latest` on pushes to the repository's
default branch. It can also be run manually from the Actions tab on that branch.

Before running it, add a repository Actions secret named `DOCKERHUB_TOKEN`
containing a Docker Hub access token for `douball` with write access to the
`douball/dumper` repository.
