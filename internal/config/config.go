package config

const (
	Port                  = "PORT"
	DBUrl                 = "DB_URL"
	NatsUrl               = "NATS_URL"
	RestTimeoutInSec      = "REST_TIMEOUT_IN_SEC"
	NatsTopStoriesSubject = "hnfetcher.topstories"
	MaxWorkers            = "MAX_WORKERS"
	MaxCapacity           = "MAX_CAPACITY"
)

var (
	Brands = []string{
		"apple", "google", "microsoft", "nvidia", "amazon", "tesla", "meta", "twitter", "netflix",
		"snapchat", "spotify", "uber", "lyft", "airbnb", "zoom", "slack", "discord", "tiktok",
		"pinterest", "reddit", "quora", "facebook", "instagram", "whatsapp", "linkedin", "youtube",
		"twitch", "github", "nintendo", "nba", "ftc", "openai", "magic", "bluesky", "python",
		"go", "linux", "zig", "joplin", "gemma", "cortex", "pope francis",
	}
)
