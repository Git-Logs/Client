// Ontos (Xenoblade Chronicles 2), the core component that recieves requests passing it down to
// Pneuma/Logos
package ontos

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"webserver/logos/eventmodifiers"
	"webserver/logos/events"
	"webserver/pneuma"
	"webserver/state"

	"github.com/infinitybotlist/eureka/crypto"
	"go.uber.org/zap"
)

func GetWebhookRoute(w http.ResponseWriter, r *http.Request) {
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
}

func HandleWebhookRoute(w http.ResponseWriter, r *http.Request) {
	logId := crypto.RandString(128)

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

	var rw events.RepoWrapper

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
		state.Logger.Warn("This repository is not configured on git-logs, ignoring", zap.Error(err), zap.String("repoName", rw.Repo.FullName), zap.String("webhookID", id))
		w.WriteHeader(http.StatusPartialContent)
		w.Write([]byte("This repository is not configured on git-logs, ignoring"))
		return
	}

	w.Write([]byte(
		"View logs at: " + state.Config.APIUrl + "/audit?log_id=" + logId + "\n",
	))

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Going to process webhook event now: " + header))

	go pneuma.HandleEvents(
		bodyBytes,
		&rw,
		repoID,
		logId,
		header,
		id,
	)

}

func IndexPage(w http.ResponseWriter, r *http.Request) {
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

func AuditEvent(w http.ResponseWriter, r *http.Request) {
	logId := r.URL.Query().Get("log_id")

	if logId == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing log_id parameter"))
		return
	}

	var log []string

	err := state.Pool.QueryRow(state.Context, "SELECT entries FROM webhook_logs WHERE id = $1", logId).Scan(&log)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error getting log: " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strings.Join(log, "\n")))
}
