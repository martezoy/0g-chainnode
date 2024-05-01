package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/0glabs/0g-chain/helper/da/service"
	"github.com/0glabs/0g-chain/helper/da/types"

	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

const (
	subscribeMsg = "{\"jsonrpc\":\"2.0\",\"method\":\"subscribe\",\"id\":1,\"params\":{\"query\":\"tm.event='Tx'\"}}"
)

var (
	rpcAddress   = flag.String("rpc-address", "34.214.2.28:32001", "address of da-light rpc server")
	wsAddress    = flag.String("ws-address", "127.0.0.1:26657", "address of emvos ws server")
	relativePath = flag.String("relative-path", "", "relative path of evmosd")
	account      = flag.String("account", "", "account to run evmosd cli")
	keyring      = flag.String("keyring", "", "keyring to run evmosd cli")
	homePath     = flag.String("home", "", "home path of evmosd node")
)

func newUpgrader() *websocket.Upgrader {
	u := websocket.NewUpgrader()
	u.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
		log.Println("onEcho:", string(data))
		ctx := context.WithValue(context.Background(), types.DA_RPC_ADDRESS, *rpcAddress)
		ctx = context.WithValue(ctx, types.NODE_CLI_RELATIVE_PATH, *relativePath)
		ctx = context.WithValue(ctx, types.NODE_CLI_EXEC_ACCOUNT, *account)
		ctx = context.WithValue(ctx, types.NODE_CLI_EXEC_KEYRING, *keyring)
		ctx = context.WithValue(ctx, types.NODE_HOME_PATH, *homePath)
		go func() { service.OnMessage(ctx, c, messageType, data) }()
	})

	u.OnClose(func(c *websocket.Conn, err error) {
		fmt.Println("OnClose:", c.RemoteAddr().String(), err)
		service.OnClose()
	})

	return u
}

func main() {
	flag.Parse()
	engine := nbhttp.NewEngine(nbhttp.Config{})
	err := engine.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}

	go func() {
		u := url.URL{Scheme: "ws", Host: *wsAddress, Path: "/websocket"}
		dialer := &websocket.Dialer{
			Engine:      engine,
			Upgrader:    newUpgrader(),
			DialTimeout: time.Second * 3,
		}
		c, res, err := dialer.Dial(u.String(), nil)
		if err != nil {
			if res != nil && res.Body != nil {
				bReason, _ := io.ReadAll(res.Body)
				fmt.Printf("dial failed: %v, reason: %v\n", err, string(bReason))
			} else {
				fmt.Printf("dial failed: %v\n", err)
			}
			return
		}
		c.WriteMessage(websocket.TextMessage, []byte(subscribeMsg))
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine.Shutdown(ctx)
}
