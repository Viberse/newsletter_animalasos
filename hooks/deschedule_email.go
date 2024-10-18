package hooks

import (
	"newsletter/tools"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func DeScheduleEmail(app *pocketbase.PocketBase, e *core.RecordDeleteEvent) error {
	scheduleID := e.Record.GetString("schedule_id")
	if len(scheduleID) != 0 {
		tools.Scheduler.Del(scheduleID)
	}
	return nil
}
