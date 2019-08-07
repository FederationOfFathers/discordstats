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
}

// Start begins a go routine that checks the list of guilds the bot is a member of.
// It saves/updates the guilds in the DB
func (g *GuildMonitor) Start() {
	ticker := time.NewTicker(15 * time.Minute)

	go g.gatherGuilds()
	// start a go routine to regularly update the guilds
	go func() {
		for range ticker.C {
			g.gatherGuilds()
		}
	}()

}

func (g *GuildMonitor) gatherGuilds() {
	log.Info("gathering guilds")
	guilds, err := discord.Guilds(g.DiscordConfig)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("could not get guilds")
	}
	log.WithFields(log.Fields{
		"guilds_count": len(guilds),
	}).Info("guilds gathered")

	//update/add each guild
	for _, guild := range guilds {
		if err := g.DB.InsertOrUpdateGuild(guild.ID, guild.Name); err != nil {
			log.WithFields(log.Fields{
				"guildID":   guild.ID,
				"guildName": guild.Name,
				"error":     err,
			}).Error("could not add/update guild")
		}
	}
}
