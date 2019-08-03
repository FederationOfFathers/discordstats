package db

import "time"

const guildTableName = tablePrefix + "_guilds"
const guildSchema = `CREATE TABLE discord_stats_guilds ( id INT NOT NULL AUTO_INCREMENT , 
	guildId VARCHAR(32) NOT NULL , 
	created_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP , 
	PRIMARY KEY (id), 
	UNIQUE discord_stats_guild_uq (guildId)) ENGINE = InnoDB;`

type Guild struct {
	ID          int
	GuildID     string
	CreatedDate time.Time
}
