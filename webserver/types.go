package main

import "webserver/events"

type RepoWrapper struct {
	Repo events.Repository `json:"repository"`
}
