package list

import (
	"bytes"
	"encoding/csv"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	stripe "github.com/stripe/stripe-go"
)

func (i *Impl) Get(c *gin.Context) {
	if c.Query("pw") != i.Config.StatsPass {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorised.",
		})
		return
	}

	params := &stripe.OrderListParams{}
	params.Expand("data.charge.balance_transaction")
	params.Filters.AddFilter("status", "", "paid")

	// day before orders went out, utc timestamp
	params.Filters.AddFilter("created", "gt", "1518714907")

	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	writer.Write([]string{
		"order_id",
		"owner_name", "owner_email",
		"uun",
		"meal_starter", "meal_main", "meal_dessert",
		"meal",
		"special_requests",
		"purchaser_name", "purchaser_email",
		"over18",
		"auth_token",
		"charge_net",
		"charge_fees",
	})

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

		writer.Write([]string{
			o.ID,
			o.Meta["owner_name"], o.Meta["owner_email"],
			o.Meta["uun"],
			o.Meta["meal_starter"], o.Meta["meal_main"], o.Meta["meal_dessert"],
			o.Meta["meal"],
			o.Meta["special_requests"],
			o.Meta["purchaser_name"], o.Meta["purchaser_email"],
			o.Meta["over18"],
			o.Meta["auth_token"],
			strconv.FormatInt(o.Charge.Tx.Net, 10),
			strconv.FormatInt(o.Charge.Tx.Fee, 10),
		})
	}

	writer.Flush()

	c.String(http.StatusOK, buf.String())
}
