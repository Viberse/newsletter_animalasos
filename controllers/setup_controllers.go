package controllers

import (
	"net/http"
	"newsletter/middlewares"
	"newsletter/tools"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func createAuthRoute(method string, path string, handler echo.HandlerFunc, app *pocketbase.PocketBase) echo.Route {
	return echo.Route{
		Method:  method,
		Path:    path,
		Handler: handler,
		Middlewares: []echo.MiddlewareFunc{
			middlewares.AuthOnly(app),
		},
	}
}

func SetupControllers(e *core.ServeEvent, app *pocketbase.PocketBase) {
	// Cargar el directorio publico con la pagina de admin
	e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS(tools.DEFAULT_PUBLIC_DIR), true))
	e.Router.GET("/principal/*", apis.StaticDirectoryHandler(os.DirFS("./principal"), true))

	e.Router.GET("/verificar/:token", func(c echo.Context) error {
		return VerifyToken(c, app)
	})

	e.Router.GET("/desuscribirse/:token", func(c echo.Context) error {
		return Unsubscribe(c, app)
	})

	e.Router.GET("/query_nft_metadata/:pagina", func(c echo.Context) error {
		return QueryNftMetadata(c, app)
	})

	e.Router.POST("/subscribirse", func(c echo.Context) error {
		return Subscribe(c, app)
	})

	e.Router.GET("/clicks/:id", func(c echo.Context) error {
		return ClicksTracker(c, app)
	})

	e.Router.GET("/api/opensea/events", func(c echo.Context) error {
		return GetCollectionEvents(c, app)
	})

	e.Router.GET("/api/opensea/stats", func(c echo.Context) error {
		return GetCollectionStats(c, app)
	})

	// privadas
	e.Router.AddRoute(createAuthRoute(http.MethodPost, "/schedule", func(c echo.Context) error {
		return ScheduleEmail(c, app)
	}, app))

	e.Router.AddRoute(createAuthRoute(http.MethodPost, "/resend", func(c echo.Context) error {
		return ResendEmail(c, app)
	}, app))

	e.Router.AddRoute(createAuthRoute(http.MethodGet, "/subs_per_month/:year/:month", func(c echo.Context) error {
		return SubsPerMonth(c, app)
	}, app))

	e.Router.AddRoute(createAuthRoute(http.MethodGet, "/subs_status_count", func(c echo.Context) error {
		return SubsStatusCount(c, app)
	}, app))

	e.Router.AddRoute(createAuthRoute(http.MethodGet, "/emails_ses_stadistics/:year/:month", func(c echo.Context) error {
		return SesStadistics(c, app)
	}, app))

	e.Router.AddRoute(createAuthRoute(http.MethodPost, "/edit_schedule", func(c echo.Context) error {
		return EditSchedule(c, app)
	}, app))

	e.Router.AddRoute(createAuthRoute(http.MethodPost, "/schedule_loop", func(c echo.Context) error {
		return ScheduleLoop(c, app)
	}, app))

	e.Router.AddRoute(createAuthRoute(http.MethodGet, "/template/:name", func(c echo.Context) error {
		return GetTemplate(c, app)
	}, app))

	e.Router.AddRoute(createAuthRoute(http.MethodGet, "/templates", func(c echo.Context) error {
		return ListTemplates(c, app)
	}, app))

	e.Router.AddRoute(createAuthRoute(http.MethodDelete, "/template/:name", func(c echo.Context) error {
		return DeleteTemplate(c, app)
	}, app))

	e.Router.AddRoute(createAuthRoute(http.MethodPost, "/template", func(c echo.Context) error {
		return CreateTemplate(c, app)
	}, app))

	e.Router.AddRoute(createAuthRoute(http.MethodPut, "/template", func(c echo.Context) error {
		return UpdateTemplate(c, app)
	}, app))

	e.Router.AddRoute(createAuthRoute(http.MethodPost, "/upload_metadata", func(c echo.Context) error {
		return UploadMetadas(c, app)
	}, app))

	e.Router.AddRoute(createAuthRoute(http.MethodPost, "/delete_metadatas", func(c echo.Context) error {
		return DeleteMetadatas(c, app)
	}, app))
}
