package web

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/slim-crown/issue-1-website/pkg/issue1.REST.client/issue1"

	"github.com/julienschmidt/httprouter"
)

// Setup is used to inject dependencies and other required data used by the handlers.
type Setup struct {
	Config
	Dependencies
	templates *template.Template
}

// Dependencies contains dependencies used by the handlers.
type Dependencies struct {
	Iss1C  *issue1.Client
	Logger *log.Logger
}

// Config contains the different settings used to set up the handlers
type Config struct {
	TemplatesStoragePath, AssetStoragePath, AssetServingRoute, HostAddress, Port string
	TokenAccessLifetime, TokenRefreshLifetime                                    time.Duration
	TokenSigningSecret                                                           []byte
}

// NewMux returns a fully configured issue1 website server.
func NewMux(s *Setup) *httprouter.Router {
	mainRouter := httprouter.New()

	s.templates = template.Must(template.ParseGlob(s.TemplatesStoragePath))

	fs := http.FileServer(http.Dir(s.AssetStoragePath))
	mainRouter.Handler("GET", s.AssetServingRoute+"*filepath", http.StripPrefix(s.AssetServingRoute, fs))

	mainRouter.HandlerFunc("GET", "/", getFront(s))

	return mainRouter
}
