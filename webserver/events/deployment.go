package events

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type DeploymentEvent struct {
	Action     string     `json:"action"`
	Repo       Repository `json:"repository"`
	Sender     User       `json:"sender"`
	Deployment struct {
		Creator               User      `json:"creator"`
		CreatedAt             time.Time `json:"created_at"`
		SHA                   string    `json:"sha"`
		Description           string    `json:"description"`
		OriginalEnvironment   string    `json:"original_environment"`
		Environment           string    `json:"environment"`
		ProductionEnvironment bool      `json:"production_environment"`
		StatusesUrl           string    `json:"statuses_url"`
	} `json:"deployment"`
}

func deploymentFn(bytes []byte) (discordgo.MessageSend, error) {
	var gh DeploymentEvent

	// Unmarshal the JSON into our struct
	err := json.Unmarshal(bytes, &gh)

	if err != nil {
		return discordgo.MessageSend{}, err
	}

	var color int
	var title string = cases.Title(language.English).String(gh.Action) + " deployment on " + gh.Repo.FullName
	if gh.Action == "created" || gh.Action == "edited" {
		color = 0x00ff1a
	} else {
		color = 0xff0000
	}

	var env = gh.Deployment.Environment

	if gh.Deployment.OriginalEnvironment != gh.Deployment.Environment && gh.Deployment.OriginalEnvironment != "" {
		env = gh.Deployment.OriginalEnvironment + " => " + gh.Deployment.Environment
	}

	if len(gh.Deployment.Description) > 996 {
		gh.Deployment.Description = gh.Deployment.Description[:996] + "..."
	}

	return discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Color:       color,
				URL:         gh.Repo.URL,
				Title:       title,
				Author:      gh.Deployment.Creator.AuthorEmbed(),
				Description: gh.Deployment.Description,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "User",
						Value:  "[" + gh.Sender.Login + "]" + "(" + gh.Sender.HTMLURL + ")",
						Inline: true,
					},
					{
						Name:   "Environment",
						Value:  env,
						Inline: true,
					},
					{
						Name:   "Commit",
						Value:  fmt.Sprintf("[%s](%s)", gh.Deployment.SHA[:7], gh.Repo.HTMLURL+"/commit/"+gh.Deployment.SHA),
						Inline: true,
					},
					{
						Name:   "Is Production",
						Value:  fmt.Sprintf("%t", gh.Deployment.ProductionEnvironment),
						Inline: true,
					},
				},
			},
		},
	}, nil
}
