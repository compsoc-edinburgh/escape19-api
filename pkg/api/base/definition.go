package base

import (
	"context"
	"net/http"

	"github.com/compsoc-edinburgh/infball-api/pkg/config"
	"github.com/stripe/stripe-go/client"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// API contains all the dependencies of the API server
type API struct {
	Config *config.Config
	Log    *logrus.Logger
	Gin    *gin.Engine
	Stripe *client.API

	Server *http.Server
}

// Start binds the API and starts listening.
func (a *API) Start() error {
	a.Server = &http.Server{
		Addr:    a.Config.BindAddress,
		Handler: a.Gin,
	}
	return a.Server.ListenAndServe()
}

// Shutdown shuts down the API
func (a *API) Shutdown(ctx context.Context) error {
	if err := a.Server.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
