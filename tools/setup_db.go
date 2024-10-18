package tools

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
)

func CreateSubscriberStatus(app *pocketbase.PocketBase) error {
	collection, err := app.Dao().FindCollectionByNameOrId("subscriber_status")
	if err != nil {
		return err
	}

	var statusDB []struct{ Id, Status string }
	q := app.DB().Select("id", "status").From(collection.Name)
	q.All(&statusDB)
	if len(statusDB) != 0 {
		for _, status := range statusDB {
			SubscriberStatus[status.Status] = status.Id
		}
		return nil
	}

	for _, status_text := range []string{SUBSCRIBER_UNVERIFIED_STATUS, SUBSCRIBER_VERIFIED_STATUS, SUBSCRIBER_UNSUBSCRIBE_STATUS} {
		record := models.NewRecord(collection)
		record.Set("status", status_text)
		if err := app.Dao().SaveRecord(record); err != nil {
			return err
		}

		SubscriberStatus[status_text] = record.Id
	}
	return nil
}

func CreateEmailStatus(app *pocketbase.PocketBase) error {
	collection, err := app.Dao().FindCollectionByNameOrId("email_status")
	if err != nil {
		return err
	}

	var statusDB []struct{ Id, Status string }
	q := app.DB().Select("id", "status").From(collection.Name)
	q.All(&statusDB)
	if len(statusDB) != 0 {
		for _, status := range statusDB {
			EmailStatus[status.Status] = status.Id
		}
		return nil
	}

	for _, status_text := range []string{EMAIL_TO_SEND_STATUS, EMAIL_SENDED_STATUS, EMAIL_ERROR_STATUS} {
		record := models.NewRecord(collection)
		record.Set("status", status_text)
		if err := app.Dao().SaveRecord(record); err != nil {
			return err
		}
		EmailStatus[status_text] = record.Id
	}
	return nil
}
