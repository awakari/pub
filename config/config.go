package config

import (
	"github.com/kelseyhightower/envconfig"
	"time"
)

type Config struct {
	Api struct {
		Source struct {
			ActivityPub ActivityPubConfig
			Feeds       FeedsConfig
			Sites       SitesConfig
			Telegram    TelegramConfig
		}
		Writer WriterConfig
		Events EventsConfig
		TgBot  TgBotConfig
		Http   struct {
			Port uint16 `envconfig:"API_HTTP_PORT" default:"8080"`
		}
		Auth  AuthConfig
		Usage UsageConfig
	}
	Log struct {
		Level int `envconfig:"LOG_LEVEL" default:"-4" required:"true"`
	}
}

type FeedsConfig struct {
	Uri string `envconfig:"API_SOURCE_FEEDS_URI" default:"source-feeds:50051" required:"true"`
}

type TelegramConfig struct {
	Uri           string `envconfig:"API_SOURCE_TELEGRAM_URI" default:"source-telegram:50051" required:"true"`
	FmtUriReplica string `envconfig:"API_SOURCE_TELEGRAM_FMT_URI_REPLICA" default:"source-telegram-%d:50051" required:"true"`
}

type SitesConfig struct {
	Uri string `envconfig:"API_SOURCE_SITES_URI" default:"source-sites:50051" required:"true"`
}

type ActivityPubConfig struct {
	Uri string `envconfig:"API_SOURCE_ACTIVITYPUB_URI" default:"int-activitypub:50051" required:"true"`
}

type WriterConfig struct {
	Internal WriterInternalConfig
}

type EventsConfig struct {
	Uri        string `envconfig:"API_EVENTS_URI" default:"events:50051" required:"true"`
	Connection struct {
		Count struct {
			Init uint32 `envconfig:"API_EVENTS_CONN_COUNT_INIT" default:"1" required:"true"`
			Max  uint32 `envconfig:"API_EVENTS_CONN_COUNT_MAX" default:"10" required:"true"`
		}
		IdleTimeout time.Duration `envconfig:"API_EVENTS_CONN_IDLE_TIMEOUT" default:"15m" required:"true"`
	}
	Topic string `envconfig:"API_EVENTS_TOPIC" default:"published" required:"true"`
	Limit uint32 `envconfig:"API_EVENTS_LIMIT" default:"100000" required:"true"`
}

type WriterInternalConfig struct {
	Name               string `envconfig:"API_WRITER_INTERNAL_NAME" default:"awkinternal" required:"true"`
	Value              int32  `envconfig:"API_WRITER_INTERNAL_VALUE" required:"true"`
	RateLimitPerMinute int    `envconfig:"API_WRITER_INTERNAL_RATE_LIMIT_PER_MINUTE" default:"1" required:"true"`
}

type TgBotConfig struct {
	Uri string `envconfig:"API_TGBOT_URI" default:"bot-telegram:50051" required:"true"`
}

type AuthConfig struct {
	Uri string `envconfig:"API_AUTH_URI" default:"auth:50051" required:"true"`
}

type UsageConfig struct {
	Uri        string `envconfig:"API_USAGE_URI" default:"usage:50051" required:"true"`
	Connection struct {
		Count struct {
			Init uint32 `envconfig:"API_USAGE_CONN_COUNT_INIT" default:"1" required:"true"`
			Max  uint32 `envconfig:"API_USAGE_CONN_COUNT_MAX" default:"10" required:"true"`
		}
		IdleTimeout time.Duration `envconfig:"API_USAGE_CONN_IDLE_TIMEOUT" default:"15m" required:"true"`
	}
}

func NewConfigFromEnv() (cfg Config, err error) {
	err = envconfig.Process("", &cfg)
	return
}
