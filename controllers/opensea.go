package controllers

import (
	"io"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
)

func createReq(path string) (*string, error) {
	req, _ := http.NewRequest("GET", "https://api.opensea.io/api/v2/"+path, nil)
	req.Header.Set("accept", "application/json")
	req.Header.Set("x-api-key", "b6af094bd8c04ec6bea247173a929472")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	str := string(bytes)
	return &str, nil
}

func GetCollectionEvents(c echo.Context, app *pocketbase.PocketBase) error {
	url := "events/collection/boredapeyachtclub"
	if len(c.QueryString()) != 0 {
		url += "?" + c.QueryString()
	}

	if jsonStr, err := createReq(url); err != nil {
		return err
	} else {
		return c.String(200, string(*jsonStr))
	}
}

func GetCollectionStats(c echo.Context, app *pocketbase.PocketBase) error {
	if jsonStr, err := createReq("collections/boredapeyachtclub/stats"); err != nil {
		return err
	} else {
		return c.String(200, string(*jsonStr))
	}
}
