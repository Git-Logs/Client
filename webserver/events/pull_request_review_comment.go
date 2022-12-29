package events

import (
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

type PullRequestReviewCommentEvent struct {
	Action      string      `json:"action"`
	Repo        Repository  `json:"repository"`
	Sender      User        `json:"sender"`
	PullRequest PullRequest `json:"pull_request"`
	Comment     struct {
		Body    string `json:"body"`
		HTMLURL string `json:"html_url"`
		User    User   `json:"user"`
	} `json:"comment"`
}

func pullRequestReviewCommentFn(bytes []byte) (discordgo.MessageSend, error) {
	var gh PullRequestReviewCommentEvent

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
				Color:       color,
				URL:         gh.PullRequest.HTMLURL,
				Author:      gh.Sender.AuthorEmbed(),
				Description: comment,
				Title:       "Pull Request Review Comment on " + gh.Repo.FullName + " (#" + strconv.Itoa(gh.PullRequest.Number) + ")",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "User",
						Value: fmt.Sprintf("[%s](%s)", gh.Comment.User.Login, gh.Comment.User.HTMLURL),
					},
					{
						Name:  "Title",
						Value: gh.PullRequest.Title,
					},
					{
						Name:  "Parent Issue",
						Value: body,
					},
				},
			},
		},
	}, nil
}
