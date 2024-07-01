// Pneuma (Xenoblade Chronicles 2), the core component that actually handles events
package pneuma

import (
	"fmt"
	"strings"

	"github.com/git-logs/client/webserver/logos/eventmodifiers"
	"github.com/git-logs/client/webserver/logos/events"
	"github.com/git-logs/client/webserver/state"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func updateLogEntries(logId, webhookId, guildId string, entries ...any) error {
	// Check for log_id in database
	var count int

	err := state.Pool.QueryRow(state.Context, "SELECT COUNT(*) FROM "+state.TableWebhookLogs+" WHERE log_id = $1 AND webhook_id = $2 AND guild_id = $3", logId, webhookId, guildId).Scan(&count)

	if err != nil {
		return err
	}

	entry := fmt.Sprintln(entries...)

	if count == 0 {
		// Insert new log_id
		_, err = state.Pool.Exec(state.Context, "INSERT INTO "+state.TableWebhookLogs+" (log_id, webhook_id, guild_id, entries) VALUES ($1, $2, $3, $4)", logId, webhookId, guildId, []string{entry})
		return err
	}

	_, err = state.Pool.Exec(state.Context, "UPDATE "+state.TableWebhookLogs+" SET entries = array_append(entries, $1) WHERE log_id = $2 AND webhook_id = $3 AND guild_id = $4", entry, logId, webhookId, guildId)
	return err
}

func HandleEvents(
	bodyBytes []byte,
	rw *events.RepoWrapper,
	repoId string,
	logId string,
	header string,
	webhookId string,
	guildId string,
) {
	// Ensure one at a time
	l := state.MapMutex.Lock(webhookId)
	defer l.Unlock()

	updateLogEntries(logId, webhookId, guildId, "Processing event: "+header, "repoName="+rw.Repo.FullName, "webhookID="+webhookId, "event="+header, "logId="+logId)

	// Check event modifiers
	modres, err := eventmodifiers.CheckEventAllowed(webhookId, repoId, header)

	if err != nil {
		updateLogEntries(logId, webhookId, guildId, "Error checking event modifiers: "+err.Error())
		state.Logger.Error("Error checking event modifiers", zap.Error(err), zap.String("repoName", rw.Repo.FullName), zap.String("webhookID", webhookId), zap.String("logId", logId))
		return
	}

	if modres == nil {
		updateLogEntries(logId, webhookId, guildId, "Internal Error: modres is nil")
		state.Logger.Error("Internal Error: modres is nil")
		return
	}

	if modres.ACLFail != "" {
		updateLogEntries(logId, webhookId, guildId, "ACL Fail: acl="+modres.ACLFail)
		state.Logger.Warn("ACL Fail", zap.String("repoName", rw.Repo.FullName), zap.String("webhookID", webhookId), zap.String("event", header), zap.String("reason", modres.ACLFail), zap.String("logId", logId))
		return
	}

	var channelIds []string

	// Channel override comes from the event modifier, in the case of an event modifier, we only send
	// to the channel specified in the event modifier, not to all channels set
	if modres.ChannelOverride != "" {
		channelIds = []string{modres.ChannelOverride}
	} else {
		// Get channel ID from database
		rows, err := state.Pool.Query(state.Context, "SELECT channel_id FROM "+state.TableRepos+" WHERE repo_name = $1 AND webhook_id = $2", strings.ToLower(rw.Repo.FullName), webhookId)

		if err != nil {
			updateLogEntries(logId, "Channel id fetch error: acl="+modres.ACLFail, "error="+err.Error())
			state.Logger.Error("Channel id fetch error", zap.Error(err), zap.String("repoName", rw.Repo.FullName), zap.String("webhookID", webhookId), zap.String("logId", logId))
			return
		}

		defer rows.Close()

		for rows.Next() {
			var channelId string

			err = rows.Scan(&channelId)

			if err != nil {
				updateLogEntries(logId, "Channel id scan error: acl="+modres.ACLFail, "error="+err.Error())
				state.Logger.Error("Channel id scan error", zap.Error(err), zap.String("repoName", rw.Repo.FullName), zap.String("webhookID", webhookId), zap.String("logId", logId))
				continue
			}

			channelIds = append(channelIds, channelId)
		}
	}

	// Early return, don't waste resources if there are no channels to send to
	if len(channelIds) == 0 {
		return
	}

	evtFn, ok := events.SupportedEvents[header]

	if !ok {
		updateLogEntries(logId, webhookId, guildId, "WARNING: This event cannot be personalized, will try propogating to configured webhooks (if supported)?")
		// Instead of just being lazy, lets actually try to solve this problem here by using discords default github handler

		/* TODO */
		return
	} else {
		// This event can be personalized
		updateLogEntries(logId, webhookId, guildId, "SUCCESS: This event can be personalized")
		messageSend, err := evtFn(bodyBytes)

		if err != nil {
			updateLogEntries(logId, webhookId, guildId, "Error processing event:", err.Error())
			state.Logger.Error("Error processing event", zap.Error(err), zap.String("repoName", rw.Repo.FullName), zap.String("webhookID", webhookId), zap.String("event", header), zap.String("logId", logId))
			return
		}

		for _, channelId := range channelIds {
			updateLogEntries(logId, webhookId, guildId, "Sending event to channel: channelId="+channelId)
			_, err := state.Discord.ChannelMessageSendComplex(channelId, &messageSend)

			if err != nil {
				state.Discord.ChannelMessageSendComplex(channelId, &discordgo.MessageSend{
					Content: "Could not send event " + header + " to channel: <#" + channelId + ">:" + err.Error(),
				})

				updateLogEntries(logId, "Could not send event "+header+" to channel: channelId="+channelId, "err="+err.Error())
			}
		}
	}
}
