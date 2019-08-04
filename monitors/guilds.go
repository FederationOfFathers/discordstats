package monitors

import (
	"time"

	"github.com/FederationOfFathers/discordstats/db"
	"github.com/FederationOfFathers/discordstats/discord"
	log "github.com/sirupsen/logrus"
)

// guilds

type GuildMonitor struct {
	DB            *db.Database
	DiscordConfig discord.DiscordConfig
	GuildIDs      []string
}

// Start begins a go routine that checks the list of guilds the bot is a member of.
// It saves/updates the guilds in the DB
func (g *GuildMonitor) Start() {
	ticker := time.NewTicker(10 * time.Second)

	// start a go routine to regularly update the guilds
	go func() {
		for range ticker.C {
			guildIDs, err := discord.Guilds(g.DiscordConfig)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("could not get guilds")
			}
			log.WithFields(log.Fields{
				"guildIDs": guildIDs,
			}).Info("guild IDs gathered")

			//update/add each guild
			for _, guildID := range guildIDs {
				if err := g.DB.InsertOrUpdateGuild(guildID); err != nil {
					log.WithFields(log.Fields{
						"guildID": guildID,
						"error":   err,
					}).Error("could not add/update guild")
				}
			}
		}
	}()

}
