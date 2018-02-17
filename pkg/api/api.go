package api

import (
	"gopkg.in/mailgun/mailgun-go.v1"

	"github.com/compsoc-edinburgh/infball-api/pkg/api/base"
	"github.com/compsoc-edinburgh/infball-api/pkg/api/charge"
	"github.com/compsoc-edinburgh/infball-api/pkg/api/stats"
	"github.com/compsoc-edinburgh/infball-api/pkg/api/ticket"
	"github.com/compsoc-edinburgh/infball-api/pkg/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/client"
)

// NewAPI sets up a new API module.
func NewAPI(
	conf *config.Config,
	log *logrus.Logger,
) *base.API {

	router := gin.Default()
	router.Use(cors.Default())

	sc := &client.API{}
	sc.Init(conf.Stripe.SecretKey, nil)

	mg := mailgun.NewMailgun(conf.Mailgun.Domain, conf.Mailgun.APIKey, conf.Mailgun.PublicAPIKey)

	a := &base.API{
		Config:  conf,
		Log:     log,
		Stripe:  sc,
		Gin:     router,
		Mailgun: mg,
	}

	charge := charge.Impl{API: a}
	router.POST("/charge", charge.MakeCharge)

	ticket := ticket.Impl{API: a}
	router.GET("/ticket", ticket.Get)
	router.POST("/ticket", ticket.Post)

	stats := stats.Impl{API: a}
	router.GET("/stats", stats.Get)

	return a
}
