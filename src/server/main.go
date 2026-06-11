package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/drabenstadtj/syncraft/src/internal/server"
)

func main() {
	port := flag.Int("port", 8080, "port to listen on")
	data := flag.String("data", "data", "directory to store diffs")
	flag.Parse()

	srv := server.New(*data)

	addr := fmt.Sprintf(":%d", *port)
	fmt.Fprintf(os.Stdout, "listening on %s, storing diffs in %s\n", addr, *data)

	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
