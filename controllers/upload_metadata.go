package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
)

type Metadata struct {
	Image      string           `json:"image"`
	Attributes []MetaAttributes `json:"attributes"`
}

type MetaAttributes struct {
	TraitType string `json:"trait_type"`
	Value     any    `json:"value"`
}

func UploadMetadas(c echo.Context, app *pocketbase.PocketBase) error {
	if filesForm, err := c.MultipartForm(); err != nil {
		fmt.Println(err)
		return err
	} else {
		collection, err := app.Dao().FindCollectionByNameOrId("nfts")
		if err != nil {
			return err
		}

		for _, fileHeader := range filesForm.File {
			file, err := fileHeader[0].Open()
			defer file.Close()
			if err != nil {
				return err
			}

			buf := bytes.NewBuffer(nil)
			if _, err := io.Copy(buf, file); err != nil {
				return err
			}

			metadata := Metadata{}
			if err := json.Unmarshal(buf.Bytes(), &metadata); err != nil {
				fmt.Println(err)
				return err
			}

			record := models.NewRecord(collection)

			record.Set("url", metadata.Image)
			for i := 0; i < len(metadata.Attributes); i++ {
				switch metadata.Attributes[i].TraitType {
				case "FONDOS":
					record.Set("fondos", metadata.Attributes[i].Value)
				case "GESTOS":
					record.Set("gestos", metadata.Attributes[i].Value)
				case "PEINADOS Y SOMBREROS":
					record.Set("peinados_y_sombreros", metadata.Attributes[i].Value)
				case "LENTES":
					record.Set("lentes", metadata.Attributes[i].Value)
				case "CAMISAS Y CHAQUETAS":
					record.Set("camisas_y_chaquetas", metadata.Attributes[i].Value)
				case "ANIMAL":
					record.Set("animal", metadata.Attributes[i].Value)
				}
			}
			if err := app.Dao().SaveRecord(record); err != nil {
				return err
			}
		}
	}
	return nil
}
