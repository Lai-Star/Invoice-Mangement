package controller

import (
	"github.com/kataras/iris/v12/context"
)

// Application Configuration
// @Summary Get Config
// @tags Config
// @id app-config
// @description Provides the configuration that should be used by the frontend application or UI.
// @Produce json
// @Router /config [get]
// @Success 200
func (c *Controller) configEndpoint(ctx *context.Context) {
	var config struct {
		RequireLegalName    bool   `json:"requireLegalName"`
		RequirePhoneNumber  bool   `json:"requirePhoneNumber"`
		VerifyLogin         bool   `json:"verifyLogin"`
		VerifyRegister      bool   `json:"verifyRegister"`
		ReCAPTCHAKey        string `json:"ReCAPTCHAKey,omitempty"`
		StripePublicKey     string `json:"stripePublicKey,omitempty"`
		AllowSignUp         bool   `json:"allowSignUp"`
		AllowForgotPassword bool   `json:"allowForgotPassword"`
		LongPollPlaidSetup  bool   `json:"longPollPlaidSetup"`
	}

	// If ReCAPTCHA is enabled then we want to provide the UI our public key as
	// well as whether or not we want it to verify logins and registrations.
	if c.configuration.ReCAPTCHA.Enabled {
		config.ReCAPTCHAKey = c.configuration.ReCAPTCHA.PublicKey
		config.VerifyLogin = c.configuration.ReCAPTCHA.VerifyLogin
		config.VerifyRegister = c.configuration.ReCAPTCHA.VerifyRegister
	}

	// We can only allow forgot password if SMTP is enabled. Otherwise we have
	// no way of sending an email to the user.
	if c.configuration.SMTP.Enabled {
		config.AllowForgotPassword = true
	}

	config.AllowSignUp = c.configuration.AllowSignUp

	if c.configuration.Plaid.EnableReturningUserExperience {
		config.RequireLegalName = true
		config.RequirePhoneNumber = true
	}

	if c.configuration.Stripe.Enabled {
		config.StripePublicKey = c.configuration.Stripe.PublicKey
	}

	// Just make this true for now, this might change in the future as I do websockets.
	config.LongPollPlaidSetup = true

	ctx.JSON(config)
}
