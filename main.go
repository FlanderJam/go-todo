package main

import (
	"github.com/kwilmot/go-todo/handlers"
	"github.com/kwilmot/go-todo/utils"
	"log"
	"net/http"
)

type Api struct {
	// We could use http.Handler as a type here; using the specific type has
	// the advantage that static analysis tools can link directly from
	// h.UserHandler.ServeHTTP to the correct definition. The disadvantage is
	// that we have slightly stronger coupling. Do the tradeoff yourself.
	TodoHandler *handlers.TodoHandler
}

func (h *Api) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string
	head, req.URL.Path = utils.ShiftPath(req.URL.Path)
	if head == "todos" {
		h.TodoHandler.ServeHTTP(res, req)
		return
	}
	http.Error(res, "Not Found", http.StatusNotFound)
}

func main() {
	api := &Api{TodoHandler: new(handlers.TodoHandler)}
	log.Fatal(http.ListenAndServe(":10000", api))
}
