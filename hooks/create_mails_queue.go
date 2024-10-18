package hooks

import (
	"newsletter/tools"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type empty struct{}

func CreateMailsQueueAndSend(app *pocketbase.PocketBase, e *core.RecordCreateEvent) error {
	emailID := e.Record.Id

	tools.QueueEmail(emailID, app)
	return tools.DequeueEmail(emailID, app, aws.String(e.Record.GetString("template_name")))
}
