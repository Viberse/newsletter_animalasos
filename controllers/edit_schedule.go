package controllers

import (
	"fmt"
	"net/http"
	"newsletter/tools"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/madflojo/tasks"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
)

type emailScheduled struct {
	ID         string    `json:"id" validate:"required"`
	Subject    string    `json:"subject" validate:"required"`
	ScheduleID string    `json:"scheduleId" validate:"required"`
	NewDate    time.Time `json:"newDate" validate:"required"`
}

func EditSchedule(c echo.Context, app *pocketbase.PocketBase) error {
	email := new(emailScheduled)
	if err := c.Bind(email); err != nil {
		return c.String(http.StatusBadRequest, "Invalid body")
	}
	if err := c.Validate(email); err != nil {
		return c.String(http.StatusBadRequest, "Invalid body data")
	}

	now := email.NewDate.Sub(time.Now())
	if now.Milliseconds() < 0 {
		return c.String(http.StatusBadRequest, "La fecha es menor a la actual")
	}

	if _, err := tools.Scheduler.Lookup(email.ScheduleID); err != nil {
		return c.String(http.StatusBadRequest, "Ya no esta programado el correo")
	}

	dao := daos.New(app.Dao().DB())
	record, err := dao.FindRecordById("emails", email.ID)
	if err != nil {
		return c.String(http.StatusBadRequest, "No se encontro el correo programado")
	}

	fmt.Printf("El correo %s se enviara en %v\n", record.Id, now)

	if scheduleID, err := tools.Scheduler.Add(&tasks.Task{
		Interval: now,
		RunOnce:  true,
		TaskFunc: func() error {
			tools.QueueEmail(record.Id, app)
			return tools.ScheduleDequeueTask(app, record.Id)
		},
	}); err != nil {
		return c.String(http.StatusInternalServerError, "Ocurrio un error reprogramando el correo.")
	} else {
		tools.Scheduler.Del(email.ScheduleID)
		record.Set("schedule_id", scheduleID)
		record.Set("subject", email.Subject)
		record.Set("schedule_date", email.NewDate)
		dao.SaveRecord(record)
	}

	return nil
}
