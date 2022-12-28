package main

import "webserver/events"

type GithubWebhook struct {
	Ref          string         `json:"ref"`           // common
	MasterBranch string         `json:"master_branch"` // create
	Description  string         `json:"description"`   // create
	Context      string         `json:"context"`       // status
	State        string         `json:"state"`         // status
	TargetURL    string         `json:"target_url"`    // status
	Name         string         `json:"name"`          // status
	PusherType   string         `json:"pusher_type"`   // create
	RefType      string         `json:"ref_type"`      // create
	BaseRef      string         `json:"base_ref,omitempty"`
	Action       string         `json:"action"` // common
	ActionsMeta  map[string]any `json:"actions_meta,omitempty"`
	CheckSuite   struct {
		ID         int    `json:"id"`
		After      string `json:"after,omitempty"`
		HeadBranch string `json:"head_branch,omitempty"`
		HeadSHA    string `json:"head_sha,omitempty"`
		Status     string `json:"status,omitempty"`
		Conclusion string `json:"conclusion,omitempty"`
		URL        string `json:"url,omitempty"`
		Before     string `json:"before,omitempty"`
		HeadCommit struct {
			ID        string `json:"id,omitempty"`
			TreeID    string `json:"tree_id,omitempty"`
			Message   string `json:"message,omitempty"`
			Timestamp string `json:"timestamp,omitempty"`
			Author    struct {
				Name     string `json:"name,omitempty"`
				Email    string `json:"email,omitempty"`
				Username string `json:"username,omitempty"`
			} `json:"author,omitempty"`
			Committer struct {
				Name     string `json:"name,omitempty"`
				Email    string `json:"email,omitempty"`
				Username string `json:"username,omitempty"`
			} `json:"committer,omitempty"`
		} `json:"head_commit,omitempty"`
	} `json:"check_suite,omitempty"`
	Commits []struct { // push
		ID        string `json:"id"`
		Message   string `json:"message"`
		Timestamp string `json:"timestamp"`
		URL       string `json:"url"`
		Author    struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"author"`
	}
	Repo   events.Repository `json:"repository"`
	Sender events.User     `json:"sender,omitempty"`
	Commit struct {          // status
		HTMLURL string `json:"html_url"`
		SHA     string `json:"sha"`
		Commit  struct {
			ID      string `json:"id"`
			Message string `json:"message"`
			Author  struct {
				Name     string `json:"name"`
				Email    string `json:"email"`
				Username string `json:"username"`
			} `json:"author,omitempty"`
		} `json:"commit"`
	} `json:"commit"`
	HeadCommit struct { // common
		ID      string `json:"id"`
		Message string `json:"message"`
		Author  struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"author,omitempty"`
	} `json:"head_commit,omitempty"`
}
