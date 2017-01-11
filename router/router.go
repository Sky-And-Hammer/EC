package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Sky-And-Hammer/wildcardRouter"
)

var rootMux *http.ServeMux
var wcRouter *wildcardRouter.WildcardRouter

func main() {
	if rootMux == nil {
		router := gin.Default()
		router.Use(func (ctx *gin.Context) {
			tx := db.
		})
	}
}
