package pkg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func UnmarshalResponse(data []byte) (Response, error) {
	var r Response
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Response) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Response struct {
	Tx         Tx         `json:"tx"`
	TxResponse TxResponse `json:"tx_response"`
}

type Tx struct {
	Body       Body     `json:"body"`
	AuthInfo   AuthInfo `json:"auth_info"`
	Signatures []string `json:"signatures"`
	Type       *string  `json:"@type,omitempty"`
}

type AuthInfo struct {
	SignerInfos []SignerInfo `json:"signer_infos"`
	Fee         Fee          `json:"fee"`
}

type Fee struct {
	Amount   []Amount `json:"amount"`
	GasLimit string   `json:"gas_limit"`
	Payer    string   `json:"payer"`
	Granter  string   `json:"granter"`
}

type Amount struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type SignerInfo struct {
	PublicKey PublicKey `json:"public_key"`
	ModeInfo  ModeInfo  `json:"mode_info"`
	Sequence  string    `json:"sequence"`
}

type ModeInfo struct {
	Single Single `json:"single"`
}

type Single struct {
	Mode string `json:"mode"`
}

type PublicKey struct {
	Type string `json:"@type"`
	Key  string `json:"key"`
}

type Body struct {
	Messages                    []Message     `json:"messages"`
	Memo                        string        `json:"memo"`
	TimeoutHeight               string        `json:"timeout_height"`
	ExtensionOptions            []interface{} `json:"extension_options"`
	NonCriticalExtensionOptions []interface{} `json:"non_critical_extension_options"`
}

type Message struct {
	Type               string      `json:"@type"`
	DelegatorAddress   *string     `json:"delegator_address,omitempty"`
	ValidatorAddress   *string     `json:"validator_address,omitempty"`
	FromAddress        *string     `json:"from_address,omitempty"`
	ToAddress          *string     `json:"to_address,omitempty"`
	Amount             interface{} `json:"amount,omitempty"`
	ProposalId         *string     `json:"proposal_id,omitempty"`
	Voter              *string     `json:"voter,omitempty"`
	Option             *string     `json:"option,omitempty"`
	SourcePort         *string     `json:"source_port,omitempty"`
	SourceChannel      *string     `json:"source_channel,omitempty"`
	Token              *Token      `json:"token,omitempty"`
	Sender             *string     `json:"sender,omitempty"`
	Receiver           *string     `json:"receiver,omitempty"`
	TimeoutHeight      *Height     `json:"timeout_height,omitempty"`
	TimeoutTimestamp   *string     `json:"timeout_timestamp,omitempty"`
	ClientID           *string     `json:"client_id,omitempty"`
	Signer             *string     `json:"signer,omitempty"`
	PacketSequence     *string     `json:"sequence,omitempty"`
	DestinationPort    *string     `json:"destination_port,omitempty"`
	DestinationChannel *string     `json:"destination_channel,omitempty"`
	// Data               *string     `json:"data,omitempty"`
	Data   *json.RawMessage `json:"data,omitempty"`
	Packet *PacketData      `json:"packet,omitempty"`
}
type PacketData struct {
	PacketSequence     *string `json:"sequence,omitempty"`
	SourcePort         *string `json:"source_port,omitempty"`
	SourceChannel      *string `json:"source_channel,omitempty"`
	DestinationPort    *string `json:"destination_port,omitempty"`
	DestinationChannel *string `json:"destination_channel,omitempty"`
	Data               *string `json:"data,omitempty"`
	TimeoutHeight      *Height `json:"timeout_height,omitempty"`
	TimeoutTimestamp   *string `json:"timeout_timestamp,omitempty"`
	Signer             *string `json:"signer,omitempty"`
}
type Token struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type Height struct {
	RevisionNumber string `json:"revision_number"`
	RevisionHeight string `json:"revision_height"`
}

type MsgVote struct {
	ProposalId string `json:"proposal_id"`
	Voter      string `json:"voter"`
	Option     string `json:"option"`
}
type TxResponse struct {
	Height    string  `json:"height"`
	Txhash    string  `json:"txhash"`
	Codespace string  `json:"codespace"`
	Code      int64   `json:"code"`
	Data      string  `json:"data"`
	Logs      []Log   `json:"logs"`
	Events    []Event `json:"events"`
	ErrorLog  string  `json:"raw_log"`
	Info      string  `json:"info"`
	GasWanted string  `json:"gas_wanted"`
	GasUsed   string  `json:"gas_used"`
	Tx        Tx      `json:"tx"`
	Timestamp string  `json:"timestamp"`
}

type Log struct {
	MsgIndex int64 `json:"msg_index,omitempty"`
	// Log      string  `json:"log,omitempty"`
	Events []Event `json:"events,omitempty"`
}

