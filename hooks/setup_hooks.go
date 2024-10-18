package hooks

import (
	"net/http"
	"newsletter/controllers"
	"newsletter/tools"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// Crear todos los hooks de pocketbase
func SetupHooks(app *pocketbase.PocketBase) {

	// Aqui se mapea el hashmap de los status dentro de la
	// aplicacion en la db y re agenda los correos.
	app.OnAfterBootstrap().Add(func(e *core.BootstrapEvent) error {
		tools.CreateSubscriberStatus(app)
		tools.CreateEmailStatus(app)

		tools.ScheduleEmailsAgain(app)

		return nil
	})

	// Setear todas las rutas
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.Validator = &CustomValidator{validator: validator.New()}
		controllers.SetupControllers(e, app)
		return nil
	})

	// Antes de crear un correo verificarlo
	app.OnRecordBeforeCreateRequest().Add(func(e *core.RecordCreateEvent) error {
		if e.Record.Collection().Name == "emails" {
			return VerifyEmailHook(e)
		}
		return nil
	})

	// Despues de crear un correo encolarlo y enviarlos
	app.OnRecordAfterCreateRequest().Add(func(e *core.RecordCreateEvent) error {
		if e.Record.Collection().Name == "emails" {
			return CreateMailsQueueAndSend(app, e)
		}
		return nil
	})

	// Antes de eliminar un correo, si estaba programado quitarlo del cron
	app.OnRecordBeforeDeleteRequest().Add(func(e *core.RecordDeleteEvent) error {
		if e.Record.Collection().Name == "emails" {
			return DeScheduleEmail(app, e)
		}
		return nil
	})

}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}
