package config

type Config struct {
	LogLevel string `default:"debug"`

	// Stripe
	Stripe StripeConfig

	// Bind Address
	BindAddress string `default:"0.0.0.0:8080"`

	// Mailgun
	Mailgun MailgunConfig

	// Pass
	StatsPass string `required:"true"`
}

// StripeConfig contains all configuration data for a CoSign connection
type StripeConfig struct {
	PublishableKey string `required:"true"`
	SecretKey      string `required:"true"`
	Product        string `required:"true"` // stripe product
	SKU            string `required:"true"` // stripe SKU
}

// Token is an "API" user
type Token struct {
	Name string `required:"true"`
	Key  string `required:"true"`
}

type MailgunConfig struct {
	Domain       string `required:"true"`
	APIKey       string `required:"true"`
	PublicAPIKey string `required:"true"`
}
