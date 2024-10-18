package controllers

import (
	"fmt"
	"net/http"
	"newsletter/tools"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
)

func Unsubscribe(c echo.Context, app *pocketbase.PocketBase) error {
	tokenStr := c.PathParam("token")
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(tools.JwtSecret), nil
	})
	if err != nil {
		return c.HTML(http.StatusBadRequest, *tools.EmailTemplates[tools.ERROR_TEMPLATE])
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok && !token.Valid {
		return c.HTML(http.StatusBadRequest, *tools.EmailTemplates[tools.ERROR_TEMPLATE])
	}

	subscriber, _ := app.Dao().FindRecordById("subscribers", claims["id"].(string))

	// Verificar que su estatus sea verificado
	if subscriber.GetString("status") != tools.SubscriberStatus[tools.SUBSCRIBER_VERIFIED_STATUS] {
		return c.HTML(http.StatusBadRequest, *tools.EmailTemplates[tools.ERROR_TEMPLATE])
	}

	subscriber.Set("status", tools.SubscriberStatus[tools.SUBSCRIBER_UNSUBSCRIBE_STATUS]) // Desuscrito

	if err := app.Dao().SaveRecord(subscriber); err != nil {
		return c.HTML(http.StatusInternalServerError, *tools.EmailTemplates[tools.ERROR_TEMPLATE])
	}

	return c.HTML(http.StatusOK, *tools.EmailTemplates[tools.UNSUBSCRIBE_TEMPLATE])
}
