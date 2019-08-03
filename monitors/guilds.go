package monitors

import (
	"time"

	"github.com/FederationOfFathers/discordstats/discord"
	log "github.com/sirupsen/logrus"
)

// guilds

type GuildMonitor struct {
	DiscordConfig discord.DiscordConfig
	GuildIDs      []string
}

// Start begins a go routine that checks the list of guilds the bot is a member of.
// It then send a slice of the guild IDs to the channel that is returned
func (g *GuildMonitor) Start() <-chan []string {
	ticker := time.NewTicker(10 * time.Second)
	updates := make(chan []string, 1)

	// start a go routine to regularly update the guilds
	go func() {
		for range ticker.C {
			guildsIds, err := discord.Guilds(g.DiscordConfig)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("could not get guilds")
			}
			updates <- guildsIds
		}
		close(updates)
	}()

	return updates

}
