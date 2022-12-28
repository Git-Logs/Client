package events

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type PushEvent struct {
	Commits []struct { // push
		ID        string `json:"id"`
		Message   string `json:"message"`
		Timestamp string `json:"timestamp"`
		URL       string `json:"url"`
		Author    struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"author"`
	} `json:"commits"`
	Repo   Repository `json:"repository"`
	Sender User       `json:"sender"`
	Pusher struct {   // push
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"pusher,omitempty"`
	Ref     string `json:"ref"`
	BaseRef string `json:"base_ref"`
}

func PushFn(bytes []byte) (discordgo.MessageSend, error) {
	var gh PushEvent

	// Unmarshal the JSON into our struct
	err := json.Unmarshal(bytes, &gh)

	if err != nil {
		return discordgo.MessageSend{}, err
	}

	var commitList string
	for _, commit := range gh.Commits {
		fmt.Println(commit.Author)

		// If the username is empty, use the name instead
		if commit.Author.Username == "" {
			commit.Author.Username = commit.Author.Name
		}

		commitList += fmt.Sprintf("%s [``%s``](%s) | [%s](%s)\n", commit.Message, commit.ID[:7], commit.URL, commit.Author.Username, strings.ReplaceAll("https://github.com/"+commit.Author.Username, " ", "%20"))
	}

	if commitList == "" {
		commitList = "No commits?"
	}

	return discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Color: 0x00ff1a,
				URL:   gh.Repo.URL,
				Author: &discordgo.MessageEmbedAuthor{
					Name:    gh.Sender.Login,
					IconURL: gh.Sender.AvatarURL,
				},
				Title: "Push on " + gh.Repo.FullName,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Branch",
						Value: "**Ref:** " + gh.Ref + "\n" + "**Base Ref:** " + gh.BaseRef,
					},
					{
						Name:  "Commits",
						Value: commitList,
					},
					{
						Name:  "Pusher",
						Value: fmt.Sprintf("[%s](%s)", gh.Pusher.Name, "https://github.com/"+gh.Pusher.Name),
					},
				},
			},
		},
	}, nil
}
