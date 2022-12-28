package main

type GithubWebhook struct {
	Ref          string         `json:"ref"`           // common
	MasterBranch string         `json:"master_branch"` // create
	Description  string         `json:"description"`   // create
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
	Repo struct { // common
		ID          int    `json:"id"`
		Name        string `json:"name"`
		FullName    string `json:"full_name"`
		Description string `json:"description"`
		URL         string `json:"url"`
		Owner       struct {
			Login            string `json:"login"`
			ID               int    `json:"id"`
			AvatarURL        string `json:"avatar_url"`
			URL              string `json:"url"`
			HTMLURL          string `json:"html_url"`
			OrganizationsURL string `json:"organizations_url"`
		} `json:"owner"`
		HTMLURL    string `json:"html_url"`
		CommitsURL string `json:"commits_url"`
	} `json:"repository"`
	Pusher struct { // push
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"pusher,omitempty"`
	Sender struct { // common
		Login            string `json:"login"`
		ID               int    `json:"id"`
		AvatarURL        string `json:"avatar_url"`
		URL              string `json:"url"`
		HTMLURL          string `json:"html_url"`
		OrganizationsURL string `json:"organizations_url"`
	} `json:"sender,omitempty"`
	HeadCommit struct { // common
		ID      string `json:"id"`
		Message string `json:"message"`
		Author  struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"author,omitempty"`
	} `json:"head_commit,omitempty"`
	PullRequest struct {
		ID      int    `json:"id"`
		Number  int    `json:"number"`
		State   string `json:"state"`
		Locked  bool   `json:"locked"`
		Title   string `json:"title"`
		Body    string `json:"body"`
		HTMLURL string `json:"html_url"`
		URL     string `json:"url"`
		User    struct {
			Login            string `json:"login"`
			ID               int    `json:"id"`
			AvatarURL        string `json:"avatar_url"`
			URL              string `json:"url"`
			HTMLURL          string `json:"html_url"`
			OrganizationsURL string `json:"organizations_url"`
		} `json:"user"`
		Base struct {
			Repo struct {
				ID          int    `json:"id"`
				Name        string `json:"name"`
				FullName    string `json:"full_name"`
				Description string `json:"description"`
				URL         string `json:"url"`
				Owner       struct {
					Login            string `json:"login"`
					ID               int    `json:"id"`
					AvatarURL        string `json:"avatar_url"`
					URL              string `json:"url"`
					HTMLURL          string `json:"html_url"`
					OrganizationsURL string `json:"organizations_url"`
				} `json:"owner"`
				HTMLURL    string `json:"html_url"`
				CommitsURL string `json:"commits_url"`
			} `json:"repo"`
			ID      int    `json:"id"`
			Number  int    `json:"number"`
			State   string `json:"state"`
			Title   string `json:"title"`
			Body    string `json:"body"`
			HTMLURL string `json:"html_url"`
			URL     string `json:"url"`
			Ref     string `json:"ref"`
			Label   string `json:"label"`
			User    struct {
				Login            string `json:"login"`
				ID               int    `json:"id"`
				AvatarURL        string `json:"avatar_url"`
				URL              string `json:"url"`
				HTMLURL          string `json:"html_url"`
				OrganizationsURL string `json:"organizations_url"`
			} `json:"user"`
			CommitsURL string `json:"commits_url"`
		} `json:"base"`
		Head struct {
			Repo struct {
				ID          int    `json:"id"`
				Name        string `json:"name"`
				FullName    string `json:"full_name"`
				Description string `json:"description"`
				URL         string `json:"url"`
				Owner       struct {
					Login            string `json:"login"`
					ID               int    `json:"id"`
					AvatarURL        string `json:"avatar_url"`
					URL              string `json:"url"`
					HTMLURL          string `json:"html_url"`
					OrganizationsURL string `json:"organizations_url"`
				} `json:"owner"`
				HTMLURL    string `json:"html_url"`
				CommitsURL string `json:"commits_url"`
			} `json:"repo"`
			ID      int    `json:"id"`
			Number  int    `json:"number"`
			State   string `json:"state"`
			Title   string `json:"title"`
			Body    string `json:"body"`
			HTMLURL string `json:"html_url"`
			URL     string `json:"url"`
			Ref     string `json:"ref"`
			Label   string `json:"label"`
			User    struct {
				Login            string `json:"login"`
				ID               int    `json:"id"`
				AvatarURL        string `json:"avatar_url"`
				URL              string `json:"url"`
				HTMLURL          string `json:"html_url"`
				OrganizationsURL string `json:"organizations_url"`
			} `json:"user"`
			CommitsURL string `json:"commits_url"`
		} `json:"head"`
	} `json:"pull_request,omitempty"`
	Issue struct {
		ID      int    `json:"id"`
		Number  int    `json:"number"`
		State   string `json:"state"`
		Title   string `json:"title"`
		Body    string `json:"body"`
		HTMLURL string `json:"html_url"`
		URL     string `json:"url"`
		User    struct {
			Login            string `json:"login"`
			ID               int    `json:"id"`
			AvatarURL        string `json:"avatar_url"`
			URL              string `json:"url"`
			HTMLURL          string `json:"html_url"`
			OrganizationsURL string `json:"organizations_url"`
		} `json:"user"`
	} `json:"issue,omitempty"`
}
