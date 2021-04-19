package controller

import (
	"github.com/monetrapp/rest-api/pkg/jobs"
	"github.com/monetrapp/rest-api/pkg/metrics"
	"net/http"
	"net/smtp"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/core/router"
	"github.com/monetrapp/rest-api/pkg/config"
	"github.com/plaid/plaid-go/plaid"
	"github.com/sirupsen/logrus"
	"github.com/xlzd/gotp"
	"gopkg.in/ezzarghili/recaptcha-go.v4"
)

const (
	TokenName = "H-Token"
)

type Controller struct {
	db             *pg.DB
	configuration  config.Configuration
	captcha        *recaptcha.ReCAPTCHA
	plaid          *plaid.Client
	smtp           *smtp.Client
	mailVerifyCode *gotp.HOTP
	log            *logrus.Entry
	job            jobs.JobManager
	stats          *metrics.Stats
}

func NewController(
	configuration config.Configuration,
	db *pg.DB,
	job jobs.JobManager,
	plaidClient *plaid.Client,
	stats *metrics.Stats,
) *Controller {
	logger := logrus.New()
	level, err := logrus.ParseLevel(configuration.Logging.Level)
	if err != nil {
		panic(err)
	}
	logger.SetLevel(level)
	entry := logrus.NewEntry(logger)
	var captcha recaptcha.ReCAPTCHA
	if configuration.ReCAPTCHA.Enabled {
		captcha, err = recaptcha.NewReCAPTCHA(
			configuration.ReCAPTCHA.PrivateKey,
			recaptcha.V2,
			30*time.Second,
		)
		if err != nil {
			panic(err)
		}
	}

	return &Controller{
		captcha:       &captcha,
		configuration: configuration,
		db:            db,
		plaid:         plaidClient,
		log:           entry,
		job:           job,
		stats:         stats,
	}
}

// @title monetr's REST API
// @version 0.0
// @description This is the REST API for our budgeting application.

// @contact.name Support
// @contact.url http://github.com/monetrapp/rest-api

// @license.name Business Source License 1.1
// @license.url https://github.com/monetrapp/rest-api/blob/main/LICENSE

// @host api.monetr.app

// @tag.name Funding Schedules
// @tag.description Funding Schedules are created by the user to tell us when money should be taken from their account to fund their goals and expenses.

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name H-Token
func (c *Controller) RegisterRoutes(app *iris.Application) {
	if c.stats != nil {
		app.UseGlobal(func(ctx *context.Context) {
			start := time.Now()
			defer func() {
				c.stats.FinishedRequest(ctx, time.Since(start))
			}()

			ctx.Next()
		})
	}

	app.Use(c.loggingMiddleware)
	app.OnAnyErrorCode(func(ctx *context.Context) {
		if err := ctx.GetErr(); err != nil {
			ctx.JSON(map[string]interface{}{
				"error": err.Error(),
			})
		}
	})
	app.OnErrorCode(http.StatusNotFound, func(ctx *context.Context) {
		if err := ctx.GetErr(); err == nil {
			ctx.JSON(map[string]interface{}{
				"path":  ctx.Path(),
				"error": "the requested path does not exist",
			})
		} else {
			ctx.JSON(map[string]interface{}{
				"error": err.Error(),
			})
		}
	})

	app.Get("/health", c.getHealth)

	if c.configuration.EnableWebhooks {
		// Webhooks use their own authentication, so we want to declare this first.
		app.Post("/api/plaid/webhook/{identifier:string}", c.handlePlaidWebhook)
	}

	// For the following endpoints we want to have a repository available to us.
	app.PartyFunc("/", func(p router.Party) {
		p.Get("/config", c.configEndpoint)

		p.Use(c.setupRepositoryMiddleware)

		p.PartyFunc("/authentication", func(p router.Party) {
			p.Post("/login", c.loginEndpoint)
			p.Post("/register", c.registerEndpoint)
			p.Post("/verify", c.verifyEndpoint)
		})

		p.Use(c.authenticationMiddleware)

		p.PartyFunc("/users", func(p router.Party) {
			p.Get("/me", c.meEndpoint)
		})

		p.PartyFunc("/links", c.linksController)

		p.PartyFunc("/plaid/link", c.handlePlaidLinkEndpoints)

		p.PartyFunc("/bank_accounts", func(p router.Party) {
			c.handleBankAccounts(p)
			c.handleTransactions(p)
			c.handleFundingSchedules(p)
			c.handleSpending(p)
		})

		p.PartyFunc("/jobs", c.handleJobs)
	})
}

// Check API Health
// @Summary Check API Health
// @ID api-health
// @tags Health
// @description Just a simple health check endpoint. This is not used at all in the frontend of the application and is meant to be used in containers to determine if the primary API listener is working.
// @Produce json
// @Router /health [get]
// @Success 200 {object} swag.HealthResponse
func (c *Controller) getHealth(ctx *context.Context) {
	err := c.db.Ping(ctx.Request().Context())

	ctx.JSON(map[string]interface{}{
		"dbHealthy":  err == nil,
		"apiHealthy": true,
	})
}
