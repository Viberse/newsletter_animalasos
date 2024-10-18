package hooks

import (
	"newsletter/tools"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/pocketbase/pocketbase/core"
)

func VerifyEmailHook(e *core.RecordCreateEvent) error {
	return tools.VerifyEmail(aws.String(e.Record.GetString("template_name")))
}
