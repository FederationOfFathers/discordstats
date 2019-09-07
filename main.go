package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/FederationOfFathers/discordstats/db"
	"github.com/FederationOfFathers/discordstats/discord"
	"github.com/FederationOfFathers/discordstats/monitors"
	"github.com/heroku/rollrus"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

func main() {

	log.SetReportCaller(true)

	//assume local by default
	env := "local"
	if e, ok := os.LookupEnv("ENVIRONMENT"); ok {
		env = e
	}

	if env == "production" {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}
	log.Debugf("start in %s env", env)

	//check if rollbar
	if rollbarToken, ok := os.LookupEnv("ROLLBAR_TOKEN"); ok {
		rollrus.SetupLoggingForLevels(rollbarToken, env, []log.Level{log.InfoLevel, log.WarnLevel, log.ErrorLevel, log.FatalLevel, log.PanicLevel})
	}

	log.Info("Discord stats starting")
	defer handlePanic()

	// get discord Bot Token
	var dCfg discord.DiscordConfig
	envconfig.MustProcess("discord", &dCfg)
	if dCfg.BotToken == "" {
		log.Fatal("empty discord bot token")
	}

	// get DB config
	var dbCfg db.DBConfig
	envconfig.MustProcess("db", &dbCfg)
	if dbCfg.ConnectionString == "" {
		log.Fatal("empty connection string is not valid")
	}

	// connect to DB
	dataB := db.Connect(dbCfg)
	defer dataB.Close()

	// initialize db with tables and things
	log.Info("Initializing DB")
	dataB.Initialize()

	// start the guild monitor
	gm := monitors.NewGuildMonitor(&dataB, dCfg)
	gm.Start()

	// start channels monitor
	cm := monitors.NewChannelsMonitor(&dataB, dCfg)
	cm.Start()

	// start message counts monitor
	mcm := monitors.NewMessageCountsMonitor(&dataB, dCfg)
	mcm.Start()

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

	<-done
	log.Info("stats exited")
}
