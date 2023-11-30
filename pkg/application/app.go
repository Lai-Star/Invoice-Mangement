package application

import (
	sentryiris "github.com/getsentry/sentry-go/iris"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
	"github.com/monetrapp/rest-api/pkg/config"
)

type Controller interface {
	RegisterRoutes(app *iris.Application)
}

func NewApp(configuration config.Configuration, controllers ...Controller) *iris.Application {
	app := iris.New()

	if configuration.Sentry.Enabled {
		app.Use(sentryiris.New(sentryiris.Options{
			Repanic: false,
		}))
	}

	app.UseRouter(cors.New(cors.Options{
		AllowedOrigins:  configuration.CORS.AllowedOrigins,
		AllowOriginFunc: nil,
		AllowedMethods: []string{
			"HEAD",
			"OPTIONS",
			"GET",
			"POST",
			"PUT",
			"DELETE",
		},
		AllowedHeaders: []string{
			"Cookies",
			"Content-Type",
			"H-Token",
		},
		ExposedHeaders:     nil,
		MaxAge:             0,
		AllowCredentials:   true,
		OptionsPassthrough: false,
		Debug:              configuration.CORS.Debug,
	}))

	for _, controller := range controllers {
		controller.RegisterRoutes(app)
	}

	return app
}
