package tools

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/joho/godotenv"
	"github.com/madflojo/tasks"
	"github.com/pocketbase/pocketbase"
	"github.com/xinguang/go-recaptcha"
)

var (
	Captcha          *recaptcha.ReCAPTCHA
	Production       = false
	JwtSecret        string
	ServerAddr       string
	SenderEmail      string
	SubscriberStatus = make(map[string]string)
	EmailStatus      = make(map[string]string)
	EmailTemplates   = make(map[string]*string)
	Scheduler        *tasks.Scheduler
)

func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	if os.Getenv("MODE") != "DEV" {
		Production = true
		log.Println("Modo en produccion")
	}

	JwtSecret = os.Getenv("JWT_SECRET")
	if len(JwtSecret) == 0 {
		log.Fatalln("No se cargo el secreto del JWT")
	}

	SenderEmail = os.Getenv("SENDER_EMAIL")
	if len(SenderEmail) == 0 {
		log.Fatalln("No se cargo el correo electronico")
	}

	if Production {
		ServerAddr = "https://newsletter.animalasos.com"
	} else {
		ServerAddr = "http://localhost:8090"
	}

	fmt.Println(ServerAddr)
}

func LoadScheduler() {
	Scheduler = tasks.New()
	if Scheduler == nil {
		log.Fatalln("No pudo crearse el scheduler")
	}
}

func LoadReCaptcha() {
	captcha, err := recaptcha.New()
	if err != nil {
		log.Fatal("Error loading reCaptcha")
	}
	Captcha = captcha
}

func LoadTemplates() {
	templates := []string{
		ERROR_TEMPLATE,
		CONFIRMATION_TEMPLATE,
		UNSUBSCRIBE_TEMPLATE,
	}

	sesCli, err := CreateSesSession(context.TODO())
	if err != nil {
		log.Fatalln(err.Error())
	}

	for i := 0; i < len(templates); i++ {
		output, err := sesCli.GetTemplate(context.TODO(), &ses.GetTemplateInput{
			TemplateName: aws.String(templates[i]),
		})
		if err != nil {
			log.Fatal(err.Error())
		}

		fmt.Println(*output.Template)
		EmailTemplates[templates[i]] = output.Template.HtmlPart

		time.Sleep(1 * time.Second)
	}
}

// Agregar el "--publicDir" flag
func SetPublicDirFlag(app *pocketbase.PocketBase) {
	var publicDirFlag string

	app.RootCmd.PersistentFlags().StringVar(
		&publicDirFlag,
		"publicDir",
		DEFAULT_PUBLIC_DIR,
		"the directory to serve static files",
	)
}
