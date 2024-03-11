package api

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed templates/*
var templates embed.FS

// ListPages returns a list of page names in the templates directory.
func ListPages() []string {
	entries, err := templates.ReadDir("templates")
	if err != nil {
		panic(fmt.Errorf("read dir: %w", err))
	}
	var pages []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), "_") {
			continue
		}
		pages = append(pages, entry.Name())
	}
	return pages
}
