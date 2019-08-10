package monitors

import (
	"time"

	"github.com/FederationOfFathers/discordstats/db"
	"github.com/FederationOfFathers/discordstats/discord"
	log "github.com/sirupsen/logrus"
)

const dateFormat = "20060102"

type MessageCountsMonitor struct {
	DiscordConfig discord.DiscordConfig
	DB            *db.Database
}

func (m *MessageCountsMonitor) Start() {
	ticker := time.NewTicker(60 * time.Minute)

	go m.gatherMessages()
	go func() {
		for range ticker.C {
			m.gatherMessages()
		}
	}()
}

func (m *MessageCountsMonitor) gatherMessages() {
	log.Info("messages counts monitor started")
	// get channels in DB
	channels, err := m.DB.GetLatestChannels()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("could not gather channels for messages")
		return
	}

	// create a channel and process it in a new goroutine
	channelsChan := make(chan (db.Channel), 5)
	defer close(channelsChan)

	go func() {
		for {
			select {
			case ch := <-channelsChan:
				m.updateChannelMessageCounts(&ch)
			}
		}
	}()

	// queue the channels into the chan
	log.WithFields(log.Fields{
		"channels": len(channels),
	}).Info("queueing up channels")
	for _, c := range channels {
		channelsChan <- c
	}

	log.Info("channels monitor complete")

}

// messages are counted and added up for each day. times are converted to Eastern timezone (America/New_York)
func (m *MessageCountsMonitor) updateChannelMessageCounts(ch *db.Channel) {
	nyTZ, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("couldn't get nyTZ")
		return
	}
	if ch.ID == "" {
		return
	}
	// get channel messages, run x number at a time, how? waitgroup?
	log.WithFields(log.Fields{
		"channelID":   ch.ID,
		"channelName": ch.Name,
	}).Info("gathering channel message data")

	// found latest message count date
	lastCountDate, err := m.DB.LastMessageCountDate(ch.ID) // TODO check this
	lastCountDate = dateOnly(lastCountDate.In(nyTZ))
	if err != nil {
		log.WithFields(log.Fields{
			"channelID":   ch.ID,
			"channelName": ch.Name,
			"error":       err,
		}).Error("could not get last message date")
		return
	}

	// get Discord connection
	d, err := discord.NewConnection(m.DiscordConfig)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("unable to connect to discord")
		return
	}

	// get the messages
	var lastMessageID string // the last message ID in the previous set
	today := dateOnly(time.Now().In(nyTZ))
	currentCountingDate := dateOnly(time.Now().In(nyTZ)) // the date that is currently being counted
	var currentDateMessageCount uint64
	for {
		// get the next 100 messages
		messages, err := d.ChannelMessages(ch.ID, lastMessageID)
		if err != nil {
			log.WithFields(log.Fields{
				"channelID": ch.ID,
				"guildID":   ch.GuildID,
				"error":     err,
			}).Error("unable to get channel messages")
			break
		}

		// process the messages
		for _, msg := range messages {
			msgDate := msg.Timestamp.In(nyTZ)
			lastMessageID = msg.ID
			log.WithFields(log.Fields{
				"date": msgDate,
			}).Debug("message")

			// skip messages from today
			if msgDate.Unix() >= today.Unix() {
				log.WithFields(log.Fields{
					"today":   today,
					"msgDate": msgDate,
				}).Debug("skipping todays message")
				continue
			}

			// if message is older than last count (or equal) save the current messageCount and finish
			if msgDate.Unix() <= lastCountDate.Unix() {
				// do the save
				log.Debug("past last count date")
				break
			}

			// if message is older than the currentCountingDate save and update counting date
			if msgDate.Unix() < currentCountingDate.Unix() {
				log.Debug("past current count date. save n reset.")
				m.saveMessageCount(ch.ID, currentCountingDate, currentDateMessageCount)
				currentCountingDate = dateOnly(msgDate)
				currentDateMessageCount = 0
			}

			currentDateMessageCount++
		}

		// if 100, we may have more, other wise we are done
		if len(messages) < 100 {
			m.saveMessageCount(ch.ID, currentCountingDate, currentDateMessageCount)
			break
		}

	}

	log.WithFields(log.Fields{
		"channelID":   ch.ID,
		"channelName": ch.Name,
	}).Info("finished counting messages")
}

func dateOnly(timestamp time.Time) time.Time {

	if t, err := time.ParseInLocation(dateFormat, timestamp.Format(dateFormat), timestamp.Location()); err != nil {
		log.WithFields(log.Fields{
			"timestamp": timestamp,
			"error":     err,
		}).Warn("couldn't parse time to date")
		return timestamp
	} else {
		return t
	}

}

func (m *MessageCountsMonitor) saveMessageCount(channelID string, date time.Time, count uint64) {
	l := log.WithFields(log.Fields{
		"channelID": channelID,
		"date":      date.Format("01-02-2006"),
		"count":     count,
	})
	if err := m.DB.SaveMessageCount(channelID, date, count); err != nil {
		l.Error("unable to save message count")
	} else {
		l.Info("saved")
	}
}
