# bitty-bingo

a small bingo application

The application runs as a web server that creates bingo boards as SVG images.
It also manages games, whose previous states can be reverted to.
Boards can be checked to determine if they has a bingo on the current game.

## Dependencies

[Go 1.17](https://golang.org/dl/) is used to build the application.

[Make](https://www.gnu.org/software/make/) is used by [Makefile](Makefile) to build the application.  This is not required, as the application can be build by entering commands manually into a terminal.

## Build

Run `make` to build the application.  This creates the `bitty-bingo` executable in the `build` folder.  Once built, the application can be moved somewhere else because it includes all resources to run the server.

To compile for Windows, run `make GO_ARGS="GOOS=windows" OBJ="bitty-bingo.exe"`.

To compile for other CPU architectures, use the `GOARCH` build flag. Example: `make GO_ARGS="GOOS=linux GOARCH=386"`.  Common values are `amd64`, and `386`.

## Testing

Run `make test` to run the tests for the application.

## Running

The application runs on the command line.  Run it with the `-h` parameter for information about the run-time arguments: `./build/bitty-bingo -h`

Examples:

* Run on specific ports: `./build/bitty-bingo --http-port=8001 --https-port=8000`

* Run on default HTTP ports with local HTTPS certificate: `sudo ./build/bitty-bingo --tls-cert-file=/home/jacobpatterson1549/tls-cert.pem --tls-key-file=/home/jacobpatterson1549/tls-key.pem`

  Use [mkcert](https://github.com/FiloSottile/mkcert) to create a local certificate.