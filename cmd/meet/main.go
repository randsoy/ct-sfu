package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	log "github.com/pion/ion-log"
	"github.com/randsoy/ct-sfu/internal/meet"
	"github.com/randsoy/ct-sfu/internal/meet/conf"
	"github.com/randsoy/ct-sfu/internal/meet/jrpc"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	fixByFile := []string{"asm_amd64.s", "proc.go", "icegatherer.go"}
	fixByFunc := []string{}
	log.Init(conf.Conf.Log.Level, fixByFile, fixByFunc)
	log.Infof("--- starting ct-sfu node ---")

	m := meet.New(conf.Conf)
	rpcs := jrpc.New(conf.Conf, m)

	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("ct-sfu get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			m.Close()
			rpcs.Close()
			log.Infof("ct-sfu [version: %s] exit", "1.0")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
