package pkg

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sacOO7/gowebsocket"
)

func TransformToWebSocketURL(rpcURL string) string {
	return strings.Replace(rpcURL, "https://", "wss://", 1) + "/websocket"
}

type WebSocketMessage struct {
	Jsonrpc string                 `json:"jsonrpc"`
	ID      int64                  `json:"id"`
	Result  WebSocketMessageResult `json:"result"`
}

type WebSocketMessageResult struct {
	Query  string              `json:"query"`
	Data   Data                `json:"data"`
	Events map[string][]string `json:"events"`
}

type Data struct {
	Type  string `json:"type"`
	Value Value  `json:"value"`
}

type Value struct {
	TxResult TxResult `json:"TxResult"`
}

type TxResult struct {
	Height string         `json:"height"`
	Tx     string         `json:"tx"`
	Result TxResultResult `json:"result"`
}

type TxResultResult struct {
	Data      string  `json:"data"`
	Log       string  `json:"log"`
	GasWanted string  `json:"gas_wanted"`
	GasUsed   string  `json:"gas_used"`
	Events    []Event `json:"events"`
}

type Event struct {
	Type       string      `json:"type"`
	Attributes []Attribute `json:"attributes"`
}

type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	// Index bool   `json:"index"`
}

// , updateFunc func(chainName string, height string, timestamp string)
func SubscribeToNewBlocks(cfg *Config, chain ChainConfig, chainName string, address string) {
	wsURL := TransformToWebSocketURL(chain.RPC)
	socket := gowebsocket.New(wsURL)
	log.Printf("Attempting to connect to WebSocket for chain: %s, address: %s", chainName, address)

	reconnectFunc := func() {
		time.Sleep(1 * time.Minute) // 5초 후 재연결 시도
		socket.Connect()
		log.Println("supscribe to : ", address)
		socket.SendText(fmt.Sprintf("{ \"jsonrpc\": \"2.0\", \"method\": \"subscribe\", \"params\": [\"transfer.sender ='%s'\"], \"id\": 1 }", address))

		// socket.SendText("{ \"jsonrpc\": \"2.0\", \"method\": \"subscribe\", \"params\": [\"tm.event = 'Tx'\"], \"id\": 1 }")

	}
	socket.OnConnected = func(socket gowebsocket.Socket) {
		log.Println("supscribe to : ", address)
		socket.SendText(fmt.Sprintf("{ \"jsonrpc\": \"2.0\", \"method\": \"subscribe\", \"params\": [\"transfer.sender ='%s'\"], \"id\": 1 }", address))

		// socket.SendText("{ \"jsonrpc\": \"2.0\", \"method\": \"subscribe\", \"params\": [\"tm.event = 'Tx'\"], \"id\": 1 }")

	}

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		txhash, err := extractDataFromMessage(message)
		// log.Println(message)

		if err != nil {
			log.Printf("Error parsing message from WebSocket: %v", err)
			return
		}
		if txhash != "" {
			alertChan <- NewAlert(chainName, txhash)
		}
	}
	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Print(fmt.Sprintln("WebSocket disconnected: ", err, ". Reconnecting...", wsURL))
		reconnectFunc()
	}
	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Print(fmt.Sprintln("Received connect error ", err, "\t : ", wsURL))
		reconnectFunc()
	}

	socket.Connect()
}

func extractDataFromMessage(message string) (string, error) {
	var msg WebSocketMessage
	err := json.Unmarshal([]byte(message), &msg)
	if err != nil {
		return "", err
	}
	var txhash string
	for k, v := range msg.Result.Events {
		if k == "tx.hash" {
			txhash = v[0]
		}

	}

	return txhash, nil
}
