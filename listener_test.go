package netfw_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-netfw/netfw"
	"github.com/go-stomp/stomp/v3"
	server "github.com/go-stomp/stomp/v3/server"
	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
)

func TestStompWebsocket(t *testing.T) {

	r := mux.NewRouter()
	testServer := httptest.NewServer(r)
	forwarder := netfw.NewListener()
	go server.Serve(forwarder)

	r.Handle("/stomp", websocket.Handler(func(c *websocket.Conn) {
		forwarder.Forward(c)
	}))

	url := strings.TrimPrefix(testServer.URL, "http://")
	url = "ws://" + url + "/stomp"

	conn, err := websocket.Dial(url, "", "http://localhost/")
	if err != nil {
		t.Fatal(err)
	}
	stompConn, err := stomp.Connect(conn)
	if err != nil {
		t.Fatal(err)
	}

	sub, err := stompConn.Subscribe("mytopic", stomp.AckAuto)
	if err != nil {
		t.Fatal(err)
	}

	msgRecv := make(chan string)
	go func() {
		msg, _ := sub.Read()
		msgRecv <- string(msg.Body)
	}()

	err = stompConn.Send("mytopic", "application/json", []byte("hello world\n"))
	if err != nil {
		t.Fatal(err)
	}

	body := <-msgRecv
	if body != "hello world\n" {
		t.Fatal(body)
	}
}
