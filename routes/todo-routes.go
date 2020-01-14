package routes

import (
	"github.com/kwilmot/go-todo/handlers"
	"net/http"
)

func SetTodoRoutes() {
	http.HandleFunc("/todos", handlers.TodosController)
}
