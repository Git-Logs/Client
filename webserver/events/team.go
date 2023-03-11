package events

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type TeamEvent struct {
	Action string     `json:"action"`
	Repo   Repository `json:"repository"`
	Sender User       `json:"sender"`
	Team   struct {
		Name        string `json:"name"`
		ID          int    `json:"id"`
		Slug        string `json:"slug"`
		Deleted     bool   `json:"deleted"`
		HTMLUrl     string `json:"html_url"`
		Permission  string `json:"permission"`
		Privacy     string `json:"privacy"`
		Description string `json:"description"`
	} `json:"team"`
}

func teamFn(bytes []byte) (discordgo.MessageSend, error) {
	var gh TeamEvent
	// Unmarshal the JSON into our struct
	err := json.Unmarshal(bytes, &gh)

	if err != nil {
		return discordgo.MessageSend{}, err
	}

	var color int
	if gh.Action != "deleted" && gh.Action != "removed_from_repository" {
		color = colorGreen
	} else {
		color = colorRed
	}

	var teamNameSlugged = gh.Team.Name

	if gh.Team.Slug != "" {
		teamNameSlugged += " | " + gh.Team.Slug
	}

	var description = gh.Team.Description

	if len(description) > 1000 {
		description = description[:1000] + "..."
	}

	if description == "" {
		description = "No description provided."
	}

	var permission = gh.Team.Permission

	if len(permission) > 1000 {
		permission = permission[:1000] + "..."
	}

	if permission == "" {
		permission = "No permissions provided."
	}

	var privacy = gh.Team.Privacy

	if len(privacy) > 1000 {
		privacy = privacy[:1000] + "..."
	}

	if privacy == "" {
		privacy = "No privacy settings set."
	}

	return discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Color:  color,
				URL:    gh.Repo.HTMLURL,
				Author: gh.Sender.AuthorEmbed(),
				Title:  "Team" + gh.Team.Name + strings.Replace(gh.Action, "_", " ", -1),
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Team",
						Value: fmt.Sprintf("[%s](%s)", teamNameSlugged, gh.Team.HTMLUrl),
					},
					{
						Name:  "User",
						Value: gh.Sender.Link(),
					},
					{
						Name:  "Permission",
						Value: permission,
					},
					{
						Name:  "Privacy",
						Value: privacy,
					},
				},
				Description: description,
			},
		},
	}, nil
}
