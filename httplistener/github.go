package httplistener

import (
	"fmt"
	"github.com/spf13/viper"
	"gopkg.in/go-playground/webhooks.v5/github"
	"net/http"
	"strings"
)

func interestingIssueAction(action string) bool {
	switch action {
	case "opened", "closed", "reopened":
		return true
	}
	return false
}

func (hl *HTTPListener) githubHandler(w http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.NotFound(w, request)
		return
	}

	hook, err := github.New(github.Options.Secret(viper.GetString("http.listeners.github.secret")))

	if err != nil {
		return
	}

	// All valid events we want to receive need to be listed here.
	payload, err := hook.Parse(request,
		github.ReleaseEvent, github.PushEvent, github.IssuesEvent, github.IssueCommentEvent,
		github.PullRequestEvent, github.CheckSuiteEvent)

	if err != nil {
		if err == github.ErrEventNotFound {
			// We've received an event we don't need to handle, return normally
			return
		}
		log.Warningf("Error parsing github webhook: %s", err)
		http.Error(w, "Error processing webhook", http.StatusBadRequest)
		return
	}

	msgs := []string{}
	repo := ""
	send := false

	switch payload.(type) {
	case github.ReleasePayload:
		pl := payload.(github.ReleasePayload)
		if pl.Action == "published" {
			send = true
			msgs, err = hl.renderTemplate("github.release", payload)
			repo = pl.Repository.Name
		}
	case github.PushPayload:
		pl := payload.(github.PushPayload)
		send = true
		msgs, err = hl.renderTemplate("github.push", payload)
		repo = pl.Repository.Name
	case github.IssuesPayload:
		pl := payload.(github.IssuesPayload)
		if interestingIssueAction(pl.Action) {
			send = true
			msgs, err = hl.renderTemplate("github.issue", payload)
			repo = pl.Repository.Name
		}
	case github.IssueCommentPayload:
		pl := payload.(github.IssueCommentPayload)
		if pl.Action == "created" {
			send = true
			msgs, err = hl.renderTemplate("github.issuecomment", payload)
			repo = pl.Repository.Name
		}
	case github.PullRequestPayload:
		pl := payload.(github.PullRequestPayload)
		if interestingIssueAction(pl.Action) {
			send = true
			msgs, err = hl.renderTemplate("github.pullrequest", payload)
			repo = pl.Repository.Name
		}
	case github.CheckSuitePayload:
		pl := payload.(github.CheckSuitePayload)
		if pl.CheckSuite.Status == "completed" && pl.CheckSuite.Conclusion == "failure" {
			send = true
			msgs, err = hl.renderTemplate("github.checksuite", payload)
			repo = pl.Repository.Name
		}
	}

	if err != nil {
		log.Errorf("Error rendering GitHub event template: %s", err)
		return
	}

	if send {
		repo = strings.ToLower(repo)
		channel := viper.GetString(fmt.Sprintf("http.listeners.github.repositories.%s", repo))
		if channel == "" {
			channel = viper.GetString("http.listeners.github.default_channel")
		}

		if channel == "" {
			log.Infof("%s GitHub event for unrecognised repository %s", request.RemoteAddr, repo)
			return
		}

		log.Infof("%s [%s -> %s] GitHub event received", request.RemoteAddr, repo, channel)
		for _, msg := range msgs {
			hl.irc.Privmsg(channel, msg)
		}
	}
}
