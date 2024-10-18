package controllers

import (
	"fmt"
	"net/http"
	"newsletter/tools"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	pb "github.com/pocketbase/pocketbase/models"
)

type subscriber struct {
	Email string `json:"email" validate:"required,email"`
}

type verifyEmailData struct {
	SubscribeUrl string
}

func Subscribe(c echo.Context, app *pocketbase.PocketBase) error {
	// Chequear el token del captcha
	if tools.Production {
		token := c.Request().Header.Get("captcha-token")
		if len(token) != 0 {
			if err := tools.Captcha.Verify(token); err != nil {
				return c.String(http.StatusUnauthorized, "Bad reputation")
			}
		}
	}

	// Chequear los datos del subscriptor
	subscriber := new(subscriber)
	if err := c.Bind(subscriber); err != nil {
		return c.String(http.StatusBadRequest, "Invalid body")
	}
	if err := c.Validate(subscriber); err != nil {
		return c.String(http.StatusBadRequest, "Invalid body data")
	}

	// Obtener la colleccion de subscriptores
	collection, err := app.Dao().FindCollectionByNameOrId("subscribers")
	if err != nil {
		return c.String(http.StatusInternalServerError, "")
	}

	// Verificar si existe y esta desuscrito
	if dbSubscriber, err := app.Dao().FindFirstRecordByData("subscribers", "email", subscriber.Email); err == nil {
		if dbSubscriber.GetString("status") == tools.SubscriberStatus[tools.SUBSCRIBER_UNSUBSCRIBE_STATUS] {
			dbSubscriber.Set("status", tools.SubscriberStatus[tools.SUBSCRIBER_VERIFIED_STATUS])
			app.Dao().SaveRecord(dbSubscriber)
			return c.String(http.StatusOK, "Subscrito otra vez")
		}
	}

	// Crear un nuevo subscriptor
	record := pb.NewRecord(collection)
	record.Set("email", subscriber.Email)

	record.Set("status", tools.SubscriberStatus[tools.SUBSCRIBER_UNVERIFIED_STATUS]) // No verificado

	// Guardarlo en la colleccion
	if err := app.Dao().SaveRecord(record); err != nil {
		return c.String(http.StatusInternalServerError, "")
	}

	// Generar el token de verificacion
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"id": record.Id,
		},
	)
	tokenString, err := token.SignedString([]byte(tools.JwtSecret))
	if err != nil {
		return c.String(http.StatusInternalServerError, "")
	}

	sesCli, err := tools.CreateSesSession(c.Request().Context())
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	fmt.Println(tools.ServerAddr)
	_, err = sesCli.SendTemplatedEmail(c.Request().Context(), &ses.SendTemplatedEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{
				subscriber.Email,
			},
		},
		Source:       aws.String(tools.SenderEmail),
		Template:     aws.String("verificacion"),
		TemplateData: aws.String(fmt.Sprintf(`{"comfirmUrl": "%s/verificar/%s"}`, tools.ServerAddr, tokenString)),
	})
	return err
}
