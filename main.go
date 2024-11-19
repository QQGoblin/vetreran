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

	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	vetreranConfig, err := pkg.LoadConfig(*config)
	if err != nil {
		log.WithError(err).Error("Load config failure")
		os.Exit(-1)
	}

	vetreran, err := pkg.NewVeteran(vetreranConfig)
	if err != nil {
		os.Exit(-1)
	}

	if err = vetreran.Start(); err != nil {
		log.WithError(err).Error("Start failed")
		os.Exit(-1)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	<-signalChan
	vetreran.Stop()

}
