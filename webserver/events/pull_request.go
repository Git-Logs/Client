package events

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type PullRequestEvent struct {
	Action      string      `json:"action"`
	Repo        Repository  `json:"repository"`
	Sender      User        `json:"sender"`
	PullRequest PullRequest `json:"pull_request"`
}

func pullRequestFn(bytes []byte) (discordgo.MessageSend, error) {
	var gh PullRequestEvent

	// Unmarshal the JSON into our struct
	err := json.Unmarshal(bytes, &gh)

	if err != nil {
		return discordgo.MessageSend{}, err
	}

	var body string = gh.PullRequest.Body
	if len(gh.PullRequest.Body) > 1000 {
		body = gh.PullRequest.Body[:1000]
	}

	if body == "" {
		body = "No description available"
	}

	var color int
	if gh.Action == "closed" {
		color = 0xff0000
	} else {
		color = 0x00ff1a
	}

	return discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Color:  color,
				URL:    gh.PullRequest.HTMLURL,
				Author: gh.Sender.AuthorEmbed(),
				Title:  fmt.Sprintf("Pull Request %s on %s (#%d)", gh.Action, gh.Repo.FullName, gh.PullRequest.Number),
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Action",
						Value: gh.Action,
					},
					{
						Name:  "User",
						Value: fmt.Sprintf("[%s](%s)", gh.Sender.Login, gh.Sender.HTMLURL),
					},
					{
						Name:  "Title",
						Value: gh.PullRequest.Title,
					},
					{
						Name:  "Body",
						Value: body,
					},
					{
						Name:  "More Information",
						Value: fmt.Sprintf("**Base Ref:** %s\n**Base Label:** %s\n**Head Ref:** %s\n**Head Label:** %s", gh.PullRequest.Base.Ref, gh.PullRequest.Base.Label, gh.PullRequest.Head.Ref, gh.PullRequest.Head.Label),
					},
				},
			},
		},
	}, nil
}
