package globals

// SERVICE DESCRIPTION
const ServiceFriendlyName = "Android store"
const ServiceName = "android-store"
const Version = "1.1.3"

// ENVIRONMENT VARIABLES CONFIGURATION STRUCTURE
type ConfigStruct struct {
	HttpPort uint   `env:"AS_HTTP_PORT" envDefault:"80"`
	Url      string `env:"AS_HTTP_URL" envDefault:"http://localhost:80"`
	BotToken string `env:"AS_TELEGRAM_BOT_TOKEN" envDefault:""`
	ChatID   int    `env:"AS_TELEGRAM_CHAT_ID" envDefault:"0"`
}

// GLOBAL VARIABLES
var (
	Config ConfigStruct
)
