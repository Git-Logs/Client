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
	"webserver/events"

	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	jsoniter "github.com/json-iterator/go"
)

var (
	json    = jsoniter.ConfigCompatibleWithStandardLibrary
	discord *discordgo.Session
	pool    *pgxpool.Pool
	ctx     = context.Background()
)

type RepoWrapper struct {
	Repo   events.Repository `json:"repository"`
	Action string            `json:"action"`
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

		repos, err := pool.Query(ctx, "SELECT id, repo_name, events, channel_id, created_at FROM repos WHERE webhook_id = $1", id)

		if err == nil {
			respStr.WriteString("This webhook is for the following repos:\n\n")

			for repos.Next() {
				var repoID string
				var repoName string
				var events []string
				var channelID string
				var createdAt time.Time

				err = repos.Scan(&repoID, &repoName, &events, &channelID, &createdAt)

				if err != nil {
					respStr.WriteString("Error: " + err.Error() + " in fetching a repo \n")
					continue
				}

				respStr.WriteString("Repo: " + repoName + "\n")
				respStr.WriteString("Repo ID: " + repoID + "\n")
				if len(events) > 0 {
					respStr.WriteString("Allowed Events: " + strings.Join(events, ", ") + "\n")
				} else {
					respStr.WriteString("This repository does not have a repo whitelist. All events will be responded to!\n")
				}
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
	var repoName string
	var allowedEvents []string

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

	// Get repo_name from database
	err = pool.QueryRow(ctx, "SELECT repo_name, events FROM repos WHERE repo_name = $1 AND webhook_id = $2", strings.ToLower(rw.Repo.FullName), id).Scan(&repoName, &allowedEvents)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusPartialContent)
		w.Write([]byte("This request has an invalid repo_url parameter"))
		return
	}

	var messageSend discordgo.MessageSend

	fmt.Println(header)

	if len(allowedEvents) > 0 {
		var gotAllowedEvent bool

		// Check if we have this event in the allowed events
		for _, evt := range allowedEvents {
			evtSplit := strings.SplitN(evt, ".", 2)

			if header == evtSplit[0] {
				if len(evtSplit) > 1 {
					if rw.Action == evtSplit[1] {
						gotAllowedEvent = true
						break
					}
				}

				gotAllowedEvent = true
				break
			}
		}

		if !gotAllowedEvent {
			w.WriteHeader(206)
			w.Write([]byte("This event is not allowed for this repo"))
			return
		}
	}

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

	// Get channel ID from database
	rows, err := pool.Query(ctx, "SELECT channel_id FROM repos WHERE repo_name = $1 AND webhook_id = $2", strings.ToLower(rw.Repo.FullName), id)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(404)
		w.Write([]byte("This request has an invalid repo_url parameter"))
		return
	}

	defer rows.Close()

	var errors string

	for rows.Next() {
		var channelId string

		err = rows.Scan(&channelId)

		if err != nil {
			fmt.Println(err)
			continue
		}

		_, err = discord.ChannelMessageSendComplex(channelId, &messageSend)

		if err != nil {
			errors += err.Error()
		}
	}

	w.Write([]byte("OK: " + repoName + "\n" + errors))
}

func main() {
	godotenv.Load()

	var err error
	pool, err = pgxpool.New(ctx, os.Getenv("DATABASE_URL"))

	if err != nil {
		panic(err)
	}

	discord, err = discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))

	discord.Identify.Intents = discordgo.IntentsGuilds

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

	http.ListenAndServe(":"+os.Getenv("PORT"), r)
}
