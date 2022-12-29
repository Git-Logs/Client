package events

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type CommitCommentEvent struct {
	Action  string     `json:"action"`
	Repo    Repository `json:"repository"`
	Sender  User       `json:"sender"`
	Comment struct {
		Body     string `json:"body"`
		HTMLURL  string `json:"html_url"`
		User     User   `json:"user"`
		CommitID string `json:"commit_id"`
	}
}

func commitCommentFn(bytes []byte) (discordgo.MessageSend, error) {
	var gh CommitCommentEvent

	// Unmarshal the JSON into our struct
	err := json.Unmarshal(bytes, &gh)

	if err != nil {
		return discordgo.MessageSend{}, err
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
				URL:   gh.Comment.HTMLURL,
				Author: &discordgo.MessageEmbedAuthor{
					Name:    gh.Sender.Login,
					IconURL: gh.Sender.AvatarURL,
				},
				Title:       "Comment on comment " + gh.Repo.FullName + " (" + gh.Comment.CommitID[:7] + ")",
				Description: comment,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "User",
						Value:  fmt.Sprintf("[%s](%s)", gh.Comment.User.Login, gh.Comment.User.HTMLURL),
						Inline: true,
					},
					{
						Name:   "Commit",
						Value:  fmt.Sprintf("[%s](%s)", gh.Comment.CommitID[:7], gh.Repo.HTMLURL+"/commit/"+gh.Comment.CommitID),
						Inline: true,
					},
				},
			},
		},
	}, nil
}
