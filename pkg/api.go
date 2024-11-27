package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (v *Veteran) apiServer() *http.Server {
	r := mux.NewRouter()

	r.Methods(http.MethodGet).Path("/status").HandlerFunc(v.StatusHandler)
	r.Methods(http.MethodPost).Path("/member/{memberID}").HandlerFunc(v.AddMemberHandler)
	r.Methods(http.MethodDelete).Path("/member/{memberID}").HandlerFunc(v.DelMemberHandler)

	return &http.Server{Addr: v.config.Listen, Handler: r}

}

func (v *Veteran) StatusHandler(w http.ResponseWriter, _ *http.Request) {

	state, err := v.core.Status()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithError(err).Error("Get cluster status failure")
		return
	}

	body, err := json.Marshal(state)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithError(err).Error("Marshal raft status failure")
		return
	}

	var prettyJSON bytes.Buffer
	if err = json.Indent(&prettyJSON, body, "", "    "); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithError(err).Error("Format output failure")
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%v\n", prettyJSON.String())

}

func (v *Veteran) AddMemberHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	params := r.URL.Query()

	id := vars["memberID"]
	address := params["address"]

	if len(address) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error("Member address is not found")
		return
	}

	if err := v.core.AddMember(id, address[0]); err != nil {
		log.WithError(err).WithFields(log.Fields{"id": id, "address": address[0]}).Error("Add member failure")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.WithFields(log.Fields{"id": id, "address": address[0]}).Info("Add member success")
	w.WriteHeader(http.StatusOK)
}

func (v *Veteran) DelMemberHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["memberID"]

	if err := v.core.DelMember(id); err != nil {
		log.WithError(err).WithField("id", id).Error("Del member failure")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.WithField("id", id).Info("Del member success")
	w.WriteHeader(http.StatusOK)
}
