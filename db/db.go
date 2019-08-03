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

func Connect(cfg DBConfig) *sqlx.DB {
	return sqlx.MustConnect("mysql", cfg.ConnectionString)
}

// Initialize initializes a database by ensuring the tables needed exist
func Initialize(d *sqlx.DB) {
	d.MustExec(guildSchema)
}
