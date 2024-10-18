package controllers

import (
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
)

func DeleteMetadatas(c echo.Context, app *pocketbase.PocketBase) error {
	records, err := app.Dao().FindRecordsByFilter("nfts", "id != 'a'", "", 100_000)
	if err != nil {
		return err
	}

	for _, r := range records {
		go func(r *models.Record) { app.Dao().DeleteRecord(r) }(r)
	}

	return nil
}
