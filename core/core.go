package core

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/jon4hz/d4eventbot/d4armory"
	"github.com/mergestat/timediff"
)

type Client struct {
	d4Client *d4armory.Client
	cache    *ttlcache.Cache[string, string]
}

func New(d4Client *d4armory.Client) *Client {
	c := &Client{
		d4Client: d4Client,
		cache: ttlcache.New[string, string](
			ttlcache.WithTTL[string, string](time.Second * 5),
		),
	}

	go c.cache.Start()

	return c
}

var (
	tmplFuncs = template.FuncMap{
		"formatTime":     formatTime,
		"formatTimeDiff": formatTimeDiff,
		"nextHelltide": func(i int) string {
			return formatTime(int(nextHelltide(i).Unix()))
		},
		"nextHelltideDiff": nextHelltideDiff,
		"nextHelltideRefresh": func(i, r int) string {
			n := nextHelltideRefresh(i, r)
			if n.IsZero() {
				return "None"
			}
			return formatTime(int(n.Unix()))
		},
		"nextHelltideRefreshDiff": func(i, r int) string {
			n := nextHelltideRefresh(i, r)
			if n.IsZero() {
				return ""
			}
			return formatTimeDiff(int(n.Unix()))
		},
		"helltideActive": func(i int) bool {
			return helltideActive(time.Unix(int64(i), 0))
		},
		"mapZoneName": func(z string) string {
			return mapZoneName(z)
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
	return time.Unix(int64(i), 0).Format("2006-01-02 15:04:05")
}

func formatTimeDiff(i int) string {
	t := time.Unix(int64(i), 0)
	diff := timediff.TimeDiff(t)
	return diff
}

var helltideDuration = time.Hour

func nextHelltide(i int) time.Time {
	t := time.Unix(int64(i), 0)
	if t.Before(time.Now()) {
		// helltide starts every 2h 15min
		nextHelltide := addHelltideCooldown(t)
		return nextHelltide
	}
	return t
}

func nextHelltideDiff(i int) string {
	t := nextHelltide(i)
	return formatTimeDiff(int(t.Unix()))
}

func addHelltideCooldown(t time.Time) time.Time {
	return t.Add(time.Hour*2 + time.Minute*15)
}

func nextHelltideRefresh(helltide, refresh int) time.Time {
	tHell := time.Unix(int64(helltide), 0)

	if helltideActive(tHell) && refresh == 0 {
		return time.Time{}
	}

	tRef := time.Unix(int64(refresh), 0)
	if helltideActive(tHell) {
		return tRef
	}

	next := addHelltideCooldown(tRef)
	return next.Truncate(time.Hour)
}

func helltideActive(t time.Time) bool {
	return t.Sub(time.Now().Add(-helltideDuration)) > 0
}

func mapZoneName(z string) string {
	if zone, ok := d4armory.ZoneMap[z]; ok {
		return zone
	}
	return z
}

func (c *Client) GetMessage() (string, error) {
	cachedMsg := c.cache.Get("msg", ttlcache.WithDisableTouchOnHit[string, string]())
	if cachedMsg != nil && !cachedMsg.IsExpired() && cachedMsg.Value() != "" {
		return cachedMsg.Value(), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	recentEvents, err := c.d4Client.GetRecent(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get recent events: %w", err)
	}

	var msg bytes.Buffer
	if err := tmpl.Execute(&msg, recentEvents); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	stringMsg := msg.String()
	c.cache.Set("msg", stringMsg, ttlcache.DefaultTTL)
	return stringMsg, nil
}
