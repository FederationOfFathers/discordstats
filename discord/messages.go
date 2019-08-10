package discord

import (
	"time"

	"github.com/pkg/errors"
)

type Message struct {
	ID        string
	Timestamp time.Time
}

// ChannelMessages gets the latest 100 messages before the id of the message provided in `before`. leave `before` emtpy to get the latest
func (d *DiscordConnection) ChannelMessages(channelID, before string) ([]Message, error) {
	var messages []Message
	ms, err := d.d.ChannelMessages(channelID, 100, before, "", "")
	if err != nil {
		return messages, errors.Wrap(err, "unable to get Discord messages")
	}

	for _, m := range ms {

		timestamp, err := m.Timestamp.Parse()
		if err != nil {
			return []Message{}, errors.Wrap(err, "timestamp parse failed")
		}

		messages = append(messages, Message{
			ID:        m.ID,
			Timestamp: timestamp,
		})
	}

	return messages, nil
}
