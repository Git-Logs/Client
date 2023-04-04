package events

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

type PageBuildEvent struct {
	Build struct {
		Commit    string    `json:"commit"`
		CreatedAt time.Time `json:"created_at"`
		Duration  int       `json:"duration"`
		Error     struct {
			Message string `json:"message"`
		} `json:"error"`
		Status string `json:"status"`
	} `json:"build"`
	Repo   Repository `json:"repository"`
	Sender User       `json:"sender"`
}

func pageBuildFn(bytes []byte) (discordgo.MessageSend, error) {
	var gh PageBuildEvent

	// Unmarshal the JSON into our struct
	err := json.Unmarshal(bytes, &gh)

	if err != nil {
		return discordgo.MessageSend{}, err
	}

	return discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Color:     colorGreen,
				URL:       gh.Repo.HTMLURL,
				Author:    gh.Sender.AuthorEmbed(),
				Title:     "Page build: " + gh.Repo.FullName,
				Timestamp: gh.Build.CreatedAt.Format(time.RFC3339),
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "User",
						Value: gh.Sender.Link(),
					},
					{
						Name:  "Changes",
						Value: gh.Repo.Commit(gh.Build.Commit),
					},
					{
						Name: "Duration",
						Value: func() string {
							if gh.Build.Duration == 0 {
								return "unknown"
							}
							return fmt.Sprintf("%d seconds", gh.Build.Duration)
						}(),
					},
					{
						Name: "Errors",
						Value: func() string {
							if gh.Build.Error.Message == "" {
								return "No errors yet!"
							}
							return gh.Build.Error.Message
						}(),
						Inline: true,
					},
					{
						Name: "Status",
						Value: func() string {
							if gh.Build.Status == "" {
								return "unknown"
							}
							return gh.Build.Status
						}(),
						Inline: true,
					},
				},
			},
		},
	}, nil
}
