package events

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

type DeploymentStatusEvent struct {
	Action           string     `json:"action"`
	Repo             Repository `json:"repository"`
	Sender           User       `json:"sender"`
	DeploymentStatus struct {
		Creator     User      `json:"creator"`
		CreatedAt   time.Time `json:"created_at"`
		Description string    `json:"description"`
		DeployURL   string    `json:"deploy_url"`
		LogURL      string    `json:"log_url"`
	} `json:"deployment_status"`
	Deployment struct {
		Task        string `json:"task"`
		Description string `json:"description"`
		SHA         string `json:"sha"`
	}
}

func deploymentStatusFn(bytes []byte) (discordgo.MessageSend, error) {
	var gh DeploymentStatusEvent

	// Unmarshal the JSON into our struct
	err := json.Unmarshal(bytes, &gh)

	if err != nil {
		return discordgo.MessageSend{}, err
	}

	var color int
	var title string = "Deployment status update (" + gh.Action + ") on " + gh.Repo.FullName
	if gh.Action == "created" || gh.Action == "edited" {
		color = 0x00ff1a
	} else {
		color = 0xff0000
	}

	if len(gh.Deployment.Description) > 0 {
		gh.DeploymentStatus.Description += "\n" + gh.Deployment.Description
	}

	if len(gh.DeploymentStatus.Description) > 996 {
		gh.DeploymentStatus.Description = gh.DeploymentStatus.Description[:996] + "..."
	}

	if gh.DeploymentStatus.DeployURL == "" {
		gh.DeploymentStatus.DeployURL = "No URL available"
	} else {
		gh.DeploymentStatus.DeployURL = "[Click here](" + gh.DeploymentStatus.DeployURL + ")"
	}

	if gh.DeploymentStatus.LogURL == "" {
		gh.DeploymentStatus.LogURL = "No URL available"
	} else {
		gh.DeploymentStatus.LogURL = "[Click here](" + gh.DeploymentStatus.LogURL + ")"
	}

	return discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Color:       color,
				URL:         gh.Repo.URL,
				Title:       title,
				Author:      gh.DeploymentStatus.Creator.AuthorEmbed(),
				Description: gh.DeploymentStatus.Description,
				Timestamp:   gh.DeploymentStatus.CreatedAt.Format(time.RFC3339),
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "User",
						Value:  "[" + gh.Sender.Login + "]" + "(" + gh.Sender.HTMLURL + ")",
						Inline: true,
					},
					{
						Name:   "Commit",
						Value:  fmt.Sprintf("[%s](%s)", gh.Deployment.SHA[:7], gh.Repo.HTMLURL+"/commit/"+gh.Deployment.SHA),
						Inline: true,
					},
					{
						Name:   "Task",
						Value:  gh.Deployment.Task,
						Inline: true,
					},
					{
						Name:   "Deploy URL",
						Value:  gh.DeploymentStatus.DeployURL,
						Inline: true,
					},
					{
						Name:   "Log URL",
						Value:  gh.DeploymentStatus.LogURL,
						Inline: true,
					},
				},
			},
		},
	}, nil
}
