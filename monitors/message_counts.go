package monitors

import (
	"time"

	"github.com/FederationOfFathers/discordstats/db"
	"github.com/FederationOfFathers/discordstats/discord"
	"github.com/sirupsen/logrus"
)

const dateFormat = "20060102"

type messageCountsMonitor struct {
	DiscordConfig discord.DiscordConfig
	DB            *db.Database
	log           *logrus.Entry
}

func NewMessageCountsMonitor(database *db.Database, dCfg discord.DiscordConfig) *messageCountsMonitor {
	return &messageCountsMonitor{
		DiscordConfig: dCfg,
		DB:            database,
		log:           logrus.WithField("_module", "monitors.message_counts"),
	}
}

func (m *messageCountsMonitor) Start() {
	ticker := time.NewTicker(60 * time.Minute)

	go m.gatherMessages()
	go func() {
		for range ticker.C {
			m.gatherMessages()
		}
	}()
}

func (m *messageCountsMonitor) gatherMessages() {
	m.log.Info("messages counts monitor started")
	// get channels in DB
	channels, err := m.DB.GetLatestChannels()
	if err != nil {
		m.log.WithFields(logrus.Fields{
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
	m.log.WithFields(logrus.Fields{
		"channels_count": len(channels),
		"channels":       channels,
	}).Info("queueing up channels")
	for _, c := range channels {
		channelsChan <- c
	}

	m.log.Info("channels monitor complete")

}

// messages are counted and added up for each day. times are converted to Eastern timezone (America/New_York)
func (m *messageCountsMonitor) updateChannelMessageCounts(ch *db.Channel) {
	nyTZ, err := time.LoadLocation("America/New_York")
	if err != nil {
		m.log.WithFields(logrus.Fields{
			"error": err,
		}).Error("couldn't get nyTZ")
		return
	}
	if ch.ID == "" {
		m.log.WithFields(logrus.Fields{
			"channel": ch,
		}).Debug("skipping channel without ID")
		return
	}
	// get channel messages, run x number at a time, how? waitgroup?
	m.log.WithFields(logrus.Fields{
		"channelID":   ch.ID,
		"channelName": ch.Name,
	}).Info("gathering channel message data")

	// found latest message count date
	lastCountDate, err := m.DB.LastMessageCountDate(ch.ID) // TODO check this
	lastCountDate = dateOnly(lastCountDate.In(nyTZ))
	if err != nil {
		m.log.WithFields(logrus.Fields{
			"channelID":   ch.ID,
			"channelName": ch.Name,
			"error":       err,
		}).Error("could not get last message date")
		return
	}

	m.log.WithFields(logrus.Fields{
		"channelID":            ch.ID,
		"channelName":          ch.Name,
		"lastMessageCountDate": lastCountDate,
	}).Debug("lastMessageCount retrieved")

	// get Discord connection
	d, err := discord.NewConnection(m.DiscordConfig)
	if err != nil {
		m.log.WithFields(logrus.Fields{
			"channelID": ch.ID,
			"error":     err,
		}).Error("unable to connect to discord")
		return
	}

	// get the messages
	var lastMessageID string // the last message ID in the previous set
	today := dateOnly(time.Now().In(nyTZ))
	currentCountingDate := dateOnly(time.Now().Add(-24 * time.Hour).In(nyTZ)) // the date that is currently being counted, starting with yesterday
	var currentDateMessageCount uint64
	pastLastCountDate := false
	for {
		// get the next 100 messages
		messages, err := d.ChannelMessages(ch.ID, lastMessageID)
		if err != nil {
			m.log.WithFields(logrus.Fields{
				"channelID": ch.ID,
				"guildID":   ch.GuildID,
				"error":     err,
			}).Error("unable to get channel messages")
			break
		}

		// process the current message batch
		for _, msg := range messages {
			msgDate := msg.Timestamp.In(nyTZ)
			lastMessageID = msg.ID
			m.log.WithFields(logrus.Fields{
				"date": msgDate,
			}).Debug("message")

			// skip messages from today
			if msgDate.Unix() >= today.Unix() {
				m.log.WithFields(logrus.Fields{
					"channelID": ch.ID,
					"today":     today,
					"msgDate":   msgDate,
				}).Debug("skipping todays message")
				continue
			}

			// if message is older than the currentCountingDate save and reset counting date
			if msgDate.Unix() < currentCountingDate.Unix() {

				m.log.WithFields(logrus.Fields{
					"msgDate":                 msgDate,
					"currentCountingDate":     currentCountingDate,
					"currentDateMessageCount": currentDateMessageCount,
					"channelID":               ch.ID,
				}).Debug("past current count date. save n reset.")

				// save what we have
				m.saveMessageCount(ch.ID, currentCountingDate, currentDateMessageCount)

				//reset current date and count
				currentCountingDate = dateOnly(msgDate)
				currentDateMessageCount = 1
				continue
			}

			// if message is older than last count (or equal) save the current messageCount and finish
			if msgDate.Unix() <= lastCountDate.Unix() {
				m.log.WithFields(logrus.Fields{
					"channelID":     ch.ID,
					"lastCountDate": lastCountDate,
					"msgDate":       msgDate,
				}).Debug("past last count date")
				pastLastCountDate = true
				break
			}

			currentDateMessageCount++
		}

		if pastLastCountDate {
			break
		}

		// if 100, we may have more, other wise we are done
		if len(messages) < 100 && currentDateMessageCount > 0 {
			m.saveMessageCount(ch.ID, currentCountingDate, currentDateMessageCount)
			break
		}

	}

	m.log.WithFields(logrus.Fields{
		"channelID":   ch.ID,
		"channelName": ch.Name,
	}).Info("finished counting messages")
}

func dateOnly(timestamp time.Time) time.Time {

	t, err := time.ParseInLocation(dateFormat, timestamp.Format(dateFormat), timestamp.Location())
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"timestamp": timestamp,
			"error":     err,
		}).Warn("couldn't parse time to date")
		return timestamp
	}

	return t

}

func (m *messageCountsMonitor) saveMessageCount(channelID string, date time.Time, count uint64) {
	l := m.log.WithFields(logrus.Fields{
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
