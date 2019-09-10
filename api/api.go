package api

import (
	"net/http"

	"github.com/FederationOfFathers/discordstats/db"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type APIHandlers struct {
	db  *db.Database
	log *logrus.Entry
}

func (a *APIHandlers) Start() {

	a.log.Info("starting API")

	r := mux.NewRouter()
	r.HandleFunc("/message_counts/{guildID}", a.MessageCountsHandler)
	r.HandleFunc("/message_counts/{guildID}/30", a.MessageCountsHandler30)
	r.HandleFunc("/message_counts/{guildID}/60", a.MessageCountsHandler60)
	r.HandleFunc("/message_counts/{guildID}/90", a.MessageCountsHandler90)
	// r.HandleFunc("/message_counts/{guildID}/alltime", a.MessageCountsHandlerAll)
	http.ListenAndServe(":8801", r)
	a.log.Info("listneing on localhost:8801")
}

func NewAPIHandlers(d *db.Database) APIHandlers {
	l := logrus.WithField("_module", "api")
	return APIHandlers{
		db:  d,
		log: l,
	}
}
