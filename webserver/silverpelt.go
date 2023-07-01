package main

import (
	"fmt"
	"net/http"
	"strings"
	"webserver/events"
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

func stats(w http.ResponseWriter, r *http.Request) {
	// Get guild count
	guildCount := len(state.Discord.State.Guilds)
	var userCount int
	var shardCount = state.Discord.ShardCount

	for _, guild := range state.Discord.State.Guilds {
		userCount += guild.MemberCount
	}

	w.Write([]byte(fmt.Sprintf("%d,%d,%d", guildCount, userCount, shardCount)))
}

func eventsListView(w http.ResponseWriter, r *http.Request) {
	events := []string{}

	for _, event := range eventList {
		events = append(events, "- "+event)
	}

	w.Write([]byte(strings.Join(events, "\n")))
}

func eventsCommaSepView(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(strings.Join(eventList, ",")))
}
