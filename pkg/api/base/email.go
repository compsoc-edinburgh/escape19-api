package base

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	mailgun "gopkg.in/mailgun/mailgun-go.v1"
)

type EmailStruct struct {
	Name      string
	AuthToken string
	OrderID   string
}

var Email *template.Template = template.Must(template.ParseFiles("email.html"))

func SendTicketEmail(c *gin.Context, mailgun mailgun.Mailgun, name, to_address, orderID, authToken string) (_ bool) {
	var tpl bytes.Buffer
	if err := Email.Execute(&tpl, EmailStruct{
		Name:      name,
		OrderID:   orderID,
		AuthToken: authToken,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Ticket was bought but email generation was unsuccessful. Please email us for assistance.",
		})
		return
	}

	message := mailgun.NewMessage(
		"SIGINT <contact@sigint.mx>",
		"Escape Room with SIGINT ticket! [#"+orderID+"]",
		"",
		to_address,
	)
	message.SetHtml(tpl.String())

	_, _, err := mailgun.Send(message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Ticket was bought but an email was not sent. Please email us for assistance.",
		})
		return
	}

	return true
}
