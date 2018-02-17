package stats

import (
	"encoding/xml"
	"net/http"

	"github.com/gin-gonic/gin"
	stripe "github.com/stripe/stripe-go"
)

func isOneOf(one string, other ...string) bool {
	for _, v := range other {
		if v == one {
			return true
		}
	}
	return false
}

func (i *Impl) Get(c *gin.Context) {
	if c.Query("pw") != i.Config.StatsPass {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorised.",
		})
		return
	}

	var data struct {
		XMLName   xml.Name `xml:"stats"`
		BeefCount int      `xml:"beefs"`
		Total     int      `xml:"payments"`
		FeeTotal  int64    `xml:"fees"`
	}

	params := &stripe.OrderListParams{}
	params.Expand("data.charge.balance_transaction")
	params.Filters.AddFilter("status", "", "paid")

	// day before orders went out, utc timestamp
	params.Filters.AddFilter("created", "gt", "1518714907")

	orders := i.Stripe.Orders.List(params)
	for orders.Next() {
		o := orders.Order()

		hasSKU := false
		for _, item := range o.Items {
			if item.Parent == i.Config.Stripe.SKU {
				hasSKU = true
			}
		}

		if !hasSKU {
			continue
		}

		data.Total++
		if o.Meta["meal_main"] == "beef" {
			data.BeefCount++
		}

		data.FeeTotal += o.Charge.Tx.Fee

	}

	c.XML(http.StatusOK, &data)
}
