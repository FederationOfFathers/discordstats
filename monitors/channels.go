package monitors

import (
	"time"

	"github.com/FederationOfFathers/discordstats/db"
	"github.com/FederationOfFathers/discordstats/discord"
	log "github.com/sirupsen/logrus"
)

type ChannelsMonitor struct {
	DiscordConfig discord.DiscordConfig
	DB            *db.Database
}

func (c *ChannelsMonitor) Start() {
	ticker := time.NewTicker(1 * time.Minute)

	go c.startMonitor()
	go func() {
		for range ticker.C {
			c.startMonitor()
		}
	}()

}

func (c *ChannelsMonitor) startMonitor() {
	log.Info("starting channels monitor")
	guilds, err := c.DB.LatestGuilds()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("unable to get latest guilds")
		return
	}
	log.WithFields(log.Fields{
		"guilds": guilds,
	}).Debug("guilds recieved")
	for _, g := range guilds {
		go c.gatherChannels(g.ID)
	}
}

func (c *ChannelsMonitor) gatherChannels(guildID string) {
	log.Infof("gathering channels %s", guildID)
	channels, err := discord.GuildChannels(c.DiscordConfig, guildID)
	if err != nil {
		log.WithFields(log.Fields{
			"guildID": guildID,
			"error":   err,
		}).Error("unable to get channels")
		return
	}

	for _, ch := range channels {
		lastMsgTime, err := discord.LastChannelMessageTime(c.DiscordConfig, ch.ID)
		if err != nil {
			log.WithFields(log.Fields{
				"guildID":     guildID,
				"channelID":   ch.ID,
				"channelName": ch.Name,
				"error":       err,
			}).Warn("could not get last message time")
			lastMsgTime = time.Unix(1, 0) //add 1 sec to be within range of DB minimum range
		}

		err2 := c.DB.SaveChannel(ch.ID, ch.Name, guildID, lastMsgTime)
		if err2 != nil {
			log.WithFields(log.Fields{
				"guildID":     guildID,
				"channelID":   ch.ID,
				"channelName": ch.Name,
				"lastMsgTime": lastMsgTime,
				"error":       err2,
			}).Error("could not save channel")
		}
	}
}
