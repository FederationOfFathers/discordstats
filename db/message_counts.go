package db

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

const messageCountsTableName = tablePrefix + "_message_counts"
const messageCountsSchema = `CREATE TABLE IF NOT EXISTS ` + messageCountsTableName + ` (
	channel_id VARCHAR(64) NOT NULL,
	count_date DATE NOT NULL,
	message_count BIGINT NOT NULL DEFAULT 0,
	PRIMARY KEY (channel_id,count_date),
	FOREIGN KEY ds_messages_channel_fk (channel_id) 
		REFERENCES discord_stats_channels (id)
		ON DELETE NO ACTION)
	ENGINE = InnoDB;`
const messageCountSelectLastDate = "SELECT count_date FROM " + messageCountsTableName + " WHERE channel_id = ? ORDER BY count_date DESC LIMIT 1"
const messageCountsInsert = "INSERT INTO " + messageCountsTableName + " (channel_id, count_date, message_count) VALUES (?,?,?) ON DUPLICATE KEY UPDATE message_count = ?"

// MessageCount keeps information about the number or messages on a day in a channel in a guild
type MessageCount struct {
	ChannelID     string    `db:"channel_id"`
	Date          time.Time `db:"count_date"`
	MessagesCount uint64    `db:"message_count"`
}

func (d *Database) LastMessageCountDate(channelID string) (time.Time, error) {
	var lastCountDate time.Time
	stmt, err := d.db.Preparex(messageCountSelectLastDate)
	if err != nil {
		return lastCountDate, errors.Wrap(err, "could not prepare last message count query")
	}
	defer stmt.Close()

	// ignoring error and treating as epoch time
	stmt.Get(&lastCountDate, channelID)

	return lastCountDate, nil

}

func (d *Database) SaveMessageCount(channelID string, date time.Time, count uint64) error {
	stmt, err := d.db.Prepare(messageCountsInsert)
	if err != nil {
		return errors.Wrap(err, "could not prepare INSERT")
	}

	_, err2 := stmt.Exec(channelID, date, count, count)
	if err2 != nil {
		return errors.Wrap(err, "unable to insert message count")
	}

	return nil
}

func (d *Database) MessageCountsByGuildChannels(guildID string, maxDays int) ([]MessageCount, error) {

	messageCounts := []MessageCount{}
	var qMaxDays string
	if maxDays > 0 {
		qMaxDays = fmt.Sprintf("AND mc.count_date > NOW() - INTERVAL %d DAY", maxDays+1)
	}
	q := `SELECT mc.channel_id, mc.count_date, mc.message_count 
			FROM discord_stats_message_counts as mc 
			INNER JOIN discord_stats_channels AS c ON mc.channel_id=c.id 
			WHERE c.guild_id=? ` + qMaxDays + ` ORDER BY mc.channel_id ASC, mc.count_date ASC`

	if err := d.db.Select(&messageCounts, q, guildID); err != nil {
		return messageCounts, errors.Wrap(err, "unable to query message counts")
	}

	return messageCounts, nil
}
