package web

import (
	"github.com/rs/cors"
	config "github.com/spf13/viper"
	"net/http"
)

func (s webServer) initCORSMiddleware() {
	corsAllowedOrigins := config.GetStringSlice("web.cors.allowedOrigins")
	corsAllowedHeaders := config.GetStringSlice("web.cors.allowedHeaders")
	corsAllowCredentials := config.GetBool("web.cors.allowCredentials")

	corsOptions := cors.Options{
		AllowedMethods: []string{
			http.MethodOptions,
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   corsAllowedHeaders,
		AllowedOrigins:   corsAllowedOrigins,
		AllowCredentials: corsAllowCredentials,
	}

	c := cors.New(corsOptions)

	s.router.Use(c.Handler)
}
