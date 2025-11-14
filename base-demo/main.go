package main

import (
	"fmt"
	"net"
	"os"

	"github.com/lyonmu/demo/base-demo/internal/gin"
	"github.com/soheilhy/cmux"
)

func main() {

	l, err := net.Listen("tcp", ":9024")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen: %v", err)
		os.Exit(1)
	}

	m := cmux.New(l)

	if err := gin.NewGin(m); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create gin engine: %v", err)
		os.Exit(1)
	}

	if err := m.Serve(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to serve: %v", err)
		os.Exit(1)
	}
}
