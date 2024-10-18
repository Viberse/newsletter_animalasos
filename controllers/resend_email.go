package controllers

import (
	"fmt"
	"net/http"
	"newsletter/tools"
	"path"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/madflojo/tasks"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

type copyEmail struct {
	EmailID      string `json:"emailId" validate:"required"`
	Subject      string `json:"subject" validate:"required"`
	ScheduleDate string `json:"scheduleDate" validate:"required"`
}

func ResendEmail(c echo.Context, app *pocketbase.PocketBase) error {
	copyEmail := new(copyEmail)

	if err := c.Bind(copyEmail); err != nil {
		return c.String(http.StatusBadRequest, "Invalid body")
	}
	if err := c.Validate(copyEmail); err != nil {
		return c.String(http.StatusBadRequest, "Invalid body data")
	}

	date, err := time.Parse(time.RFC3339, copyEmail.ScheduleDate)
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

	fromEmail, err := dao.FindRecordById(collection.Id, copyEmail.EmailID)
	if err != nil {
		return err
	}

	record := models.NewRecord(collection)
	form := forms.NewRecordUpsert(app, record)
	form.LoadData(map[string]any{
		"subject":       copyEmail.Subject,
		"schedule_date": copyEmail.ScheduleDate,
	})

	filesys, _ := app.NewFilesystem()
	reader, _ := filesys.GetFile(path.Join(fromEmail.BaseFilesPath(), fromEmail.GetString("html")))
	defer reader.Close()

	data := make([]byte, reader.Size())
	reader.Read(data)

	newHtmlFile, _ := filesystem.NewFileFromBytes(data, copyEmail.Subject+".html")
	form.AddFiles("html", newHtmlFile)

	if err := form.Submit(); err != nil {
		return err
	}

	emailID := record.Id

	fmt.Printf("El correo %s se enviara en %v\n", offsetDate, emailID)

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
