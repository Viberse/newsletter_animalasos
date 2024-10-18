package controllers

import (
	"fmt"
	"net/http"
	"newsletter/tools"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/labstack/echo/v5"
	"github.com/madflojo/tasks"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
)

func ScheduleEmail(c echo.Context, app *pocketbase.PocketBase) error {
	dateStr := c.Request().FormValue("schedule_date")
	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "Falta la fecha o esta malformada")
	}

	offsetDate := date.Sub(time.Now())
	if offsetDate.Milliseconds() < 0 {
		return c.String(http.StatusBadRequest, "La fecha es menor a la actual")
	}

	dao := daos.New(app.Dao().DB())
	collection, err := dao.FindCollectionByNameOrId("emails")
	if err != nil {
		return err
	}
	record := models.NewRecord(collection)
	form := forms.NewRecordUpsert(app, record)
	if err := form.LoadRequest(c.Request(), ""); err != nil {
		return nil
	}

	if err := tools.VerifyEmail(aws.String(form.Data()["template_name"].(string))); err != nil {
		return nil
	}

	if err := form.Submit(); err != nil {
		return err
	}

	emailID := record.Id

	fmt.Printf("El correo %s se enviara en %v\n", emailID, offsetDate)

	if scheduleID, err := tools.Scheduler.Add(&tasks.Task{
		Interval: offsetDate,
		RunOnce:  true,
		TaskFunc: func() error {
			tools.QueueEmail(emailID, app)
			return tools.ScheduleDequeueTask(app, emailID)
		},
	}); err != nil {
		dao.Delete(record)
		return c.String(http.StatusInternalServerError, "Ocurrio un error programando el correo.")
	} else {
		record.Set("schedule_id", scheduleID)
		dao.SaveRecord(record)
	}

	return nil
}
