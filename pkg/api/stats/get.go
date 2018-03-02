package stats

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (i *Impl) Get(c *gin.Context) {
	sku, err := i.Stripe.Skus.Get(i.Config.Stripe.SKU, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Unknown error.",
		})
	}

	c.JSON(http.StatusOK, struct{ Quantity int64 }{sku.Inventory.Quantity})
}
