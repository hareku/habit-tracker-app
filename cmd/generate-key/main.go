package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gorilla/securecookie"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("invalid arg len. got %d", len(os.Args))
	}

	if err := run(os.Args[1]); err != nil {
		log.Fatal(err)
	}
}

func run(name string) error {
	if _, err := os.Stat(name); err != nil && os.IsNotExist(err) {
		f, err := os.Create(name)
		if err != nil {
			return fmt.Errorf("create a file: %w", err)
		}
		defer f.Close()
		if _, err := f.Write(securecookie.GenerateRandomKey(64)); err != nil {
			return fmt.Errorf("write key: %w", err)
		}
	}
	return nil
}
