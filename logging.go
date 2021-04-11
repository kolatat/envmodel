package envmodel

import (
	"strings"

	"github.com/rs/zerolog"
)

var EnvMask = []string{
	"password",
	"passcode",
	"secret",
	"passphrase",
	"privkey",
	"privatekey",
	"private_key",
	"mongo_uri",
	"access_token",
	"api_key",
}

func logEnv(key, value string, event *zerolog.Event) *zerolog.Event {
	event.Str("envKey", key)
	for _, mask := range EnvMask {
		if strings.Contains(strings.ToLower(key), mask) {
			event.Bool("envMasked", true)
			return event
		}
	}
	event.Str("envValue", value)
	return event
}
