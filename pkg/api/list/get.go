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
	params.AddExpand("data.charge.balance_transaction")
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
			// todo: check if item.Parent always exists
			if item.Parent.ID == i.Config.Stripe.SKU {
				hasSKU = true
			}
		}

		if !hasSKU {
			continue
		}

		writer.Write([]string{
			o.ID,
			o.Metadata["owner_name"], o.Metadata["owner_email"],
			o.Metadata["uun"],
			o.Metadata["meal_starter"], o.Metadata["meal_main"], o.Metadata["meal_dessert"],
			o.Metadata["special_requests"],
			o.Metadata["purchaser_name"], o.Metadata["purchaser_email"],
			o.Metadata["over18"],
			o.Metadata["auth_token"],
			strconv.FormatInt(o.Charge.BalanceTransaction.Net, 10),
			strconv.FormatInt(o.Charge.BalanceTransaction.Fee, 10),
		})
	}

	writer.Flush()

	c.String(http.StatusOK, buf.String())
}
