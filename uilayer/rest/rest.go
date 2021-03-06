package rest

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/tonyalaribe/ninja/core"
)

type Server struct {
	core core.Manager
}

func Register(manager core.Manager) error {
	server := &Server{
		core: manager,
	}
	server.Run()
	return nil
}

func (server *Server) Run() {
	port := "8082"
	baseCtx := context.Background()
	router := server.Routes()

	if err := chi.Walk(router, ChiWalkFunc); err != nil {
		log.Panicf("⚠️  Logging err: %s\n", err.Error())
	}

	srv := http.Server{
		Addr:    ":" + port,
		Handler: chi.ServerBaseContext(baseCtx, router),
	}

	idleConnsClosed := make(chan struct{})
	go ShutdownOnNotify(baseCtx, &srv, idleConnsClosed)

	log.Printf("Serving at 🔥 :%s \n", port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Printf("HTTP server ListenAndServe: %v", err)
	}
	<-idleConnsClosed
}

type ResponseResource struct {
	Code  int         `json:"code,omitempty"`
	Error string      `json:"error,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}

func ResponseWrapper(f func(w http.ResponseWriter, r *http.Request) (interface{}, int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responseData, statusCode, err := f(w, r)

		resp := ResponseResource{
			Code: statusCode,
			Data: responseData,
		}
		if err != nil {
			resp.Error = err.Error()
		}
		render.Status(r, statusCode)
		render.JSON(w, r, resp)
	}
}

func ShutdownOnNotify(ctx context.Context, srv *http.Server, idleConnsClosed chan struct{}) {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	// We received an interrupt signal, shut down.
	log.Println("😔 Shutting down. Goodbye..")
	if err := srv.Shutdown(ctx); err != nil {
		// Error from closing listeners, or context timeout:
		log.Fatalf("⚠️  HTTP server ListenAndServe error: %v", err)
	}
	close(idleConnsClosed)
}

func ChiWalkFunc(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
	log.Printf("👉 %s %s\n", method, route)
	return nil
}

func ResponseMessage(statusCode int, message string) map[string]interface{} {
	resp := make(map[string]interface{})
	resp["code"] = statusCode
	resp["message"] = message
	return resp
}
