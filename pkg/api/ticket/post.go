package ticket

import (
	"net/http"
	"net/mail"
	"strconv"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go"

	"github.com/compsoc-edinburgh/infball-api/pkg/api/base"
	"github.com/gin-gonic/gin"
)

func (i *Impl) Post(c *gin.Context) {
	var result struct {
		OrderID   string
		AuthToken string

		FullName string
		UUN      string
		Email    string
		Over18   bool
	}

	if err := c.BindJSON(&result); err != nil {
		base.BadRequest(c, err.Error())
		return
	}

	if result.OrderID == "" {
		base.BadRequest(c, "Order ID missing.")
		return
	}

	if !result.Over18 {
		base.BadRequest(c, "You must be atleast 18 years of age to attend.")
		return
	}

	if result.FullName == "" {
		base.BadRequest(c, "Full name missing.")
		return
	}

	toAddress := result.FullName + "<" + result.Email + ">"
	_, err := mail.ParseAddress(toAddress)
	if err != nil {
		base.BadRequest(c, "Invalid email format provided. Please email infball@comp-soc.com if this is a mistake.")
		return
	}

	order, err := i.getOrder(result.OrderID)
	if err != nil {
		base.BadRequest(c, base.StripeError(err))
		return
	}

	if order.Meta["auth_token"] != result.AuthToken {
		base.BadRequest(c, "Authorisation token does not match the code provided in your email.")
		return
	}

	if !base.CheckUUN(c, result.UUN) {
		return
	}

	authToken := result.AuthToken
	newToken := false
	if order.Meta["owner_email"] != result.Email {
		newToken = true
		authToken = uuid.New().String()
	}

	_, err = i.Stripe.Orders.Update(order.ID, &stripe.OrderUpdateParams{
		Params: stripe.Params{
			Meta: map[string]string{
				"owner_email": result.Email,
				"owner_name":  result.FullName,
				"uun":         result.UUN,
				"over18":      strconv.FormatBool(result.Over18),
				"auth_token":  authToken,
			},
		},
	})

	if err != nil {
		base.BadRequest(c, base.StripeError(err))
		return
	}

	if newToken {
		if !base.SendTicketEmail(c, i.Mailgun, result.FullName, toAddress, order.ID, authToken) {
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   true,
	})
}
