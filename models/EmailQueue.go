package models

type EmailQueue struct {
	Id               string `db:"id"`
	EmailId          string `db:"email_id"`
	Status           string `db:"status"`
	SubscriberEmail  string `db:"subscriber_email"`
	Subject          string `db:"subject"`
	UnsubscribeToken string `db:"unsubscribe_token"`
}
