package tools

import (
	"context"
	"fmt"
	"log"
	"math"
	"newsletter/models"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/madflojo/tasks"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	pbModels "github.com/pocketbase/pocketbase/models"
)

type emailData struct {
	UnsubscribeUrl string
}

func VerifyEmail(templateName *string) error {
	sesCli, err := CreateSesSession(context.TODO())
	if err != nil {
		return err
	}
	_, err = sesCli.GetTemplate(context.TODO(), &ses.GetTemplateInput{
		TemplateName: templateName,
	})

	return err
}

func QueueEmail(emailID string, app *pocketbase.PocketBase) {
	db := app.Dao().DB()
	db.NewQuery(fmt.Sprintf(`
		INSERT INTO emails_queue(
			email_id,
			status,
			subscriber_email	
		) SELECT "%s", "%s", email FROM subscribers
		WHERE status = "%s";
	`, emailID, EmailStatus[EMAIL_TO_SEND_STATUS], SubscriberStatus[SUBSCRIBER_VERIFIED_STATUS],
	)).Execute()
}

func DequeueEmail(emailID string, app *pocketbase.PocketBase, templateName *string) error {
	db := app.Dao().DB()

	var queueLen int
	db.NewQuery(fmt.Sprintf(`
		SELECT
			COUNT(*)
		FROM emails_queue INNER JOIN subscribers
		ON subscribers.email = emails_queue.subscriber_email
		WHERE emails_queue.email_id = "%s" AND emails_queue.status = "%s"
	`, emailID, EmailStatus[EMAIL_TO_SEND_STATUS])).Row(&queueLen)
	rounds := int(math.Ceil(float64(queueLen) / EMAILS_PER_SECONDS))
	log.Printf("El correo %s tendra una rondas de %d.", emailID, rounds)

	sesCli, err := CreateSesSession(context.TODO())
	if err != nil {
		return err
	}

	collection, err := app.Dao().FindCollectionByNameOrId("clicks")
	if err != nil {
		return err
	}
	record := pbModels.NewRecord(collection)
	record.Set("counter", 0)

	if err := app.Dao().SaveRecord(record); err != nil {
		return err
	}

	recordId := record.Id

	var wg sync.WaitGroup
	for i := 0; i < rounds; i++ {
		defer TimeTrack(time.Now(), fmt.Sprintf("Para el correo %s la ronda %d tardo ", emailID, i))
		wg.Add(1)

		var queue []models.EmailQueue
		log.Println(app.Dao().DB().NewQuery(fmt.Sprintf(`
			SELECT
				emails_queue.*,
				(SELECT subject FROM emails WHERE id = "%s") as subject,
				subscribers.unsubscribe_token as unsubscribe_token
			FROM emails_queue INNER JOIN subscribers
			ON subscribers.email = emails_queue.subscriber_email
			WHERE emails_queue.email_id = "%s" AND emails_queue.status = "%s" LIMIT %d
		`, emailID, emailID, EmailStatus[EMAIL_TO_SEND_STATUS], EMAILS_PER_SECONDS)).All(&queue))

		go func() {
			defer wg.Done()
			destinations := make([]types.BulkEmailDestination, len(queue))
			for i, q := range queue {
				destinations[i] = types.BulkEmailDestination{
					Destination: &types.Destination{
						ToAddresses: []string{q.SubscriberEmail},
					},
					ReplacementTemplateData: aws.String(
						fmt.Sprintf(
							`{"unsubscribeUrl": "%s/desuscribirse/%s", "clicksTracker": "%s/clicks/%s"}`, ServerAddr, q.UnsubscribeToken, ServerAddr, recordId,
						),
					),
				}
			}

			log.Println(destinations[0].Destination.ToAddresses)
			output, err := sesCli.SendBulkTemplatedEmail(context.TODO(), &ses.SendBulkTemplatedEmailInput{
				Source:               &SenderEmail,
				Template:             templateName,
				Destinations:         destinations,
				DefaultTemplateData:  aws.String(`{"unsubscribeUrl": "%s/desuscribirse/%s"}`),
				ConfigurationSetName: aws.String("tracking"),
			})
			if err != nil {
				log.Println(err)
				return
			}

			var wg2 sync.WaitGroup
			for e, o := range output.Status {
				if o.Error != nil {
				}
				wg2.Add(1)
				go func() {
					defer wg2.Done()
					if o.Error != nil {
						db.NewQuery(fmt.Sprintf(`UPDATE emails_queue SET status = "%s", id = "%s" WHERE id = "%s";`, EmailStatus[EMAIL_ERROR_STATUS], *o.MessageId, queue[e].Id)).Execute()
					} else {
						db.NewQuery(fmt.Sprintf(`UPDATE emails_queue SET status = "%s", id = "%s" WHERE id = "%s";`, EmailStatus[EMAIL_SENDED_STATUS], *o.MessageId, queue[e].Id)).Execute()
						log.Printf("Se envio exitosamenta el correo a %s .\n", queue[e].SubscriberEmail)
					}
				}()
				wg2.Wait()
			}
		}()

		time.Sleep(1 * time.Second)
	}
	wg.Wait()

	return nil
}

