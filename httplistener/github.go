package httplistener

import (
	"fmt"
	"github.com/irccloud/irccat/util"
	"github.com/spf13/viper"
	"gopkg.in/go-playground/webhooks.v5/github"
	"net/http"
	"strings"
)

func formatRef(ref string) string {
	parts := strings.Split(ref, "/")
	res := ""
	if parts[1] == "heads" {
		res = "branch "
	} else if parts[1] == "tags" {
		res = "tag "
	}
	return res + parts[2]
}

func handleRelease(payload github.ReleasePayload) ([]string, string, bool) {
	if payload.Action != "published" {
		return []string{""}, "", false
	}

	var release string
	if payload.Release.Name != nil {
		release = *payload.Release.Name
	}
	if release == "" {
		release = payload.Release.TagName
	}

	msg := fmt.Sprintf("[\x02%s\x0f] release \x02%s\x0f has been published by %s: %s",
		payload.Repository.Name, release, payload.Release.Author.Login, payload.Release.HTMLURL)
	return []string{msg}, payload.Repository.Name, true
}

func handlePush(payload github.PushPayload) ([]string, string, bool) {
	var msgs []string

	msgs = append(msgs, fmt.Sprintf("[\x02%s\x0f] %s pushed %d new commits to %s: %s",
		payload.Repository.Name, payload.Sender.Login, len(payload.Commits),
		formatRef(payload.Ref), payload.Compare))

	commits_shown := 0
	for _, commit := range payload.Commits {
		if commits_shown == 3 {
			break
		}
		if !commit.Distinct {
			continue
		}
		msgs = append(msgs, fmt.Sprintf("\t%s: %s", commit.Author.Username, commit.Message))
		commits_shown++
	}
	return msgs, payload.Repository.Name, true
}

func handleIssue(payload github.IssuesPayload) ([]string, string, bool) {
	msg := fmt.Sprintf("[\x02%s\x0f] %s ", payload.Repository.Name, payload.Sender.Login)
	issue_id := fmt.Sprintf("issue #%d", payload.Issue.Number)

	show := true
	switch payload.Action {
	case "opened":
		msg = msg + "opened " + issue_id
	case "closed":
		msg = msg + "closed " + issue_id
	default:
		// Don't know what to do with this, so don't show it
		show = false
	}

	msg = msg + fmt.Sprintf(": %s %s", payload.Issue.Title, payload.Issue.HTMLURL)
	return []string{msg}, payload.Repository.Name, show
}

func handleIssueComment(payload github.IssueCommentPayload) ([]string, string, bool) {
	if payload.Action != "created" {
		return []string{}, payload.Repository.Name, false
	}

	msg := fmt.Sprintf("[\x02%s\x0f] %s commented on issue %d: %s",
		payload.Repository.Name, payload.Comment.User.Login,
		payload.Issue.Number, util.Truncate(payload.Comment.Body, 150))
	return []string{msg}, payload.Repository.Name, true
}

func (hl *HTTPListener) githubHandler(w http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.NotFound(w, request)
		return
	}
	hook, err := github.New() // TODO: webhook secret
	if err != nil {
		return
	}

	// All valid events we want to receive need to be listed here.
	payload, err := hook.Parse(request,
		github.ReleaseEvent, github.PushEvent, github.IssuesEvent, github.IssueCommentEvent)

	if err != nil {
		log.Errorf("Error parsing github webhook: %s", err)
		return
	}

	msgs := []string{}
	repo := ""
	send := false

	switch payload.(type) {
	case github.ReleasePayload:
		msgs, repo, send = handleRelease(payload.(github.ReleasePayload))
	case github.PushPayload:
		msgs, repo, send = handlePush(payload.(github.PushPayload))
	case github.IssuesPayload:
		msgs, repo, send = handleIssue(payload.(github.IssuesPayload))
	case github.IssueCommentPayload:
		msgs, repo, send = handleIssueComment(payload.(github.IssueCommentPayload))
	}

	if send {
		repo = strings.ToLower(repo)
		channelKey := fmt.Sprintf("http.listeners.github.repositories.%s", repo)
		channel := viper.GetString(channelKey)
		if channel == "" {
			log.Infof("%s GitHub event for unrecognised repository %s", request.RemoteAddr, repo)
			return
		}
		log.Infof("%s [%s -> %s] GitHub event received", request.RemoteAddr, repo, channel)
		for _, msg := range msgs {
			hl.irc.Privmsgf(channel, msg)
		}
	}
}
