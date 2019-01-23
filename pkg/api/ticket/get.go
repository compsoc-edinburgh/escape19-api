package ticket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/qaisjp/infball-api/pkg/api/base"
)

var exposableKeys = []string{
	"purchaser_email", "owner_email",
	"purchaser_name", "owner_name",
	"meal_dessert", "meal_main", "meal_starter",
	"over18", "uun", "special_requests",
}

// Returns the parameters associated with a particular order
func (i *Impl) Get(c *gin.Context) {
	orderID := c.Query("id")
	order, err := i.getOrder(orderID)
	if err != nil {
		base.BadRequest(c, base.StripeError(err))
		return
	}

	data := make(map[string]string)
	for _, key := range exposableKeys {
		data[key] = order.Metadata[key]
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   data,
	})
}
