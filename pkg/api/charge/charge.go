package charge

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/badoux/checkmail"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/currency"
)

var uunRegex = regexp.MustCompile(`(s\d{7}|[a-zA-Z]{2,})`)
var studentUUN = regexp.MustCompile(`s\d{7}`)

func badRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status":  "error",
		"message": msg,
	})
}

func isOneOf(one string, other ...string) bool {
	for _, v := range other {
		if v == one {
			return true
		}
	}
	return false
}

func (i *Impl) MakeCharge(c *gin.Context) {
	var result struct {
		Token string

		FullName    string
		UUN         string
		Email       string
		Underage    bool
		Starter     string
		Main        string
		Dessert     string
		SpecialReqs string
	}

	if err := c.BindJSON(&result); err != nil {
		badRequest(c, err.Error())
		return
	}

	if result.Token == "" {
		badRequest(c, "token missing")
		return
	}

	if result.FullName == "" {
		badRequest(c, "missing full name")
		return
	}

	if result.UUN != "" && uunRegex.MatchString(result.UUN) {
		badRequest(c, "Invalid UUN. Please contact infball@comp-soc.com for assistance.")
		return
	} else if result.UUN != "" {
		checkEmail := result.UUN + "@staffmail.ed.ac.uk"
		if studentUUN.MatchString(result.UUN) {
			checkEmail = result.UUN + "@sms.ed.ac.uk"

			err := checkmail.ValidateHost(checkEmail)
			if smtpErr, ok := err.(checkmail.SmtpError); ok && err != nil && smtpErr.Code() == "550" {
				fmt.Printf("Code: %s, Msg: %s", smtpErr.Code(), smtpErr)
			}
		}
	}

	if !isOneOf(result.Starter, "soup", "salmon", "pork") || !isOneOf(result.Main, "beef", "salmon", "chicken", "mushrooms") || !isOneOf(result.Dessert, "brownie") || !isOneOf(result.Dessert, "toffee") {
		badRequest(c, "Invalid food selection.")
		return
	}

	if len(result.SpecialReqs) > 500 {
		badRequest(c, "Sorry, your request is limited to 500 characters. Please email infball@comp-soc.com for assistance.")
		return
	}

	sku, err := i.Stripe.Skus.Get(i.Config.Stripe.SKU, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.(*stripe.Error).Msg,
		})
		return
	}

	if sku.Inventory.Quantity == 0 {
		c.JSON(http.StatusGone, gin.H{
			"status":  "error",
			"message": "Sorry! We have run out of tickets... for now.",
		})
		return
	}

	ticketID := uuid.New().String()

	order, err := i.Stripe.Orders.New(&stripe.OrderParams{
		Currency: currency.GBP,
		Items: []*stripe.OrderItemParams{
			&stripe.OrderItemParams{
				Type:   "sku",
				Parent: i.Config.Stripe.SKU,
			},
		},
		Params: stripe.Params{
			Meta: map[string]string{
				"uun":              result.UUN,
				"email":            result.Email,
				"underage":         strconv.FormatBool(result.Underage),
				"meal_starter":     result.Starter,
				"meal_main":        result.Main,
				"meal_dessert":     result.Dessert,
				"special_requests": result.SpecialReqs,
				"id":               ticketID,
			},
		},
		Email: result.Email,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.(*stripe.Error).Msg,
		})
		return
	}

	// Charge the user's card:
	params := &stripe.OrderPayParams{}
	params.SetSource(result.Token)

	o, err := i.Stripe.Orders.Pay(order.ID, params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.(*stripe.Error).Msg,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   o,
	})
}
