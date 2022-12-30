package events

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

type CheckRunEvent struct {
	Action   string     `json:"action"`
	Repo     Repository `json:"repository"`
	Sender   User       `json:"sender"`
	CheckRun struct {
		Name       string    `json:"name"`
		HTMLURL    string    `json:"html_url"`
		StartedAt  time.Time `json:"started_at"`
		Status     string    `json:"status"`
		DetailsURL string    `json:"details_url"`
		Conclusion string    `json:"conclusion"`
		HeadSHA    string    `json:"head_sha"`
	} `json:"check_run"`
}

func checkRunFn(bytes []byte) (discordgo.MessageSend, error) {
	var gh CheckRunEvent

	// Unmarshal the JSON into our struct
	err := json.Unmarshal(bytes, &gh)

	if err != nil {
		return discordgo.MessageSend{}, err
	}

	if gh.CheckRun.Conclusion == "" {
		gh.CheckRun.Conclusion = "No conclusion yet!"
	}

	if gh.CheckRun.Status == "" {
		gh.CheckRun.Status = "No status yet!"
	}

	return discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Color:     colorGreen,
				URL:       gh.Repo.HTMLURL,
				Author:    gh.Sender.AuthorEmbed(),
				Title:     "Check Run " + gh.CheckRun.Name + " " + gh.Action + " on " + gh.Repo.FullName,
				Timestamp: gh.CheckRun.StartedAt.Format(time.RFC3339),
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "User",
						Value:  gh.Sender.Link(),
						Inline: true,
					},
					{
						Name:   "Status",
						Value:  gh.CheckRun.Status,
						Inline: true,
					},
					{
						Name:   "Name",
						Value:  gh.CheckRun.Name,
						Inline: true,
					},
					{
						Name:   "Conclusion",
						Value:  gh.CheckRun.Conclusion,
						Inline: true,
					},
					{
						Name:   "URL",
						Value:  gh.CheckRun.HTMLURL,
						Inline: true,
					},
					{
						Name:   "Details URL",
						Value:  gh.CheckRun.DetailsURL,
						Inline: true,
					},
				},
			},
		},
	}, nil
}
