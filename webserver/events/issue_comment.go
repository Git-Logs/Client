package events

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type IssueCommentEvent struct {
	Action  string     `json:"action"`
	Repo    Repository `json:"repository"`
	Sender  User       `json:"sender"`
	Issue   Issue      `json:"issue"`
	Comment struct {
		Body    string `json:"body"`
		HTMLURL string `json:"html_url"`
		User    User   `json:"user"`
	} `json:"comment"`
}

func IssueCommentFn(bytes []byte) (discordgo.MessageSend, error) {
	var gh IssueCommentEvent

	// Unmarshal the JSON into our struct
	err := json.Unmarshal(bytes, &gh)

	if err != nil {
		return discordgo.MessageSend{}, err
	}

	var body string = gh.Issue.Body
	if len(gh.Issue.Body) > 1000 {
		body = gh.Issue.Body[:1000] + "..."
	}

	if body == "" {
		body = "No description available"
	}

	var comment string = gh.Comment.Body

	if len(gh.Comment.Body) > 1000 {
		comment = gh.Comment.Body[:1000] + "..."
	}

	if comment == "" {
		comment = "No description available"
	}

	var color int
	if gh.Action == "deleted" {
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
				Title: fmt.Sprintf("Comment on %s (#%d) %s", gh.Repo.FullName, gh.Issue.Number, gh.Action),
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "User",
						Value: fmt.Sprintf("[%s](%s)", gh.Sender.Login, gh.Sender.HTMLURL),
					},
					{
						Name:  "Title",
						Value: gh.Issue.Title,
					},
					{
						Name:  "Parent Issue",
						Value: body,
					},
					{
						Name:  "Comment",
						Value: comment,
					},
				},
			},
		},
	}, nil
}
