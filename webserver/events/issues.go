package events

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type IssuesEvent struct {
	Action string     `json:"action"`
	Repo   Repository `json:"repository"`
	Sender User       `json:"sender"`
	Issue  Issue      `json:"issue"`
}

func issuesFn(bytes []byte) (discordgo.MessageSend, error) {
	var gh IssuesEvent

	// Unmarshal the JSON into our struct
	err := json.Unmarshal(bytes, &gh)

	if err != nil {
		return discordgo.MessageSend{}, err
	}

	var body string = gh.Issue.Body
	if len(gh.Issue.Body) > 996 {
		body = gh.Issue.Body[:996] + "..."
	}

	if body == "" {
		body = "No description available"
	}

	var color int
	if gh.Action == "deleted" || gh.Action == "unpinned" {
		color = 0xff0000
	} else {
		color = 0x00ff1a
	}

	return discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Color: color,
				URL:   gh.Issue.HTMLURL,
				Author: &discordgo.MessageEmbedAuthor{
					Name:    gh.Sender.Login,
					IconURL: gh.Sender.AvatarURL,
				},
				Description: body,
				Title:       fmt.Sprintf("Issue %s on %s (#%d)", gh.Action, gh.Repo.FullName, gh.Issue.Number),
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Action",
						Value:  gh.Action,
						Inline: true,
					},
					{
						Name:   "User",
						Value:  fmt.Sprintf("[%s](%s)", gh.Sender.Login, gh.Sender.HTMLURL),
						Inline: true,
					},
					{
						Name:   "Title",
						Value:  gh.Issue.Title,
						Inline: true,
					},
				},
			},
		},
	}, nil
}
