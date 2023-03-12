package telegram

import tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Config struct {
	Token string
}

type Bot struct {
	client *tg.BotAPI
}

func NewBot(config Config) (*Bot, error) {
	client, err := tg.NewBotAPI(config.Token)
	if err != nil {
		return nil, err
	}
	return &Bot{
		client: client,
	}, nil
}

func (b *Bot) GetUpdates() tg.UpdatesChannel {
	return b.client.GetUpdatesChan(tg.UpdateConfig{})
}

func (b *Bot) Send(c tg.Chattable) (tg.Message, error) {
	return b.client.Send(c)
}

func (b *Bot) Shutdown() {
	b.client.StopReceivingUpdates()
}
