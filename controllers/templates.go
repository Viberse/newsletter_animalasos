package controllers

import (
	"fmt"
	"net/http"
	"newsletter/tools"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
)

type Template struct {
	Html    string `json:"html" validate:"required"`
	Subject string `json:"subject" validate:"required"`
	Text    string `json:"text" validate:"required"`
	Name    string `json:"name" validate:"required"`
}

func GetTemplate(c echo.Context, app *pocketbase.PocketBase) error {
	templateName := c.PathParam("name")

	sesCli, err := tools.CreateSesSession(c.Request().Context())
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	output, err := sesCli.GetTemplate(c.Request().Context(), &ses.GetTemplateInput{
		TemplateName: aws.String(templateName),
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Template{
		Html:    *output.Template.HtmlPart,
		Subject: *output.Template.SubjectPart,
		Text:    *output.Template.TextPart,
		Name:    *output.Template.TemplateName,
	})
}

func UpdateTemplate(c echo.Context, app *pocketbase.PocketBase) error {
	template := new(Template)

	if err := c.Bind(template); err != nil {
		return c.String(http.StatusBadRequest, "Invalid body")
	}
	if err := c.Validate(template); err != nil {
		return c.String(http.StatusBadRequest, "Invalid body data")
	}

	sesCli, err := tools.CreateSesSession(c.Request().Context())
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	if _, err = sesCli.UpdateTemplate(c.Request().Context(), &ses.UpdateTemplateInput{
		Template: &types.Template{
			TemplateName: aws.String(template.Name),
			HtmlPart:     aws.String(template.Html),
			TextPart:     aws.String(template.Text),
			SubjectPart:  aws.String(template.Subject),
		},
	}); err != nil {
		return err
	}

	tools.EmailTemplates[template.Name] = aws.String(template.Html)
	return nil
}

func CreateTemplate(c echo.Context, app *pocketbase.PocketBase) error {
	template := new(Template)

	if err := c.Bind(template); err != nil {
		return c.String(http.StatusBadRequest, "Invalid body")
	}
	if err := c.Validate(template); err != nil {
		return c.String(http.StatusBadRequest, "Invalid body data")
	}

	sesCli, err := tools.CreateSesSession(c.Request().Context())
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	if _, err = sesCli.CreateTemplate(c.Request().Context(), &ses.CreateTemplateInput{
		Template: &types.Template{
			TemplateName: aws.String(template.Name),
			HtmlPart:     aws.String(template.Html),
			TextPart:     aws.String(template.Text),
			SubjectPart:  aws.String(template.Subject),
		},
	}); err != nil {
		return err
	}

	return nil
}

func DeleteTemplate(c echo.Context, app *pocketbase.PocketBase) error {
	templateName := c.PathParam("name")

	if tools.EmailTemplates[templateName] != nil || templateName == "verificacion" {
		return c.String(http.StatusBadRequest, "No se puede eliminar esta plantilla")
	}

	sesCli, err := tools.CreateSesSession(c.Request().Context())
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	_, err = sesCli.DeleteTemplate(c.Request().Context(), &ses.DeleteTemplateInput{
		TemplateName: aws.String(templateName),
	})

	if err != nil {
		return err
	}

	return c.String(http.StatusOK, "")
}

type TemplatesMetadata struct {
	Name             string    `json:"name"`
	CreatedTimestamp time.Time `json:"createdTimestamp"`
}

func ListTemplates(c echo.Context, app *pocketbase.PocketBase) error {
	sesCli, err := tools.CreateSesSession(c.Request().Context())
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	templates := make([]TemplatesMetadata, 0)
	for {
		output, err := sesCli.ListTemplates(c.Request().Context(), &ses.ListTemplatesInput{})
		if err != nil {
			return err
		}

		for _, m := range output.TemplatesMetadata {
			templates = append(templates, TemplatesMetadata{
				Name:             *m.Name,
				CreatedTimestamp: *m.CreatedTimestamp,
			})
		}

		if output.NextToken == nil {
			return c.JSON(http.StatusOK, templates)
		}
	}
}
