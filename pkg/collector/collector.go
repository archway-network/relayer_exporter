package collector

import (
	"regexp"
	"strings"

	"github.com/archway-network/relayer_exporter/pkg/config"
)

const (
	successStatus = "success"
	errorStatus   = "error"
)

func getDiscordIDs(ops []config.Operator) string {
	var ids []string

	pattern := regexp.MustCompile(`^\d+$`)

	for _, op := range ops {
		if pattern.MatchString(op.Discord.ID) {
			ids = append(ids, op.Discord.ID)
		}
	}

	return strings.Join(ids, ",")
}
