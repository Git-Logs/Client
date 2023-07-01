package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"webserver/eventmodifiers"
	"webserver/events"
	"webserver/state"

	"github.com/bwmarrin/discordgo"
	"github.com/dgraph-io/badger/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/infinitybotlist/eureka/crypto"
	"go.uber.org/zap"
)

type RepoWrapper struct {
	Repo   events.Repository `json:"repository"`
	Action string            `json:"action"`
}

func sendWebhToChannel(
	msg discordgo.MessageSend,
	event string,
	channelId string,
) string {
	_, err := state.Discord.ChannelMessageSendComplex(channelId, &msg)

	if err != nil {
		state.Discord.ChannelMessageSendComplex(channelId, &discordgo.MessageSend{
			Content: "Could not send event " + event + " to channel: <#" + channelId + ">:" + err.Error(),
		})

		return err.Error()
	}

	return ""
}

func webhookRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Find the webhook in the database
		var comment string

		id := r.URL.Query().Get("id")

		if id == "" {
			w.WriteHeader(400)
			w.Write([]byte("This request is missing the id parameter"))
			return
		}

		err := state.Pool.QueryRow(state.Context, "SELECT comment FROM webhooks WHERE id = $1", id).Scan(&comment)

		if err != nil {
			w.WriteHeader(404)
			w.Write([]byte("This request has an invalid id parameter"))
			return
		}

		var respStr = strings.Builder{}

		respStr.WriteString("Comment: " + comment + "\n\n")

		// Get all event modifiers on this webhook
		modifiers, err := eventmodifiers.GetEventModifiers(id, "")

		if err != nil {
			respStr.WriteString("ERROR: " + err.Error() + " in fetching event modifiers for webhook\n")
		} else {
			respStr.WriteString("EventModifiers:\n\n")

			for _, modifier := range modifiers {
				data := map[string]string{
					"ID":     modifier.ID,
					"Events": strings.Join(modifier.Events, ","),
					"RepoID": modifier.RepoID,
					"Blacklisted": func() string {
						if modifier.Blacklisted {
							return "true"
						}

						return "false"
					}(),
					"Whitelisted": func() string {
						if modifier.Whitelisted {
							return "true"
						}

						return "false"
					}(),
					"RedirectChannel": modifier.RedirectChannel,
					"Priority":        strconv.Itoa(modifier.Priority),
				}

				for k, v := range data {
					respStr.WriteString(k + ": " + v + "\n")
				}

				respStr.WriteString("\n")
			}

			respStr.WriteString("\n\n")
		}

		repos, err := state.Pool.Query(state.Context, "SELECT id, repo_name, channel_id, created_at FROM repos WHERE webhook_id = $1", id)

		if err == nil {
			respStr.WriteString("Repositories:\n\n")

			for repos.Next() {
				var repoID string
				var repoName string
				var channelID string
				var createdAt time.Time

				err = repos.Scan(&repoID, &repoName, &channelID, &createdAt)

				if err != nil {
					respStr.WriteString("Error: " + err.Error() + " in fetching a repo \n")
					continue
				}

				respStr.WriteString("Repo: " + repoName + "\n")
				respStr.WriteString("Repo ID: " + repoID + "\n")
				respStr.WriteString("Channel ID: " + channelID + "\n")
				respStr.WriteString("Created At: " + createdAt.Format(time.RFC3339) + "\n\n")
			}
		} else {
			respStr.WriteString("This webhook doesn't seem to have any added repositories yet!\n")
		}

		w.WriteHeader(200)
		w.Write([]byte(respStr.String()))
		return
	}

	var secret string

	id := r.URL.Query().Get("id")

	if id == "" {
		w.WriteHeader(400)
		w.Write([]byte("This request is missing the id parameter"))
		return
	}

	err := state.Pool.QueryRow(state.Context, "SELECT secret FROM webhooks WHERE id = $1", id).Scan(&secret)

	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("This request has an invalid id parameter"))
		return
	}

	var bodyBytes []byte

	defer r.Body.Close()
	if r.Body != nil {
		bodyBytes, _ = io.ReadAll(r.Body)
	}

	var signature = r.Header.Get("X-Hub-Signature-256")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(bodyBytes))
	expected := hex.EncodeToString(mac.Sum(nil))

	if "sha256="+expected != signature {
		w.WriteHeader(401)
		w.Write([]byte("This request has a bad signature, recheck the secret and ensure it isnt the id...."))
		return
	}

	if r.Header.Get("X-GitHub-Event") == "ping" {
		w.WriteHeader(200)
		w.Write([]byte("pong"))
		return
	}

	var rw RepoWrapper

	err = json.Unmarshal(bodyBytes, &rw)

	if err != nil {
		state.Logger.Error("JSON unmarshal error", zap.Error(err))
		w.WriteHeader(400)
		w.Write([]byte("This request is not a valid JSON:" + err.Error()))
		return
	}

	var header = r.Header.Get("X-GitHub-Event")

	// Get repo_name from database
	var repoName string
	var repoID string
	err = state.Pool.QueryRow(state.Context, "SELECT id, repo_name FROM repos WHERE repo_name = $1 AND webhook_id = $2", strings.ToLower(rw.Repo.FullName), id).Scan(&repoID, &repoName)

	if err != nil {
		state.Logger.Error("Invalid repository parameter", zap.Error(err), zap.String("repoName", rw.Repo.FullName), zap.String("webhookID", id))
		w.WriteHeader(http.StatusPartialContent)
		w.Write([]byte("This request has an invalid repo parameter"))
		return
	}

	logId := crypto.RandString(128)

	w.Write([]byte(
		"View possible logs: " + state.Config.APIUrl + "/audit/" + logId + "\n",
	))

	// Check event modifiers
	modres, err := eventmodifiers.CheckEventAllowed(id, repoID, header)

	if err != nil {
		state.Logger.Info("Error checking event modifiers", zap.Error(err), zap.String("repoName", rw.Repo.FullName), zap.String("webhookID", id), zap.String("logId", logId))
		w.WriteHeader(http.StatusPartialContent)
		w.Write([]byte("Could not check event modifiers: " + err.Error()))
		return
	}

	if modres == nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Error: modres is nil"))
		return
	}

	if modres.ACLFail != "" {
		state.Logger.Warn("ACL Fail", zap.String("repoName", rw.Repo.FullName), zap.String("webhookID", id), zap.String("event", header), zap.String("reason", modres.ACLFail), zap.String("logId", logId))
		w.WriteHeader(http.StatusPartialContent)
		w.Write([]byte(modres.ACLFail))
		return
	}

	var messageSend discordgo.MessageSend

	evtFn, ok := events.SupportedEvents[header]

	if !ok {
		messageSend = discordgo.MessageSend{
			Content: "This event is not supported yet: " + header,
		}
	} else {
		messageSend, err = evtFn(bodyBytes)
	}

	if err != nil {
		state.Logger.Error("Error processing event", zap.Error(err), zap.String("repoName", rw.Repo.FullName), zap.String("webhookID", id), zap.String("event", header), zap.String("logId", logId))
		w.WriteHeader(400)
		w.Write([]byte("Error processing event:" + err.Error()))
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Accepted webhook event: " + header))

	go func() {
		var errList []string

		// Channel override comes from the event modifier, in the case of an event modifier, we only send
		// to the channel specified in the event modifier, not to all channels set
		if modres.ChannelOverride != "" {
			err := sendWebhToChannel(
				messageSend,
				header,
				modres.ChannelOverride,
			)

			if err != "" {
				errList = append(errList, err)
			}
		} else {
			// Get channel ID from database
			rows, err := state.Pool.Query(state.Context, "SELECT channel_id FROM repos WHERE repo_name = $1 AND webhook_id = $2", strings.ToLower(rw.Repo.FullName), id)

			if err != nil {
				state.Logger.Error("Channel id fetch error", zap.Error(err), zap.String("repoName", rw.Repo.FullName), zap.String("webhookID", id), zap.String("logId", logId))
				w.WriteHeader(404)
				w.Write([]byte("This request has an invalid repo_url parameter"))
				return
			}

			defer rows.Close()

			for rows.Next() {
				var channelId string

				err = rows.Scan(&channelId)

				if err != nil {
					state.Logger.Error("Channel id scan error", zap.Error(err), zap.String("repoName", rw.Repo.FullName), zap.String("webhookID", id), zap.String("logId", logId))
					continue
				}

				err := sendWebhToChannel(
					messageSend,
					header,
					channelId,
				)

				if err != "" {
					state.Logger.Warn("Error sending webhook to channel", zap.Error(errors.New(err)), zap.String("repoName", rw.Repo.FullName), zap.String("webhookID", id), zap.String("event", header), zap.String("channelID", channelId), zap.String("logId", logId))
					errList = append(errList, err)
				}
			}
		}

		err := state.Badger.Update(func(txn *badger.Txn) error {
			err := txn.Set([]byte(logId), []byte("OK: "+repoName+"\n\nErrors:\n"+strings.Join(errList, "\n")))

			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			state.Logger.Error("Error saving log", zap.Error(err), zap.String("repoName", rw.Repo.FullName), zap.String("webhookID", id), zap.String("event", header), zap.String("logId", logId))
		}
	}()
}

func index(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`This is the API for the Git Logs service. It handles webhooks from GitHub and sends them to Discord.

You may also be looking for:

- API (possibly unstable): /api/
  - Counts: /counts/
    - <server_count>,<user_count>,<shard_count>

- Webhooks: /kittycat?id=ID
  - Get Webhook Info: GET /kittycat?id=ID
  - Handle Github Webhook: POST /kittycat?id=ID
`))
}

func main() {
	state.Setup()

	defer state.Close()

	r := chi.NewMux()

	r.Use(middleware.Logger, middleware.Recoverer, middleware.RealIP, middleware.RequestID, middleware.Timeout(60*time.Second))

	// Webhook route
	r.HandleFunc("/kittycat", webhookRoute)
	r.HandleFunc("/", index)

	// API
	r.HandleFunc("/api/counts", stats)
	r.HandleFunc("/api/events/listview", eventsListView)
	r.HandleFunc("/api/events/csview", eventsCommaSepView)

	http.ListenAndServe(state.Config.Port, r)
}
