package main

import (
	"errors"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/didntpot/pregdk"
	"github.com/go-jose/go-jose/v4/json"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"golang.org/x/oauth2"
)

// config ...
type config struct {
	LocalAddress  string
	RemoteAddress string
}

// The following program implements a proxy that forwards players from one local address to a remote address.
func main() {
	log := slog.Default()
	conf := config{
		LocalAddress:  "0.0.0.0:19133",
		RemoteAddress: "0.0.0.0:19132",
	}
	token := tokenSource()

	prov, err := minecraft.NewForeignStatusProvider(conf.RemoteAddress)
	if err != nil {
		panic(err)
	}

	listener, err := minecraft.ListenConfig{
		StatusProvider:    prov,
		AcceptedProtocols: single(pregdk.Protocol(true)),
	}.Listen("raknet", conf.LocalAddress)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	log.Info("started listener", "on", conf.LocalAddress, "to", conf.RemoteAddress)
	for {
		c, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go handleConn(c.(*minecraft.Conn), listener, conf, token, log)
	}
}

// handleConn handles a new incoming minecraft.Conn from the minecraft.Listener passed.
func handleConn(conn *minecraft.Conn, listener *minecraft.Listener, conf config, src oauth2.TokenSource, log *slog.Logger) {
	serverConn, err := minecraft.Dialer{
		ErrorLog:    log,
		ClientData:  conn.ClientData(),
		TokenSource: src,
	}.DialTimeout("raknet", conf.RemoteAddress, time.Minute)
	if err != nil {
		panic(err)
	}

	var g sync.WaitGroup
	g.Add(2)
	go func() {
		if err := conn.StartGame(serverConn.GameData()); err != nil {
			panic(err)
		}
		g.Done()
	}()
	go func() {
		if err := serverConn.DoSpawn(); err != nil {
			panic(err)
		}
		g.Done()
	}()
	g.Wait()

	go func() {
		defer listener.Disconnect(conn, "connection lost")
		defer serverConn.Close()
		for {
			pk, err := conn.ReadPacket()
			if err != nil {
				return
			}
			if err := serverConn.WritePacket(pk); err != nil {
				var disc minecraft.DisconnectError
				if ok := errors.As(err, &disc); ok {
					_ = listener.Disconnect(conn, disc.Error())
				}
				return
			}
		}
	}()
	go func() {
		defer serverConn.Close()
		defer listener.Disconnect(conn, "connection lost")
		for {
			pk, err := serverConn.ReadPacket()
			if err != nil {
				var disc minecraft.DisconnectError
				if ok := errors.As(err, &disc); ok {
					_ = listener.Disconnect(conn, disc.Error())
				}
				return
			}
			if err := conn.WritePacket(pk); err != nil {
				return
			}
		}
	}()
}

// tokenSource returns a token source for using with a gophertunnel client. It either reads it from the
// token.tok file if cached or requests logging in with a device code.
func tokenSource() oauth2.TokenSource {
	check := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	token := new(oauth2.Token)
	tokenData, err := os.ReadFile("token.tok")
	if err == nil {
		_ = json.Unmarshal(tokenData, token)
	} else {
		token, err = auth.RequestLiveToken()
		check(err)
	}
	src := auth.RefreshTokenSource(token)
	_, err = src.Token()
	if err != nil {
		// The cached refresh token expired and can no longer be used to obtain a new token. We require the
		// user to log in again and use that token instead.
		token, err = auth.RequestLiveToken()
		check(err)
		src = auth.RefreshTokenSource(token)
	}
	tok, _ := src.Token()
	b, _ := json.Marshal(tok)
	_ = os.WriteFile("token.tok", b, 0644)
	return src
}

// single ...
func single[Y any](x Y) []Y {
	return []Y{x}
}
