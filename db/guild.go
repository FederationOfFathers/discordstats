package db

import (
	"time"

	"github.com/pkg/errors"
)

const guildTableName = tablePrefix + "_guilds"
const guildSchema = `CREATE TABLE IF NOT EXISTS discord_stats_guilds (
	id VARCHAR(64) NOT NULL PRIMARY KEY, 
	last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP )
	ENGINE = InnoDB;`

type Guild struct {
	ID          string    `db:"id"`
	LastUpdated time.Time `db:"last_updated"`
}

// InsertOrUpdateGuild adds the guildID or updated the last_updated field if it already exists
func (d *Database) InsertOrUpdateGuild(guildID string) error {

	stmt, err := d.db.Prepare("INSERT INTO discord_stats_guilds (id, last_updated) VALUES(?, NOW()) ON DUPLICATE KEY UPDATE last_updated = NOW()")
	if err != nil {
		return errors.Wrap(err, "unable to prepare stmt")
	}

	_, err2 := stmt.Exec(guildID)

	return err2
}
