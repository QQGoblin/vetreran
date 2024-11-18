package main

import (
	"flag"
	"github.com/QQGoblin/veteran/pkg"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
)

func main() {

	config := flag.String("config", "veteran.json", "config file path")
	flag.Parse()

	vetreranConfig, err := pkg.LoadConfig(*config)
	if err != nil {
		log.WithError(err).Error("load config failure")
		os.Exit(-1)
	}

	log.WithFields(log.Fields{
		"ID":   vetreranConfig.ID,
		"Bind": vetreranConfig.Bind,
		"Peer": vetreranConfig.Peers,
	}).Infof("read configure from %s", *config)

	vetreran, err := pkg.NewVeteran(vetreranConfig)
	if err != nil {
		os.Exit(-1)
	}

	if error := vetreran.Start(); error != nil {
		log.WithFields(log.Fields{"error": error}).Error("Start failed")
		os.Exit(-1)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	<-signalChan
	vetreran.Stop()

}
