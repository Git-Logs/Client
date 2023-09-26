// Ontos (Xenoblade Chronicles 2), the core component that recieves requests passing it down to Pneuma/Logos
package ontos

import (
	"fmt"
	"net/http"
	"strings"
	"webserver/logos/events"
	"webserver/state"
)

// Precomputed values
var eventList []string

func init() {
	eventList = []string{}

	for event := range events.SupportedEvents {
		eventList = append(eventList, event)
	}
}

func ApiStats(w http.ResponseWriter, r *http.Request) {
	// Get guild count
	guildCount := len(state.Discord.State.Guilds)
	var userCount int
	var shardCount = state.Discord.ShardCount

	for _, guild := range state.Discord.State.Guilds {
		userCount += guild.MemberCount
	}

	w.Write([]byte(fmt.Sprintf("%d,%d,%d", guildCount, userCount, shardCount)))
}

func ApiEventsListView(w http.ResponseWriter, r *http.Request) {
	events := []string{}

	for _, event := range eventList {
		events = append(events, "- "+event)
	}

	w.Write([]byte(strings.Join(events, "\n")))
}

func ApiEventsCommaSepView(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(strings.Join(eventList, ",")))
}