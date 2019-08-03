package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/FederationOfFathers/discordstats/db"
	"github.com/FederationOfFathers/discordstats/discord"
	"github.com/FederationOfFathers/discordstats/monitors"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("Discord stats starting")
	defer handlePanic()

	var dCfg discord.DiscordConfig
	envconfig.MustProcess("discord", &dCfg)
	if dCfg.BotToken == "" {
		log.Fatal("empty discord bot token")
	}

	// test new guild
	var dbCfg db.DBConfig
	envconfig.MustProcess("db", &dbCfg)
	if dbCfg.ConnectionString == "" {
		log.Fatal("empty connection string is not valid")
	}
	s := db.Connect(dbCfg)
	// if err != nil {
	// 	log.WithFields(log.Fields{
	// 		"error": err.Error(),
	// 	}).Fatal("unable to connect to db")
	// }
	defer s.Close()

	db.Initialize(s)

	gm := &monitors.GuildMonitor{
		DiscordConfig: dCfg,
	}
	guildUpdates := gm.Start()
	go func() {
		for gs := range guildUpdates {
			log.WithFields(log.Fields{
				"guilds": gs,
			}).Info("guilds updated")
		}
	}()

	// guild := &db.Guild{
	// 	GuildID: "1231412312312314",
	// }

	log.Info("record created")

	awaitSignal()
}

func handlePanic() {
	if r := recover(); r != nil {
		log.WithFields(log.Fields{
			"panic_state": r,
		}).Error("exiting on panic")
	}
}

func awaitSignal() {
	// wait for kill
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.WithFields(log.Fields{
			"signal": sig,
		}).Info("signal recieved")
		done <- true
	}()

	// start monitos that write to DB
	// discord_stats_guilds
	// discord_stats_channel_messages
	// discord_stats_presence_updates

	<-done
	log.Info("stats exited")
}
