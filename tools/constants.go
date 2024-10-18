package tools

const (
	SUBSCRIBER_UNVERIFIED_STATUS  = "NO VERIFICADO"
	SUBSCRIBER_VERIFIED_STATUS    = "SUBSCRITO"
	SUBSCRIBER_UNSUBSCRIBE_STATUS = "DESUSCRITO"
)

const (
	EMAIL_TO_SEND_STATUS = "NO ENVIADO"
	EMAIL_SENDED_STATUS  = "ENVIADO"
	EMAIL_ERROR_STATUS   = "ERROR"
)

const (
	CONFIRMATION_TEMPLATE = "pagina_de_comfirmacion"
	UNSUBSCRIBE_TEMPLATE  = "pagina_de_desuscribirse"
	ERROR_TEMPLATE        = "pagina_de_error"
)

const (
	JWT_HOURS          = 3600
	EMAILS_PER_SECONDS = 1
)

const DEFAULT_PUBLIC_DIR = "./build"