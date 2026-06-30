package main

import (
	"fmt"
	"os"

	"github.com/pedrosousa13/radio/internal/app"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println("radio", app.Version())
		return
	}
	if err := app.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
