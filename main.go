package main

import (
	"fmt"
	"net/http"
	"strings"
)

func main() {
	mux := http.NewServeMux()
	mux.Handler("/")
}
