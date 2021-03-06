# bitty-bingo

[![Docker Image CI](https://github.com/jacobpatterson1549/bitty-bingo/actions/workflows/go.yml/badge.svg)](https://github.com/jacobpatterson1549/bitty-bingo/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jacobpatterson1549/bitty-bingo)](https://goreportcard.com/report/github.com/jacobpatterson1549/bitty-bingo)
[![GoDoc](https://godoc.org/github.com/jacobpatterson1549/bitty-bingo?status.svg)](https://godoc.org/github.com/jacobpatterson1549/bitty-bingo)

## a small bingo application

The application runs as a web server that creates bingo board SVG images.
It also manages games, whose previous states can be reverted to.
Boards can be checked to verify if they have a "bingo" in a game.
Using a phone, the built in Barcode scanner simplifies checking boards.
The cell square on boards can be customized to be a QR, Aztec, or Data Matrix bar code.

## Screenshot

![bitty-bingo screenshot](screenshot.png)

## Dependencies

[Go 1.17](https://golang.org/dl/) is used to build the application.

[Make](https://www.gnu.org/software/make/) is used by [Makefile](Makefile) to build the application.  This is not required, as the application can be manually built by entering commands in a terminal.

## Build

Run `make` to build the application.  This creates a `bitty-bingo` executable in the `build` folder.  The application is very portable when built because it has no external dependencies.

To build for specifically for Windows, run `make GO_ARGS="GOOS=windows" OBJ="bitty-bingo.exe"`.

To build for other CPU architectures, use the `GOARCH` build flag. Example: `make GO_ARGS="GOOS=linux GOARCH=386"`.  Common values are `amd64`, and `386`.

### go dependencies

[boombuler/barcode](https://github.com/boombuler/barcode/) is used to generate bar codes for boards as the center "free cell"

## Testing

Run `make test` to run the tests for the application.

## Running

The application runs in a command-line terminal.  Run it with the `-h` parameter for information about the run-time arguments: `./build/bitty-bingo -h`.

TLS certificate public/private key files are needed to run the application.  If running on a local/trusted network (not the Internet), use [mkcert](https://github.com/FiloSottile/mkcert) to create the certificates.

The application runs two TCP servers on different ports.  The HTTP server redirects all traffic to the HTTPS server.
on startup.  

However, if a "PORT" environment variable is defined, the redirect server is not run.   The HTTPS server runs on the numeric value of the "PORT" variable.  The TLS certificates will also not be loaded.

Examples:

* Run on specific ports: `./build/bitty-bingo --http-port=8001 --https-port=8000`

* Run on default HTTP ports with local HTTPS certificate: `sudo ./build/bitty-bingo --tls-cert-file=/home/jacobpatterson1549/tls-cert.pem --tls-key-file=/home/jacobpatterson1549/tls-key.pem`

* Run only the HTTPS server, using managed TLS certificates: `sudo PORT=443 ./build/bitty-bingo`

* Special: If PORT is defined in a file named `.env` (`PORT=8000`), the server can be started in HTTPS-only mode with `make serve`