package api

import (
	"github.com/compsoc-edinburgh/infball-api/pkg/api/base"
	"github.com/compsoc-edinburgh/infball-api/pkg/api/charge"
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

	a := &base.API{
		Config: conf,
		Log:    log,
		Stripe: sc,
		Gin:    router,
	}

	charge := charge.Impl{API: a}
	router.POST("/charge", charge.MakeCharge)

	return a
}