func ScheduleDequeueTask(app *pocketbase.PocketBase, emailID string) error {
	record, err := app.Dao().FindRecordById("emails", emailID)
	if err != nil {
		return err
	}
	templateName := record.GetString("template_name")
	return DequeueEmail(emailID, app, aws.String(templateName))
}

// Esta funcion revisa los correos que estan pendientes para agendarlos
// otra vez. Y los que no se pudieron enviar se enviaran de una vez.
func ScheduleEmailsAgain(app *pocketbase.PocketBase) {
	dao := app.Dao().WithoutHooks()

	// Enviar los correos programados (que no estan en bucles) que no se pudieron enviar
	records, err := dao.FindRecordsByExpr("emails",
		dbx.NewExp("length(schedule_date) != 0"),
		dbx.NewExp("loop = 0"),
		dbx.NewExp("emails.id NOT IN (SELECT email_id FROM emails_queue)"),
		dbx.NewExp(`schedule_date <= date("now")`),
	)
	if err == nil {
		for i := 0; i < len(records); i++ {
			email := records[i]
			QueueEmail(email.Id, app)
			ScheduleDequeueTask(app, email.Id)
		}
	}

	// Volver a agendar los correos en bucle
	records, err = dao.FindRecordsByExpr("emails",
		dbx.NewExp("loop = 1"),
		dbx.NewExp(`schedule_date <= date("now")`),
	)
	if err == nil {
		for i := 0; i < len(records); i++ {
			email := records[i]

			// Dereferenciar la instacia actual de email para que se vaya de memoria
			// O eso espero
			emailID := email.Id

			QueueEmail(emailID, app)
			ScheduleDequeueTask(app, emailID)

			// Enviar los correos que tienen intervalos de dias
			var intervals time.Duration
			if daysIntervals := email.GetInt("intervals"); daysIntervals != 0 {
				intervals = time.Duration(daysIntervals) * (time.Hour * 24)
				scheduleID, err := Scheduler.Add(&tasks.Task{
					Interval: intervals,
					TaskFunc: func() error {
						QueueEmail(emailID, app)
						record, _ := app.Dao().FindRecordById("emails", emailID)
						if err := ScheduleDequeueTask(app, emailID); err != nil {
							return err
						}

						record.Set("schedule_date", time.Now().Add(intervals))
						return dao.SaveRecord(record)
					},
				})
				if err == nil {
					email.Set("schedule_id", scheduleID)
					email.Set("schedule_date", time.Now().Add(intervals))
					dao.SaveRecord(email)
				}
				// Enviar los correos en intervalos de dias de la semana
			} else {
				date, _ := ParseStrDateFromDB(email.GetString("schedule_date"))
				intervals = time.Duration(7) * (time.Hour * 24)
				now := time.Now()
				offsetday := (date.Weekday()-1-now.Weekday()+7)%7 + 1
				startDay := now.Add(time.Hour * time.Duration(24*offsetday))
				scheduleID, err := Scheduler.Add(&tasks.Task{
					RunOnce:  true,
					Interval: -time.Now().Sub(startDay),
					TaskFunc: func() error {
						QueueEmail(emailID, app)
						record, _ := app.Dao().FindRecordById("emails", emailID)
						if err := ScheduleDequeueTask(app, emailID); err != nil {
							return err
						}
						scheduleID, _ := Scheduler.Add(&tasks.Task{
							Interval: time.Duration(7) * (time.Hour * 24),
							TaskFunc: func() error {
								QueueEmail(emailID, app)
								record, _ := app.Dao().FindRecordById("emails", emailID)
								if err := ScheduleDequeueTask(app, emailID); err != nil {
									return err
								}
								record.Set("schedule_date", time.Now().Add(time.Duration(7)*(time.Hour*24)))
								return dao.SaveRecord(record)
							},
						})
						record.Set("schedule_date", time.Now().Add(time.Duration(7)*(time.Hour*24)))
						record.Set("schedule_id", scheduleID)
						return dao.SaveRecord(record)
					},
				})
				if err == nil {
					email.Set("schedule_id", scheduleID)
					email.Set("schedule_date", startDay)
					dao.SaveRecord(email)
				}
			}
		}
	}
}
