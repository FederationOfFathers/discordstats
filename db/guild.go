package db

import (
	"time"

	"github.com/pkg/errors"
)

const guildTableName = tablePrefix + "_guilds"
const guildSchema = `CREATE TABLE IF NOT EXISTS discord_stats_guilds (
	id VARCHAR(64) NOT NULL PRIMARY KEY, 
	name VARCHAR(256) NULL,
	last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP )
	ENGINE = InnoDB;`
const insertGuildStmt = `INSERT INTO discord_stats_guilds (id, name, last_updated) 
							VALUES(?, ?, NOW()) 
							ON DUPLICATE KEY UPDATE 
								name = ?,
								last_updated = NOW()`

type Guild struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	LastUpdated time.Time `db:"last_updated"`
}

// InsertOrUpdateGuild adds the guildID or updated the last_updated field if it already exists
func (d *Database) InsertOrUpdateGuild(guildID, guildName string) error {

	stmt, err := d.db.Prepare(insertGuildStmt)
	if err != nil {
		return errors.Wrap(err, "unable to prepare stmt")
	}

	_, err2 := stmt.Exec(guildID, guildName, guildName)

	return err2
}

// LatestGuilds gets all guils that have been updated within the last 1 hour
func (d *Database) LatestGuilds() ([]Guild, error) {
	var guilds []Guild
	err := d.db.Select(&guilds, "SELECT * FROM "+guildTableName+" WHERE last_updated > NOW() - INTERVAL 60 MINUTE ORDER BY last_updated DESC")

	return guilds, errors.Wrap(err, "unable to get latest guilds")
}
