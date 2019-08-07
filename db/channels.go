package db

import (
	"time"

	"github.com/pkg/errors"
)

const channelsTablename = tablePrefix + "_channels"
const channelsSchema = `CREATE TABLE IF NOT EXISTS ` + channelsTablename + ` (
	id VARCHAR(64) NOT NULL PRIMARY KEY,
	name VARCHAR(256) NULL,
	guild_id VARCHAR(64) NOT NULL,
	last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	last_message_seen TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY ds_channel_guild_fk (guild_id)
		REFERENCES discord_stats_guilds (id) 
		ON DELETE CASCADE)
	ENGINE = InnoDB;`
const insertChannelStmt = `INSERT INTO ` + channelsTablename + ` (id, name, guild_id, last_updated, last_message_seen) 
							VALUES (?, ?, ?, NOW(), ?) 
							ON DUPLICATE KEY UPDATE 
								name = ?,
								last_updated = NOW(), 
								last_message_seen = GREATEST(?, last_message_seen)`

type Channel struct {
	ID              string
	Name            string
	GuildID         string
	LastUpdated     time.Time
	LastMessageSeen time.Time
}

func (d *Database) SaveChannel(chID string, chName string, gID string, lastMessageSeen time.Time) error {

	stmt, err := d.db.Prepare(insertChannelStmt)
	if err != nil {
		return errors.Wrap(err, "unable to prepare channel insert statement")
	}

	_, err2 := stmt.Exec(chID, chName, gID, lastMessageSeen, chName, lastMessageSeen)
	return err2

}
