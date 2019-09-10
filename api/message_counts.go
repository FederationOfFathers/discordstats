package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type MessageCountResponse struct {
	ChannelCounts map[string][]MessageCount
}

type MessageCount struct {
	Date  string
	Count uint64
}

// MessageCountsHandler returns the count of messages for channels in a guild
func (a *APIHandlers) MessageCountsHandler(w http.ResponseWriter, r *http.Request) {
	a.getMessageCounts(w, r, 1)
}

func (a *APIHandlers) MessageCountsHandler30(w http.ResponseWriter, r *http.Request) {
	a.getMessageCounts(w, r, 30)
}

func (a *APIHandlers) MessageCountsHandler60(w http.ResponseWriter, r *http.Request) {
	a.getMessageCounts(w, r, 60)
}

func (a *APIHandlers) MessageCountsHandler90(w http.ResponseWriter, r *http.Request) {
	a.getMessageCounts(w, r, 90)
}

func (a *APIHandlers) getMessageCounts(w http.ResponseWriter, r *http.Request, maxDays int) {
	vars := mux.Vars(r)
	guildID := vars["guildID"]
	messageCounts, err := a.db.MessageCountsByGuildChannels(guildID, maxDays)
	if err != nil {
		a.log.WithFields(logrus.Fields{
			"error": err,
		}).Error("unable to get message counts")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "500 - unable to get message counts")
	}

	mcr := make(map[string][]MessageCount)

	for _, m := range messageCounts {
		mc := MessageCount{
			Date:  m.Date.Format("1/2/2006"),
			Count: m.MessagesCount,
		}
		if _, ok := mcr[m.ChannelID]; !ok {
			mcr[m.ChannelID] = []MessageCount{}
		}

		mcr[m.ChannelID] = append(mcr[m.ChannelID], mc)
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mcr)
}
