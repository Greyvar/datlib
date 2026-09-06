package lint

import (
	log "github.com/sirupsen/logrus"
)

// withQuietLogs suppresses loader info chatter so callers own reporting.
func withQuietLogs(fn func()) {
	prev := log.GetLevel()
	log.SetLevel(log.ErrorLevel)
	defer log.SetLevel(prev)
	fn()
}
