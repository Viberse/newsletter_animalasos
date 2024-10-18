package controllers

import (
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
)

func ClicksTracker(c echo.Context, app *pocketbase.PocketBase) error {
	recordId := c.PathParam("id")

	record, err := app.Dao().FindRecordById("clicks", recordId)
	if err != nil {
		return err
	}

	counter := record.Get("counter").(float64)
	record.Set("counter", counter+1)

	if err := app.Dao().SaveRecord(record); err != nil {
		return err
	}

	return c.File("static/Untitled.gif")
}
