package discord

// Performs all of the discord functions withou exposing the underlying discord implementation library.

import (
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

type DiscordConfig struct {
	BotToken string `split_words:"true"`
}

func connect(dCfg DiscordConfig) (*discordgo.Session, error) {
	d, err := discordgo.New("Bot " + dCfg.BotToken)
	if err != nil {
		return nil, errors.Wrap(err, "unable to connect to Discord")
	}

	return d, nil
}

func Guilds(dCfg DiscordConfig) ([]string, error) {

	d, err := connect(dCfg)
	if err != nil {
		return nil, errors.Wrap(err, "could not get guilds")
	}

	var guildIDs []string
	lastGuildID := "0"

	// start a loop to iterate over each set of 100 guilds from the API
	for {
		guilds, err := d.UserGuilds(100, "", lastGuildID)
		if err != nil {
			return guildIDs, errors.Wrap(err, "guilds call failed")
		}

		// add each guild id to the slice
		for _, guild := range guilds {
			lastGuildID = guild.ID
			guildIDs = append(guildIDs, guild.ID)
		}

		// if we had < 100 guilds, then we've reached the last set
		if len(guilds) < 100 {
			break
		}

	}

	return guildIDs, nil
}
