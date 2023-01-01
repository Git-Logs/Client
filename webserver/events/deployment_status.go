package events

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

type DeploymentStatusEvent struct {
	Action           string     `json:"action"`
	Repo             Repository `json:"repository"`
	Sender           User       `json:"sender"`
	DeploymentStatus struct {
		Creator        User      `json:"creator"`
		CreatedAt      time.Time `json:"created_at"`
		Description    string    `json:"description"`
		EnvironmentURL string    `json:"environment_url"`
		LogURL         string    `json:"log_url"`
		TargetURL      string    `json:"target_url"`
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
		color = colorGreen
	} else {
		color = colorRed
	}

	if len(gh.Deployment.Description) > 0 {
		gh.DeploymentStatus.Description += "\n" + gh.Deployment.Description
	}

	if len(gh.DeploymentStatus.Description) > 996 {
		gh.DeploymentStatus.Description = gh.DeploymentStatus.Description[:996] + "..."
	}

	if gh.DeploymentStatus.EnvironmentURL == "" {
		gh.DeploymentStatus.EnvironmentURL = "No URL available"
	}

	if gh.DeploymentStatus.LogURL == "" {
		gh.DeploymentStatus.LogURL = "No URL available"
	}

	if gh.DeploymentStatus.TargetURL == "" {
		gh.DeploymentStatus.TargetURL = "No URL available"
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
						Value:  gh.Sender.Link(),
						Inline: true,
					},
					{
						Name:   "Commit",
						Value:  gh.Repo.Commit(gh.Deployment.SHA),
						Inline: true,
					},
					{
						Name:   "Task",
						Value:  gh.Deployment.Task,
						Inline: true,
					},
					{
						Name:   "Environment URL",
						Value:  gh.DeploymentStatus.EnvironmentURL,
						Inline: true,
					},
					{
						Name:   "Log URL",
						Value:  gh.DeploymentStatus.LogURL,
						Inline: true,
					},
					{
						Name:   "Target URL",
						Value:  gh.DeploymentStatus.TargetURL,
						Inline: true,
					},
				},
			},
		},
	}, nil
}
