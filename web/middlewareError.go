package web

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

func errorMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer (func() {
			if err := recover(); err != nil {
				log.Errorf("Internal error: %s", err)
			}
		})()

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})

}

func (s webServer) initErrorMiddleware() {
	s.router.Use(errorMiddleware)
}
