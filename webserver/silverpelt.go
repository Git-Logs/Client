package main

import (
	"fmt"
	"net/http"
)

func stats(w http.ResponseWriter, r *http.Request) {
	// Get guild count
	guildCount := len(discord.State.Guilds)
	var userCount int
	var shardCount = discord.ShardCount

	for _, guild := range discord.State.Guilds {
		userCount += guild.MemberCount
	}

	w.Write([]byte(fmt.Sprintf("%d,%d,%d", guildCount, userCount, shardCount)))
}
