package main

import (
	"fmt"
	"os"
	"facet-backend/pkg/password"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./cmd/genthash <password>")
		os.Exit(1)
	}
	h, err := password.HashPassword(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Println(h)
}
