package globals

// SERVICE DESCRIPTION
const ServiceFriendlyName = "Android store"
const ServiceName = "android-store"
const Version = "v.1.0.0"

// ENVIRONMENT VARIABLES CONFIGURATION STRUCTURE
type ConfigStruct struct {
	HttpPort uint   `env:"AS_HTTP_PORT" envDefault:"80"`
	Url      string `env:"AS_HTTP_PORT" envDefault:"http://localhost:80"`
	BotToken string `env:"TELEGRAM_BOT_TOKEN"`
	ChatID   int    `env:"TELEGRAM_CHAT_ID"`
}

// GLOBAL VARIABLES
var (
	Config ConfigStruct
)
