package monitors

import (
	"time"

	"github.com/FederationOfFathers/discordstats/db"
	"github.com/FederationOfFathers/discordstats/discord"
	"github.com/sirupsen/logrus"
)

// guilds

type guildMonitor struct {
	DB            *db.Database
	DiscordConfig discord.DiscordConfig
	log           *logrus.Entry
}

func NewGuildMonitor(database *db.Database, discordConfig discord.DiscordConfig) *guildMonitor {

	return &guildMonitor{
		DB:            database,
		DiscordConfig: discordConfig,
		log:           logrus.WithField("_module", "monitor.guilds"),
	}
}

// Start begins a go routine that checks the list of guilds the bot is a member of.
// It saves/updates the guilds in the DB
func (g *guildMonitor) Start() {
	ticker := time.NewTicker(15 * time.Minute)

	go g.gatherGuilds()
	// start a go routine to regularly update the guilds
	go func() {
		for range ticker.C {
			g.gatherGuilds()
		}
	}()

}

func (g *guildMonitor) gatherGuilds() {
	g.log.Info("gathering guilds")
	guilds, err := discord.Guilds(g.DiscordConfig)
	if err != nil {
		g.log.WithFields(logrus.Fields{
			"error": err,
		}).Error("could not get guilds")
	}
	g.log.WithFields(logrus.Fields{
		"guilds_count": len(guilds),
	}).Info("guilds gathered")

	//update/add each guild
	for _, guild := range guilds {
		if err := g.DB.InsertOrUpdateGuild(guild.ID, guild.Name); err != nil {
			g.log.WithFields(logrus.Fields{
				"guildID":   guild.ID,
				"guildName": guild.Name,
				"error":     err,
			}).Error("could not add/update guild")
		}
	}
}
