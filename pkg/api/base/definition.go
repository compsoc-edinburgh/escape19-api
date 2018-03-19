package base

import (
	"context"
	"net/http"
	"net/mail"
	"regexp"

	"github.com/badoux/checkmail"
	"github.com/compsoc-edinburgh/infball-api/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/client"
	mailgun "gopkg.in/mailgun/mailgun-go.v1"
)

// API contains all the dependencies of the API server
type API struct {
	Config  *config.Config
	Log     *logrus.Logger
	Gin     *gin.Engine
	Stripe  *client.API
	Mailgun mailgun.Mailgun

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

func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status":  "error",
		"message": msg,
	})
}

func IsOneOf(one string, other ...string) bool {
	for _, v := range other {
		if v == one {
			return true
		}
	}
	return false
}

func StripeError(err error) string {
	msg := err.Error()
	if stripeErr, ok := err.(*stripe.Error); ok {
		msg = stripeErr.Msg
	}
	return msg
}

// var uunRegex = regexp.MustCompile(`(s\d{7}|[a-zA-Z]{2,})`)
// the above uun does not account for numbers in staff uuns (which is possible at the end)
// and absolutely doesn't account for visitors (v1hreede)

var studentUUN = regexp.MustCompile(`^s\d{7}$`)

func CheckUUN(c *gin.Context, uun string) (success bool) {
	if uun != "" {
		checkEmail := uun + "@staffmail.ed.ac.uk"
		if studentUUN.MatchString(uun) {
			checkEmail = uun + "@sms.ed.ac.uk"
		}

		_, err := mail.ParseAddress(checkEmail)
		if err != nil {
			BadRequest(c, "Invalid uun provided. Please email infball@comp-soc.com if this is a mistake.")
			return
		}

		err = checkmail.ValidateHost(checkEmail)
		if smtpErr, ok := err.(checkmail.SmtpError); ok && err != nil && smtpErr.Code() == "550" {
			BadRequest(c, "Unknown UUN. Staff should put in the username they use to log in with EASE, not their @ed.ac.uk alias.")
			return
		} else if err != nil {
			BadRequest(c, "Something went wrong with UUN validation. Please contact infball@comp-soc.com for assistance.")
			return
		}
	}

	return true
}
