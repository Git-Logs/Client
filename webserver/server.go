package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	jsoniter "github.com/json-iterator/go"
	"github.com/spewerspew/spew"
)

var (
	json    = jsoniter.ConfigCompatibleWithStandardLibrary
	discord *discordgo.Session
	pool    *pgxpool.Pool
	ctx     = context.Background()
)

func webhookRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.WriteHeader(200)
		w.Write([]byte("This is the webhook route. You need to put this in Github as per the instructions in your DM"))
		return
	}

	var secret string
	var channelId string
	var repoName string

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

	// Restore the io.ReadCloser to its original state
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var signature = r.Header.Get("X-Hub-Signature-256")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(bodyBytes))
	expected := hex.EncodeToString(mac.Sum(nil))

	if "sha256="+expected != signature {
		w.WriteHeader(401)
		w.Write([]byte("This request has a bad signature, recheck the secret"))
		return
	}

	if r.Header.Get("X-GitHub-Event") == "ping" {
		w.WriteHeader(200)
		w.Write([]byte("pong"))
		return
	}

	var gh GithubWebhook

	err = json.Unmarshal(bodyBytes, &gh)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(400)
		w.Write([]byte("This request is not a valid JSON:" + err.Error()))
		return
	}

	// Get channel ID from database
	err = pool.QueryRow(ctx, "SELECT channel_id, repo_name FROM repos WHERE repo_url = $1 AND webhook_id = $2", strings.ToLower(gh.Repo.FullName), id).Scan(&channelId, &repoName)

	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("This request has an invalid repo_url parameter"))
		return
	}

	var header = r.Header.Get("X-GitHub-Event")

	var messageSend discordgo.MessageSend

	switch header {

	case "push":
		var commitList string
		for _, commit := range gh.Commits {
			commitList += fmt.Sprintf("%s [%s](%s) | [%s](%s)\n", commit.Message, commit.ID[:7], commit.URL, commit.Author.Username, "https://github.com/"+commit.Author.Username)
		}

		messageSend = discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				{
					Color: 0x00ff1a,
					URL:   gh.Repo.URL,
					Author: &discordgo.MessageEmbedAuthor{
						Name:    gh.Sender.Login,
						IconURL: gh.Sender.AvatarURL,
					},
					Title: "Push on " + gh.Repo.FullName,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Branch",
							Value: "**Ref:** " + gh.Ref + "\n" + "**Base Ref:** " + gh.BaseRef,
						},
						{
							Name:  "Commits",
							Value: commitList,
						},
						{
							Name:  "Pusher",
							Value: fmt.Sprintf("[%s](%s)", gh.Pusher.Name, "https://github.com/"+gh.Pusher.Name),
						},
					},
				},
			},
		}

	case "star":
		var color int
		var title string
		if gh.Action == "created" {
			color = 0x00ff1a
			title = "Starred: " + gh.Repo.FullName
		} else {
			color = 0xff0000
			title = "Unstarred: " + gh.Repo.FullName
		}
		messageSend = discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				{
					Color: color,
					URL:   gh.Repo.URL,
					Title: title,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "User",
							Value: "[" + gh.Sender.Login + "]" + "(" + gh.Sender.HTMLURL + ")",
						},
					},
				},
			},
		}

	case "issues":
		var body string = gh.Issue.Body
		if len(gh.Issue.Body) > 996 {
			body = gh.Issue.Body[:996] + "..."
		}

		if body == "" {
			body = "No description available"
		}

		var color int
		if gh.Action == "deleted" || gh.Action == "unpinned" {
			color = 0xff0000
		} else {
			color = 0x00ff1a
		}

		messageSend = discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				{
					Color: color,
					URL:   gh.Issue.HTMLURL,
					Author: &discordgo.MessageEmbedAuthor{
						Name:    gh.Sender.Login,
						IconURL: gh.Sender.AvatarURL,
					},
					Title: fmt.Sprintf("Issue %s on %s (#%d)", gh.Action, gh.Repo.FullName, gh.Issue.Number),
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Action",
							Value: gh.Action,
						},
						{
							Name:  "User",
							Value: fmt.Sprintf("[%s](%s)", gh.Sender.Login, gh.Sender.HTMLURL),
						},
						{
							Name:  "Title",
							Value: gh.Issue.Title,
						},
						{
							Name:  "Body",
							Value: body,
						},
					},
				},
			},
		}

	case "pull_request":
		var body string = gh.PullRequest.Body
		if len(gh.PullRequest.Body) > 1000 {
			body = gh.PullRequest.Body[:1000]
		}

		if body == "" {
			body = "No description available"
		}

		var color int
		if gh.Action == "closed" {
			color = 0xff0000
		} else {
			color = 0x00ff1a
		}

		messageSend = discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				{
					Color: color,
					URL:   gh.PullRequest.HTMLURL,
					Author: &discordgo.MessageEmbedAuthor{
						Name:    gh.Sender.Login,
						IconURL: gh.Sender.AvatarURL,
					},
					Title: fmt.Sprintf("Pull Request %s on %s (#%d)", gh.Action, gh.Repo.FullName, gh.PullRequest.Number),
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Action",
							Value: gh.Action,
						},
						{
							Name:  "User",
							Value: fmt.Sprintf("[%s](%s)", gh.Sender.Login, gh.Sender.HTMLURL),
						},
						{
							Name:  "Title",
							Value: gh.PullRequest.Title,
						},
						{
							Name:  "Body",
							Value: body,
						},
						{
							Name:  "More Information",
							Value: fmt.Sprintf("**Base Ref:** %s\n**Base Label:** %s\n**Head Ref:** %s\n**Head Label:** %s", gh.PullRequest.Base.Ref, gh.PullRequest.Base.Label, gh.PullRequest.Head.Ref, gh.PullRequest.Head.Label),
						},
					},
				},
			},
		}

	case "issue_comment":
		var body string = gh.Issue.Body
		if len(gh.Issue.Body) > 1000 {
			body = gh.Issue.Body[:1000]
		}

		if body == "" {
			body = "No description available"
		}

		var color int
		if gh.Action == "deleted" {
			color = 0xff0000
		} else {
			color = 0x00ff1a
		}

		messageSend = discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				{
					Color: color,
					URL:   gh.Issue.HTMLURL,
					Author: &discordgo.MessageEmbedAuthor{
						Name:    gh.Sender.Login,
						IconURL: gh.Sender.AvatarURL,
					},
					Title: fmt.Sprintf("Comment on %s (#%d) %s", gh.Repo.FullName, gh.Issue.Number, gh.Action),
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "User",
							Value: fmt.Sprintf("[%s](%s)", gh.Sender.Login, gh.Sender.HTMLURL),
						},
						{
							Name:  "Title",
							Value: gh.Issue.Title,
						},
						{
							Name:  "Body",
							Value: body,
						},
					},
				},
			},
		}

	case "pull_request_review_comment":
		var body string = gh.PullRequest.Body
		if len(gh.PullRequest.Body) > 1000 {
			body = gh.PullRequest.Body[:1000]
		}

		if body == "" {
			body = "No description available"
		}

		var color int
		if gh.Action == "deleted" {
			color = 0xff0000
		} else {
			color = 0x00ff1a
		}

		messageSend = discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				{
					Color: color,
					URL:   gh.PullRequest.HTMLURL,
					Author: &discordgo.MessageEmbedAuthor{
						Name:    gh.Sender.Login,
						IconURL: gh.Sender.AvatarURL,
					},
					Title: "Pull Request Review Comment on " + gh.Repo.FullName + " (#" + strconv.Itoa(gh.PullRequest.Number) + ")",
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "User",
							Value: fmt.Sprintf("[%s](%s)", gh.Sender.Login, gh.Sender.HTMLURL),
						},
						{
							Name:  "Title",
							Value: gh.PullRequest.Title,
						},
						{
							Name:  "Body",
							Value: body,
						},
					},
				},
			},
		}
	default:
		messageSend = discordgo.MessageSend{
			Content: "**Action: " + header + "**",
			TTS:     false,
			File: &discordgo.File{
				Name:        "gh-event.txt",
				ContentType: "application/octet-stream",
				Reader:      strings.NewReader(spew.Sdump(gh)),
			},
		}
	}

	_, err = discord.ChannelMessageSendComplex(channelId, &messageSend)

	if err != nil {
		fmt.Println("Error sending message: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error sending message: " + err.Error()))
		return
	}

	w.Write([]byte("OK: " + repoName))
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

	http.HandleFunc("/kittycat", webhookRoute)

	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
