package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"webserver/config"
	"webserver/eventmodifiers"
	"webserver/events"

	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/infinitybotlist/genconfig"
	"github.com/jackc/pgx/v5/pgxpool"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/yaml.v3"
)

var (
	json    = jsoniter.ConfigCompatibleWithStandardLibrary
	discord *discordgo.Session
	pool    *pgxpool.Pool
	ctx     = context.Background()
	v       = validator.New()
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
	_, err := discord.ChannelMessageSendComplex(channelId, &msg)

	if err != nil {
		discord.ChannelMessageSendComplex(channelId, &discordgo.MessageSend{
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

		err := pool.QueryRow(ctx, "SELECT comment FROM webhooks WHERE id = $1", id).Scan(&comment)

		if err != nil {
			w.WriteHeader(404)
			w.Write([]byte("This request has an invalid id parameter"))
			return
		}

		var respStr = strings.Builder{}

		respStr.WriteString("Comment: " + comment + "\n\n")

		// Get all event modifiers on this webhook
		modifiers, err := eventmodifiers.GetEventModifiers(ctx, pool, id, "")

		if err != nil {
			respStr.WriteString("Error: " + err.Error() + " in fetching event modifiers for webhook\n")
		} else {
			respStr.WriteString("Event modifiers:\n\n")

			for _, modifier := range modifiers {
				respStr.WriteString(fmt.Sprintf("ID: %s\n", modifier.ID))
				respStr.WriteString(fmt.Sprintf("Events: %s\n", modifier.Events))
				respStr.WriteString(fmt.Sprintf("RepoID: %s\n", modifier.RepoID))
				respStr.WriteString(fmt.Sprintf("Blacklisted: %t\n", modifier.Blacklisted))
				respStr.WriteString(fmt.Sprintf("Whitelisted: %t\n", modifier.Whitelisted))
				respStr.WriteString(fmt.Sprintf("Redirect Channel: %s\n", modifier.RedirectChannel))
			}

			respStr.WriteString("\n\n")
		}

		repos, err := pool.Query(ctx, "SELECT id, repo_name, channel_id, created_at FROM repos WHERE webhook_id = $1", id)

		if err == nil {
			respStr.WriteString("This webhook is for the following repos:\n\n")

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

	err := pool.QueryRow(ctx, "SELECT secret FROM webhooks WHERE id = $1", id).Scan(&secret)

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
		fmt.Println(err)
		w.WriteHeader(400)
		w.Write([]byte("This request is not a valid JSON:" + err.Error()))
		return
	}

	var header = r.Header.Get("X-GitHub-Event")
	fmt.Println(header)

	// Get repo_name from database
	var repoName string
	var repoID string
	err = pool.QueryRow(ctx, "SELECT id, repo_name FROM repos WHERE repo_name = $1 AND webhook_id = $2", strings.ToLower(rw.Repo.FullName), id).Scan(&repoID, &repoName)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusPartialContent)
		w.Write([]byte("This request has an invalid repo parameter"))
		return
	}

	// Check event modifiers
	modres, err := eventmodifiers.CheckEventAllowed(ctx, pool, id, repoID, header)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusPartialContent)
		w.Write([]byte("Could not check event modifiers: " + err.Error()))
	}

	if modres == nil {
		panic("Event modifier is nil")
	}

	if modres.ACLFail != "" {
		fmt.Println("ACL Fail: " + modres.ACLFail)
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
		fmt.Println(err)
		w.WriteHeader(400)
		w.Write([]byte("This request is not a valid JSON:" + err.Error()))
		return
	}

	var errors string

	// Channel override comes from the event modifier, in the case of an event modifier, we only send
	// to the channel specified in the event modifier, not to all channels set
	if modres.ChannelOverride != "" {
		err := sendWebhToChannel(
			messageSend,
			header,
			modres.ChannelOverride,
		)

		if err != "" {
			errors += "-" + err
		}
	} else {
		// Get channel ID from database
		rows, err := pool.Query(ctx, "SELECT channel_id FROM repos WHERE repo_name = $1 AND webhook_id = $2", strings.ToLower(rw.Repo.FullName), id)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(404)
			w.Write([]byte("This request has an invalid repo_url parameter"))
			return
		}

		defer rows.Close()

		for rows.Next() {
			var channelId string

			err = rows.Scan(&channelId)

			if err != nil {
				fmt.Println(err)
				continue
			}

			err := sendWebhToChannel(
				messageSend,
				header,
				channelId,
			)

			if err != "" {
				errors += "-" + err + "\n"
			}
		}
	}

	w.Write([]byte("OK: " + repoName + "\n\nErrors:\n" + errors))
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
	genconfig.SampleFileName = "api-config.yaml.sample"

	genconfig.GenConfig(config.Config{})

	cfg, err := os.ReadFile("api-config.yaml")

	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(cfg, &config.Global)

	if err != nil {
		panic(err)
	}

	err = v.Struct(config.Global)

	if err != nil {
		panic("configError: " + err.Error())
	}

	pool, err = pgxpool.New(ctx, config.Global.PostgresURL)

	if err != nil {
		panic(err)
	}

	discord, err = discordgo.New("Bot " + config.Global.Token)

	discord.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMembers

	if err != nil {
		panic(err)
	}

	err = discord.Open()

	if err != nil {
		panic(err)
	}

	r := chi.NewMux()

	r.Use(middleware.Logger, middleware.Recoverer, middleware.RealIP, middleware.RequestID, middleware.Timeout(60*time.Second))

	// Webhook route
	r.HandleFunc("/kittycat", webhookRoute)
	r.HandleFunc("/", index)

	// API
	r.HandleFunc("/api/counts", stats)
	r.HandleFunc("/api/events/listview", eventsListView)
	r.HandleFunc("/api/events/csview", eventsCommaSepView)

	http.ListenAndServe(config.Global.Port, r)
}
