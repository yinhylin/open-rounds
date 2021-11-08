# building

```
go build
```

# running

# Local Client (spins up server on its own if it can't connect)
`go run .`

# Standalone server
`go run . server`

# Web Browser
## Pre-requisite
```
go install github.com/hajimehoshi/wasmserve@latest
```

## Start the server
`go run . server`

## Serve the game at localhost:8080
`wasmserve .`

That's it. Connect to `localhost:8080` and you should be ready to play.

