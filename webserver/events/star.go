package events

import (
	"github.com/bwmarrin/discordgo"
)

type StarEvent struct {
	Action string     `json:"action"`
	Repo   Repository `json:"repository"`
	Sender User       `json:"sender"`
}

func starFn(bytes []byte) (discordgo.MessageSend, error) {
	var gh StarEvent

	// Unmarshal the JSON into our struct
	err := json.Unmarshal(bytes, &gh)

	if err != nil {
		return discordgo.MessageSend{}, err
	}

	var color int
	var title string
	if gh.Action == "created" {
		color = 0x00ff1a
		title = "Starred: " + gh.Repo.FullName
	} else {
		color = 0xff0000
		title = "Unstarred: " + gh.Repo.FullName
	}
	return discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Color: color,
				URL:   gh.Repo.URL,
				Title: title,
				Author: &discordgo.MessageEmbedAuthor{
					Name:    gh.Sender.Login,
					IconURL: gh.Sender.AvatarURL,
				},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "User",
						Value: "[" + gh.Sender.Login + "]" + "(" + gh.Sender.HTMLURL + ")",
					},
				},
			},
		},
	}, nil
}
