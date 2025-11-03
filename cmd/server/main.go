package main

import (
	"log/slog"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/didntpot/pregdk"
)

func main() {
	chat.Global.Subscribe(chat.StdoutSubscriber{})
	slog.SetLogLoggerLevel(slog.LevelDebug)
	log := slog.Default()

	cfg := server.DefaultConfig()
	cfg.World.Folder = "dragonfly/world"
	cfg.Players.Folder = "dragonfly/players"
	cfg.Resources.Folder = "dragonfly/resources"

	conf, err := cfg.Config(log)
	if err != nil {
		panic(err)
	}
	listenerFunc(&conf, cfg.Network.Address, single(pregdk.Protocol(false)))

	srv := conf.New()
	srv.CloseOnProgramEnd()

	srv.Listen()
	for p := range srv.Accept() {
		_ = p
	}
}