func (m *Message) UnmarshalJSON(data []byte) error {
	type Alias Message
	aux := &struct {
		Amount json.RawMessage `json:"amount,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch m.Type {
	case "/cosmos.bank.v1beta1.MsgSend":
		var amounts []Amount
		if err := json.Unmarshal(aux.Amount, &amounts); err != nil {
			return err
		}
		m.Amount = amounts

	case "/cosmos.staking.v1beta1.MsgDelegate":
		var amount Amount
		if err := json.Unmarshal(aux.Amount, &amount); err != nil {
			return err
		}
		m.Amount = amount
		// case "/ibc.applications.transfer.v1.MsgTransfer":
		// 	// IBC 전송 메시지에 대한 처리
		// 	var token Token
		// 	if err := json.Unmarshal(aux.Amount, &token); err != nil {
		// 		return err
		// 	}
		// 	m.Amount = token

	}

	return nil
}

func buildAPIURL(baseURL, tx string) string {
	return fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", baseURL, tx)
}

func fetchAPIData(url string) (*Response, error) {
	log.Print(url)
	resp, err := http.Get(url)

	if err != nil {
		log.Printf("Error making HTTP request: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		log.Printf("Received non-200 status code: %d\n", resp.StatusCode)
		return nil, fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v\n", err)
		return nil, err
	}

	var apiData Response
	if err := json.Unmarshal(body, &apiData); err != nil {
		log.Printf("Error unmarshalling response: %v\n", err)
		return nil, err
	}

	return &apiData, nil
}

type AlertData struct {
	TxHash         string
	Height         string
	Timestamp      string
	ChainName      string
	ExplorerURL    string
	MessageDetails []MessageDetail
	Fees           string
	Memo           string
	Error          string
}
type MessageDetail struct {
	Index   int
	Action  string
	Details []map[string]string
}

func appendIfNotNil(details *[]map[string]string, key string, value *string) {
	if value != nil {
		*details = append(*details, map[string]string{key: *value})
	}
}

func transformData(apiData *Response, alerts *AlertData) {
	if apiData == nil {
		log.Println("apiData is nil")
		return
	}
	if apiData.TxResponse.Code != 0 {
		alerts.Error = apiData.TxResponse.ErrorLog
	}

	// Common data extraction
	alerts.Timestamp = apiData.TxResponse.Timestamp
	alerts.Height = apiData.TxResponse.Height
	alerts.TxHash = apiData.TxResponse.Txhash
	alerts.Memo = apiData.Tx.Body.Memo

	// Extract fees
	if len(apiData.Tx.AuthInfo.Fee.Amount) > 0 {
		amount := extractNumber(apiData.Tx.AuthInfo.Fee.Amount[0].Amount) / 1000000
		denom := extractDenom(apiData.Tx.AuthInfo.Fee.Amount[0].Denom)
		alerts.Fees = fmt.Sprintf("%f %s", amount, denom)
	} else {
		alerts.Fees = "0"
	}

	// Message processing
	for i, message := range apiData.Tx.Body.Messages {
		// var messageDetail MessageDetail
		// messageDetail.Index = i + 1
		messageDetail := MessageDetail{
			Index:   i + 1,
			Details: make([]map[string]string, 0),
		}

		// Depending on whether logs or events are available, choose the appropriate function
		var amount float64
		var denom string
		eventType := getEventType(message.Type)

		if len(apiData.TxResponse.Logs) > 0 {
			amount, denom = extractAmountFromLogs(apiData.TxResponse.Logs, i, eventType)
		} else {
			amount, denom = extractAmountFromEvents(apiData.TxResponse.Events, eventType)
		}

		// Populate message details based on the type
		messageDetail.Action = getMessageAction(message.Type)
		populateMessageDetails(&messageDetail, message, amount, denom)
		alerts.MessageDetails = append(alerts.MessageDetails, messageDetail)
	}
}

func getEventType(messageType string) string {
	switch messageType {
	case "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward":
		return "withdraw_rewards"
	case "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission":
		return "withdraw_commission"
	case "/cosmos.staking.v1beta1.MsgDelegate":
		return "delegate"
	default:
		return ""
	}
}

func getMessageAction(messageType string) string {
	switch messageType {
	case "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward":
		return "Get Reward"
	case "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission":
		return "Get Commission"
	case "/cosmos.staking.v1beta1.MsgDelegate":
		return "Delegate"
	case "/ibc.applications.transfer.v1.MsgTransfer":
		return "IBC Transfer"
	case "/cosmos.gov.v1beta1.MsgVote":
		return "Vote"
	case "/cosmos.bank.v1beta1.MsgSend":
		return "Send"
	case "/ibc.core.client.v1.MsgUpdateClient":
		return "IBC Update Client"
	case "/ibc.core.channel.v1.MsgRecvPacket":
		return "IBC Received"
	case "/ibc.core.channel.v1.MsgAcknowledgement":
		return "IBC Acknowledgement"
	default:
		return messageType
	}
}

func populateMessageDetails(details *MessageDetail, message Message, amount float64, denom string) {
	appendIfNotNil(&details.Details, "Delegator Address", message.DelegatorAddress)
	appendIfNotNil(&details.Details, "Validator Address", message.ValidatorAddress)
	appendIfNotNil(&details.Details, "From Address", message.FromAddress)
	appendIfNotNil(&details.Details, "To Address", message.ToAddress)
	appendIfNotNil(&details.Details, "Sender", message.Sender)
	appendIfNotNil(&details.Details, "Receiver", message.Receiver)
	appendIfNotNil(&details.Details, "Source Channel", message.SourceChannel)
	appendIfNotNil(&details.Details, "Port", message.SourcePort)
	appendIfNotNil(&details.Details, "Proposal Id", message.ProposalId)
	appendIfNotNil(&details.Details, "Voter", message.Voter)
	appendIfNotNil(&details.Details, "Option", message.Option)
	appendIfNotNil(&details.Details, "Signer", message.Signer)
	appendIfNotNil(&details.Details, "Client ID", message.ClientID)
	appendIfNotNil(&details.Details, "Source Port", message.SourcePort)
	appendIfNotNil(&details.Details, "Timeout Timestamp", message.TimeoutTimestamp)
	appendIfNotNil(&details.Details, "Sequence", message.PacketSequence)
	appendIfNotNil(&details.Details, "Destination Port", message.DestinationPort)
	if amount != 0 {
		Amount := fmt.Sprintf("%f %s", amount, denom)
		appendIfNotNil(&details.Details, "Amount", &Amount)
	}

	if packet := message.Packet; packet != nil {
		appendIfNotNil(&details.Details, "Sequence", packet.PacketSequence)
		appendIfNotNil(&details.Details, "Source Port", packet.SourcePort)
		appendIfNotNil(&details.Details, "Source Channel", packet.SourceChannel)
		appendIfNotNil(&details.Details, "Destination Port", packet.DestinationPort)
		appendIfNotNil(&details.Details, "Destination Channel", packet.DestinationChannel)
		// Decode any packet data if available
		if packet.Data != nil {
			var packetData map[string]interface{}
			decodedData, err := base64.StdEncoding.DecodeString(*packet.Data)
			if err == nil {
				json.Unmarshal(decodedData, &packetData)
				for key, value := range packetData {
					valueStr := fmt.Sprintf("%v", value) // Convert interface{} to string
					appendIfNotNil(&details.Details, key, &valueStr)
				}
			}
		}
	}
}

func extractNumber(str string) float64 {
	reg, err := regexp.Compile("^[0-9]+")
	if err != nil {
		log.Println(err)
	}
	// Find the first match which is a continuous sequence of digits at the beginning
	match := reg.FindString(str)
	if match == "" {
		log.Println("No leading number found in string:", str)
		return 0
	}

	number, err := strconv.ParseFloat(match, 64)
	if err != nil {
		log.Println(str)
		log.Println("Error converting to integer:", err)
		return 0
	}
	return number
}

// Extracts the alphabetic part from the string.
func extractDenom(str string) string {
	reg, err := regexp.Compile("[^a-zA-Z]+")
	if err != nil {
		log.Println(err)
	}
	denom := reg.ReplaceAllString(str, "")

	// Remove 'u' prefix if it exists
	if strings.HasPrefix(denom, "u") {
		denom = strings.TrimPrefix(denom, "u")
	}

	return denom
}
func extractAmountFromEvents(events []Event, eventType string) (float64, string) {
	if len(events) == 0 {
		log.Println("No events found")
		return 0, ""
	}

	for _, event := range events {
		if event.Type == eventType {
			for _, attr := range event.Attributes {
				if attr.Key == "amount" {
					amount := extractNumber(attr.Value) / 1000000
					denom := extractDenom(attr.Value)
					return amount, denom
				}
			}
		}
	}
	return 0, ""
}
func extractAmountFromLogs(logs []Log, msgIndex int, eventType string) (float64, string) {
	if len(logs) == 0 {
		log.Println("No logs found")
		return 0, ""
	}

	for _, log := range logs {
		if log.MsgIndex == int64(msgIndex) {
			for _, event := range log.Events {
				if event.Type == eventType {
					for _, attr := range event.Attributes {
						if attr.Key == "amount" {
							amount := extractNumber(attr.Value) / 1000000
							denom := extractDenom(attr.Value)
							return amount, denom
						}
					}
				}
			}
		}
	}
	return 0, ""
}
