package core

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"github.com/jon4hz/d4eventbot/d4armory"
	"github.com/mergestat/timediff"
)

type Client struct {
	d4Client *d4armory.Client
}

func New(d4Client *d4armory.Client) *Client {
	return &Client{
		d4Client: d4Client,
	}
}

var (
	tmplFuncs = template.FuncMap{
		"formatTime":          formatTime,
		"formatTimeDiff":      formatTimeDiff,
		"nextHelltide":        nextHelltide,
		"nextHelltideRefresh": nextHelltideRefresh,
		"helltideActive": func(i int) bool {
			return helltideActive(time.Unix(int64(i), 0))
		},
	}
	tmpl *template.Template
)

const tmplFile = "msg.tmpl"

func init() {
	var err error
	tmpl, err = template.New(tmplFile).Funcs(tmplFuncs).ParseFiles(tmplFile)
	if err != nil {
		panic(fmt.Errorf("failed to parse template: %w", err))
	}
}

func formatTime(i int) string {
	t := time.Unix(int64(i), 0)
	diff := timediff.TimeDiff(t)

	return fmt.Sprintf("%s (%s)", t.Format("2006-01-02 15:04:05"), diff)
}

func formatTimeDiff(i int) string {
	t := time.Unix(int64(i), 0)
	diff := timediff.TimeDiff(t)
	return diff
}

var helltideDuration = time.Hour

func nextHelltide(i int) string {
	t := time.Unix(int64(i), 0)
	if t.Before(time.Now()) {
		// helltide starts every 2h 15min
		nextHelltide := addHelltideCooldown(t)
		return formatTime(int(nextHelltide.Unix()))
	}
	return formatTime(i)
}

func addHelltideCooldown(t time.Time) time.Time {
	return t.Add(time.Hour*2 + time.Minute*15)
}

func nextHelltideRefresh(helltide, refresh int) string {
	tHell := time.Unix(int64(helltide), 0)
	if addHelltideCooldown(tHell).Minute() == 0 {
		return "None"
	}
	tRef := time.Unix(int64(refresh), 0)
	next := addHelltideCooldown(tRef)
	return formatTime(int(next.Truncate(time.Hour).Unix()))
}

func helltideActive(t time.Time) bool {
	return t.Sub(time.Now().Add(-helltideDuration)) > 0
}

func (c *Client) GetMessage() (string, error) {
	recentEvents, err := c.d4Client.GetRecent(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get recent events: %w", err)
	}

	var msg bytes.Buffer
	if err := tmpl.Execute(&msg, recentEvents); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return msg.String(), nil
}
