package events

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type CheckSuiteEvent struct {
	Action     string     `json:"action"`
	Repo       Repository `json:"repository"`
	Sender     User       `json:"sender"`
	CheckSuite struct {
		ID         int    `json:"id"`
		After      string `json:"after,omitempty"`
		HeadBranch string `json:"head_branch,omitempty"`
		HeadSHA    string `json:"head_sha,omitempty"`
		Status     string `json:"status,omitempty"`
		Conclusion string `json:"conclusion,omitempty"`
		URL        string `json:"url,omitempty"`
		Before     string `json:"before,omitempty"`
		HeadCommit struct {
			ID        string `json:"id,omitempty"`
			TreeID    string `json:"tree_id,omitempty"`
			Message   string `json:"message,omitempty"`
			Timestamp string `json:"timestamp,omitempty"`
			Author    struct {
				Name     string `json:"name,omitempty"`
				Email    string `json:"email,omitempty"`
				Username string `json:"username,omitempty"`
			} `json:"author,omitempty"`
			Committer struct {
				Name     string `json:"name,omitempty"`
				Email    string `json:"email,omitempty"`
				Username string `json:"username,omitempty"`
			} `json:"committer,omitempty"`
		} `json:"head_commit,omitempty"`
	} `json:"check_suite"`
}

func CheckSuiteFn(bytes []byte) (discordgo.MessageSend, error) {
	var gh CheckSuiteEvent

	// Unmarshal the JSON into our struct
	err := json.Unmarshal(bytes, &gh)

	if err != nil {
		return discordgo.MessageSend{}, err
	}

	return discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Color: 0x00ff1a,
				URL:   gh.Repo.HTMLURL,
				Author: &discordgo.MessageEmbedAuthor{
					Name:    gh.Sender.Login,
					IconURL: gh.Sender.AvatarURL,
				},
				Title: "Check Suite " + gh.Action + " on " + gh.Repo.FullName,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "User",
						Value:  fmt.Sprintf("[%s](%s)", gh.Sender.Login, gh.Sender.HTMLURL),
						Inline: true,
					},
					{
						Name:   "Status",
						Value:  gh.CheckSuite.Status,
						Inline: true,
					},
					{
						Name:   "Conclusion",
						Value:  gh.CheckSuite.Conclusion,
						Inline: true,
					},
					{
						Name:   "URL",
						Value:  gh.CheckSuite.URL,
						Inline: true,
					},
				},
			},
		},
	}, nil
}
