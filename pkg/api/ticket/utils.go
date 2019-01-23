package ticket

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go"
)

var errInvalidOrder = errors.New("Invalid order ID provided")

func (i *Impl) getOrder(orderID string) (*stripe.Order, error) {
	if orderID == "" {
		return nil, errInvalidOrder
	}

	order, err := i.Stripe.Orders.Get(orderID, nil)
	if err != nil {
		return nil, err
	}

	// The order must have an item that contains our Informatics Ball SKU.
	// Items can also be tax and shipping (which for us is £0),
	// so we can't do a simple items[0] check (also checking for length).
	hasSKU := false

	for _, item := range order.Items {
		// todo: check if item.Parent always exists
		if item.Parent.ID == i.Config.Stripe.SKU {
			hasSKU = true
			break
		}
	}

	if !hasSKU {
		return nil, errInvalidOrder
	}

	if order.Status != string(stripe.OrderStatusPaid) {
		return nil, errors.New("This order must be paid for")
	}

	return order, nil
}

func (i *Impl) getRequestParams(c *gin.Context) map[string]string {
	return nil
}
