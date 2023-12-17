package web

import (
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type webServer struct {
	host   string
	port   string
	server *http.Server
	router *mux.Router
}

func NewWebServer(host string, port string) (s *webServer) {

	// Create a mux router.
	r := mux.NewRouter()

	// Init webServer wrapper struct.
	s = &webServer{router: r, host: host, port: port}

	// Error
	s.initErrorMiddleware()

	// Logging.
	s.initLoggingMiddleware()

	// CORS.
	s.initCORSMiddleware()

	// HTTP routing.
	s.initRouting()

	return
}

func (s webServer) Start() {

	// Create HTTP server.
	s.server = &http.Server{
		Handler:      s.router,
		Addr:         fmt.Sprintf("%s:%s", s.host, s.port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	if err := s.server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
