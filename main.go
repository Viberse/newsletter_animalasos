package main

import (
	"log"
	"newsletter/hooks"
	"newsletter/tools"

	"github.com/pocketbase/pocketbase"
)

func main() {
	tools.LoadEnv()
	tools.LoadReCaptcha()
	tools.LoadScheduler()
	tools.LoadTemplates()

	app := pocketbase.New()

	tools.SetPublicDirFlag(app)
	hooks.SetupHooks(app)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
