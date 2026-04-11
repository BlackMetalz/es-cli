package main

import (
	"fmt"

	"github.com/kienlt/es-cli/internal/es"
)

func main() {
	c := es.NewClient("http://localhost:9200", "elastic", "elastic")
	indices, err := c.ListIndices()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	for _, idx := range indices {
		fmt.Printf("name=%-20s docs=%-8s size=[%s]\n", idx.Name, idx.DocsCount, idx.StoreSize)
	}
}
