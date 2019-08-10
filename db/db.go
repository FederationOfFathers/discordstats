package db

import (
	// import mysql drivers
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	tablePrefix = "discord_stats"
)

type DBConfig struct {
	ConnectionString string `split_words:"true";default:"fofgaming:fofdev@tcp(localhost:3306)/fofgaming?charset=utf8mb4_general_ci&parseTime=true"`
}

type Database struct {
	db *sqlx.DB
}

func Connect(cfg DBConfig) Database {
	db := sqlx.MustConnect("mysql", cfg.ConnectionString)

	return Database{
		db: db,
	}
}

func (d Database) Close() {
	d.db.Close()
}

// Initialize initializes a database by ensuring the tables needed exist
func (d Database) Initialize() {
	d.db.MustExec(guildSchema)
	d.db.MustExec(channelsSchema)
	d.db.MustExec(messageCountsSchema)
}
