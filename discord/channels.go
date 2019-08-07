package discord

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

type Channel struct {
	ID   string
	Name string
}

func GuildChannels(dCfg DiscordConfig, guildID string) ([]Channel, error) {
	var channels []Channel
	d, err := connect(dCfg)
	if err != nil {
		return channels, errors.Wrap(err, "could not connect to discord")
	}

	chSet, err := d.GuildChannels(guildID)
	if err != nil {
		return channels, errors.Wrapf(err, "unable to reitreve guild channels (%s)", guildID)
	}

	for _, ch := range chSet {
		// skip non text channels
		if ch.Type != discordgo.ChannelTypeGuildText {
			continue
		}
		channels = append(channels, Channel{
			ID:   ch.ID,
			Name: ch.Name,
		})
	}

	return channels, nil
}

func LastChannelMessageTime(dCfg DiscordConfig, channelID string) (time.Time, error) {
	var t time.Time
	d, err := connect(dCfg)
	if err != nil {
		return t, err
	}

	messages, err := d.ChannelMessages(channelID, 1, "", "", "")
	if err != nil {
		return t, errors.Wrapf(err, "unable to get channel messages (%s)", channelID)
	}

	if len(messages) <= 0 {
		return t, fmt.Errorf("no messages were found for channel %s", channelID)
	}
	m := messages[0]
	t, err = m.Timestamp.Parse()
	if err != nil {
		return t, errors.Wrap(err, "could not parse message time")
	}
	return t, nil
}
