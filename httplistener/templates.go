package httplistener

import (
	"bytes"
	"errors"
	"github.com/irccloud/irccat/util"
	"gopkg.in/go-playground/webhooks.v5/github"
	"strings"
	"text/template"
)

var defaultTemplates = map[string]string{
	"github.release": "[{{b .Repository.Name}}] release {{h .Release.TagName}} has been published by {{g .Release.Author.Login}}: {{.Release.HTMLURL}}",
	"github.push": `[{{b .Repository.Name}}] {{g .Sender.Login}} {{if .Forced}}force-{{end}}{{if .Deleted}}deleted{{else}}pushed{{end}} {{if .Commits}}{{.Commits|len}} commit{{if .Commits|len|lt 1}}s{{end}} to {{end}}{{.Ref|refType}} {{.Ref|refName|h}}: {{.Compare}}
{{range commitLimit . 3}}
 • {{g .Username}} ({{.Sha|truncateSha|h}}): {{trunc .Message 150}}
{{end}}`,
	"github.issue":        "[{{b .Repository.Name}}] {{g .Sender.Login}} {{.Action}} issue #{{.Issue.Number}}: {{.Issue.Title}} {{.Issue.HTMLURL}}",
	"github.issuecomment": "[{{b .Repository.Name}}] {{g .Comment.User.Login}} commented on issue #{{.Issue.Number}}: {{trunc .Comment.Body 150}} {{.Comment.HTMLURL}}",
	"github.pullrequest":  "[{{b .Repository.Name}}] {{g .Sender.Login}} {{if .PullRequest.Merged}}merged{{else}}{{.Action}}{{end}} pull request #{{.PullRequest.Number}} (\x0303{{.PullRequest.Base.Ref}}…{{.PullRequest.Head.Ref}}\x0f): {{.PullRequest.Title}} {{.PullRequest.HTMLURL}}",
	"github.checksuite":   "[{{b .Repository.Name}}] check suite {{b .CheckSuite.Conclusion}}",
	"prometheus.alert": `{{range .Alerts}}[{{b "Prometheus"}}] {{if (eq .Status "firing")}}{{b "alerting"}}{{else}}{{h "resolved"}}{{end}}: {{.Annotations.summary}}
{{end}}`,
}

func refName(ref string) string {
	parts := strings.Split(ref, "/")
	return parts[2]
}

func refType(ref string) string {
	parts := strings.Split(ref, "/")
	if parts[1] == "heads" {
		return "branch"
	} else if parts[1] == "tags" {
		return "tag"
	}
	return ""
}

func truncateSha(sha string) string {
	if len(sha) < 8 {
		return ""
	}
	return sha[:8]
}

// Colour helper functions to try and declutter
func boldFormat(text string) string {
	return "\x02" + text + "\x0f"
}

func greyFormat(text string) string {
	return "\x0314" + text + "\x0f"
}

func highlightFormat(text string) string {
	return "\x0303" + text + "\x0f"
}

func parseTemplates() *template.Template {

	funcMap := template.FuncMap{
		"trunc":       util.Truncate,
		"truncateSha": truncateSha,
		"refType":     refType,
		"refName":     refName,
		"commitLimit": commitLimit,
		"b":           boldFormat,
		"g":           greyFormat,
		"h":           highlightFormat,
	}

	t := template.New("irccat")

	for k, v := range defaultTemplates {
		template.Must(t.New(k).Funcs(funcMap).Parse(v))
	}

	return t
}

func (hl *HTTPListener) renderTemplate(tpl_name string, data interface{}) ([]string, error) {
	var out bytes.Buffer
	t := hl.tpls.Lookup(tpl_name)
	if t == nil {
		return []string{}, errors.New("Nonexistent template")
	}
	t.Execute(&out, data)
	// The \r character is also a delimiter in IRC so strip it out.
	outStr := strings.Replace(out.String(), "\r", "", -1)
	return strings.Split(outStr, "\n"), nil
}

// We need this additional struct and function because the GitHub webhook package represents
// commits as anonymous inner structs and Go's type system is bad. Unless I'm missing something very obvious.
type Commit struct {
	Message  string
	Username string
	Sha      string
}

func commitLimit(pl github.PushPayload, length int) []Commit {
	res := make([]Commit, 0)
	i := 0
	for _, c := range pl.Commits {
		if !c.Distinct {
			continue
		}
		res = append(res, Commit{Message: c.Message, Username: c.Author.Username, Sha: c.ID})
		i += 1
		if i == length {
			break
		}
	}
	return res
}
