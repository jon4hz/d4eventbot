package telegram

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/jon4hz/d4eventbot/d4armory"
	"github.com/mergestat/timediff"
)

type Client struct {
	bot      *gotgbot.Bot
	d4Client *d4armory.Client
}

func New(token string, d4Client *d4armory.Client) (*Client, error) {
	b, err := gotgbot.NewBot(token, &gotgbot.BotOpts{
		Client: http.Client{},
		DefaultRequestOpts: &gotgbot.RequestOpts{
			Timeout: gotgbot.DefaultTimeout,
			APIURL:  gotgbot.DefaultAPIURL,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}
	return &Client{
		bot:      b,
		d4Client: d4Client,
	}, nil
}

func (c *Client) Run() error {
	updater := ext.NewUpdater(&ext.UpdaterOpts{
		Dispatcher: ext.NewDispatcher(&ext.DispatcherOpts{
			// If an error is returned by a handler, log it and continue going.
			Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
				log.Println("an error occurred while handling update:", err.Error())
				return ext.DispatcherActionNoop
			},
			MaxRoutines: ext.DefaultMaxRoutines,
		}),
	})
	dispatcher := updater.Dispatcher

	dispatcher.AddHandler(handlers.NewCommand("start", c.startHandler))

	err := updater.StartPolling(c.bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to start polling: %w", err)
	}
	log.Printf("%s has been started...\n", c.bot.User.Username)

	updater.Idle()
	return nil
}

var (
	tmplFuncs = template.FuncMap{
		"formatTime":          formatTime,
		"nextHelltide":        nextHelltide,
		"nextHelltideRefresh": nextHelltideRefresh,
	}
	tmpl *template.Template
)

func formatTime(i int) string {
	t := time.Unix(int64(i), 0)
	diff := timediff.TimeDiff(t)

	return fmt.Sprintf("%s (%s)", t.Format("2006-01-02 15:04:05"), diff)
}

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

func nextHelltideRefresh(i int) string {
	t := time.Unix(int64(i), 0)
	next := addHelltideCooldown(t)
	if next.Minute() == 0 {
		return "No refresh"
	}
	return formatTime(int(next.Truncate(time.Hour).Unix()))
}

const tmplFile = "msg.tmpl"

func init() {
	var err error
	tmpl, err = template.New(tmplFile).Funcs(tmplFuncs).ParseFiles(tmplFile)
	if err != nil {
		panic(fmt.Errorf("failed to parse template: %w", err))
	}
}

func (c *Client) startHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	recentEvents, err := c.d4Client.GetRecent(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get recent events: %w", err)
	}

	var msg bytes.Buffer
	if err := tmpl.Execute(&msg, recentEvents); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	_, err = ctx.EffectiveChat.SendMessage(b, msg.String(), &gotgbot.SendMessageOpts{
		ParseMode: "html",
	})
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}
	return nil
}
