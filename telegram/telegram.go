package telegram

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/jon4hz/d4eventbot/core"
)

type Client struct {
	bot  *gotgbot.Bot
	core *core.Client
}

func New(token string, coreClient *core.Client) (*Client, error) {
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
		bot:  b,
		core: coreClient,
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
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("refresh"), c.refreshMessage))

	err := updater.StartPolling(c.bot, &ext.PollingOpts{
		DropPendingUpdates: false,
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

func (c *Client) startHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	msg, err := c.core.GetMessage()
	if err != nil {
		return err
	}

	_, err = ctx.EffectiveChat.SendMessage(b, "<code>"+msg+"</code>", &gotgbot.SendMessageOpts{
		ParseMode: "html",
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{{Text: "  Refresh", CallbackData: "refresh"}},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}
	return nil
}

func (c *Client) refreshMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery
	_, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
		Text: "Refreshing...",
	})
	if err != nil {
		return fmt.Errorf("failed to answer callback query: %w", err)
	}

	msg, err := c.core.GetMessage()
	if err != nil {
		return err
	}

	_, _, err = cb.Message.EditText(b, "<code>"+msg+"</code>", &gotgbot.EditMessageTextOpts{
		ParseMode: "html",
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{{Text: "  Refresh", CallbackData: "refresh"}},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to edit message: %w", err)
	}

	return nil
}
