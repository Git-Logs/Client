package events

import jsoniter "github.com/json-iterator/go"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type User struct {
	Login            string `json:"login"`
	ID               int    `json:"id"`
	AvatarURL        string `json:"avatar_url"`
	URL              string `json:"url"`
	HTMLURL          string `json:"html_url"`
	OrganizationsURL string `json:"organizations_url"`
}

type Repository struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Owner       User   `json:"owner"`
	HTMLURL     string `json:"html_url"`
	CommitsURL  string `json:"commits_url"`
}

type Issue struct {
	ID      int    `json:"id"`
	Number  int    `json:"number"`
	State   string `json:"state"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
	URL     string `json:"url"`
	User    User   `json:"user"`
}

type PullRequestCommit struct {
	Repo       Repository `json:"repo"`
	ID         int        `json:"id"`
	Number     int        `json:"number"`
	State      string     `json:"state"`
	Title      string     `json:"title"`
	Body       string     `json:"body"`
	HTMLURL    string     `json:"html_url"`
	URL        string     `json:"url"`
	Ref        string     `json:"ref"`
	Label      string     `json:"label"`
	User       User       `json:"user"`
	CommitsURL string     `json:"commits_url"`
}

type PullRequest struct {
	ID      int                   `json:"id"`
	Number  int                   `json:"number"`
	State   string                `json:"state"`
	Locked  bool                  `json:"locked"`
	Title   string                `json:"title"`
	Body    string                `json:"body"`
	HTMLURL string                `json:"html_url"`
	URL     string                `json:"url"`
	User    User                  `json:"user"`
	Base    PullRequestCommit `json:"base"`
	Head    PullRequestCommit `json:"head"`
}
