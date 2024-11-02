# Reverse proxy

Simple reverse proxy plus HTTP server (1.1) for study purposes.

## Installation

- Git [clone the repo](https://github.com/Guilospanck/reverse-proxy.git);

- Run `go mod tidy` to install dependencies;

## Running

Run `go run .`

There will be available four endpoints:

- `0.0.0.0:3000`, server A;
- `0.0.0.0:4000`, server B;
- `0.0.0.0:5000`, server C;
- `0.0.0.0:6000`, proxy server.

To access the servers via the proxy, use the corresponding paths:

- `/a`, for server A;
- `/b`, for server B;
- `/c`, for server C.

Example:

```shell
curl http://0.0.0.0:6000/a
```
