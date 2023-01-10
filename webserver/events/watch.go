package events

import (
	"github.com/bwmarrin/discordgo"
)

type WatchEvent struct {
	Action string     `json:"action"`
	Repo   Repository `json:"repository"`
	Sender User       `json:"sender"`
}

func watchFn(bytes []byte) (discordgo.MessageSend, error) {
	var gh WatchEvent

	// Unmarshal the JSON into our struct
	err := json.Unmarshal(bytes, &gh)

	if err != nil {
		return discordgo.MessageSend{}, err
	}

	var color int
	var title string
	if gh.Action == "created" {
		color = colorGreen
		title = "Watching: " + gh.Repo.FullName
	} else {
		color = colorRed
		title = "No longer watching: " + gh.Repo.FullName
	}
	return discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Color:  color,
				URL:    gh.Repo.URL,
				Title:  title,
				Author: gh.Sender.AuthorEmbed(),
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "User",
						Value: gh.Sender.Link(),
					},
				},
			},
		},
	}, nil
}
