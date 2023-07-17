package eventmodifiers

import (
	"webserver/state"

	"github.com/jackc/pgx/v5/pgtype"
)

func isNull(s pgtype.Text) bool {
	return !s.Valid || s.String == ""
}

type EventCheck struct {
	// Whether or not the ACL check passed
	ACLFail string

	// What channel to redirect to
	ChannelOverride string

	// Overridden by higher priority modifiers
	// Only applies to whitelists
	Overriden bool
}

type EventModifier struct {
	ID              string
	RepoID          string
	Events          []string
	Blacklisted     bool
	Whitelisted     bool
	RedirectChannel string
	Priority        int
}

func GetEventModifiers(
	webhookId string,
	ghRepoId string,
) ([]*EventModifier, error) {
	// Get all event_modifiers for webhook
	rows, err := state.Pool.Query(state.Context, "SELECT id, repo_id, events, blacklisted, whitelisted, redirect_channel, priority FROM event_modifiers WHERE webhook_id = $1 ORDER BY priority DESC", webhookId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var modifiers []*EventModifier

	for rows.Next() {
		var id string
		var repoId pgtype.Text
		var events []string
		var blacklisted bool
		var whitelisted bool
		var redirectChannel pgtype.Text
		var priority int

		err = rows.Scan(&id, &repoId, &events, &blacklisted, &whitelisted, &redirectChannel, &priority)

		if err != nil {
			return nil, err
		}

		// Check repo id first
		//
		// If repo_id is null, then it matches all repos
		// If repo_id is not null, then it matches only that repo
		if ghRepoId != "" && (!isNull(repoId) && repoId.String != ghRepoId) {
			// Look for another modifier, this one doesn't match
			continue
		}

		modifiers = append(modifiers, &EventModifier{
			ID:              id,
			RepoID:          repoId.String,
			Events:          events,
			Blacklisted:     blacklisted,
			Whitelisted:     whitelisted,
			RedirectChannel: redirectChannel.String,
			Priority:        priority,
		})
	}

	return modifiers, nil
}

func CheckEventAllowed(
	webhookId string,
	ghRepoId string,
	ghEvent string,
) (*EventCheck, error) {
	// Get all event_modifiers for webhook
	modifiers, err := GetEventModifiers(webhookId, ghRepoId)

	if err != nil {
		return nil, err
	}

	var resultantEventCheck *EventCheck = &EventCheck{}

	for _, modifier := range modifiers {
		// Check if the event is in the list of events
		var matched bool
		for _, event := range modifier.Events {
			if isMatch(event, ghEvent) {
				matched = true
				break
			}
		}

		if !matched {
			// Ensure that the modifier does not set whitelisted to true
			if modifier.Whitelisted {
				// Check if theres also a matching redirect channel
				if resultantEventCheck.Overriden {
					// We can short-circuit here because we have a matching redirect channel on a higher priority modifier
					return resultantEventCheck, nil
				}

				return &EventCheck{
					ACLFail: "event_modifier " + modifier.ID + ": whitelist-only event modifier but event not matched",
				}, nil
			}

			// Look for another modifier if this one doesn't match
			continue
		}

		if modifier.Blacklisted {
			return &EventCheck{
				ACLFail: "event_modifier " + modifier.ID + ": blacklisted event modifier and event matches modifier",
			}, nil
		}

		// Set the channel override if it's not null on the modifier
		if modifier.RedirectChannel != "" {
			resultantEventCheck.ChannelOverride = modifier.RedirectChannel
			resultantEventCheck.Overriden = true
		}

		// We cannot short-circuit here because we may have modifiers matching the same event
	}

	return resultantEventCheck, nil
}
