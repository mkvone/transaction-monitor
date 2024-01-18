package pkg

import (
	"encoding/json"
	"fmt"
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
	Type             string      `json:"@type"`
	DelegatorAddress *string     `json:"delegator_address,omitempty"`
	ValidatorAddress string      `json:"validator_address"`
	FromAddress      *string     `json:"from_address,omitempty"`
	ToAddress        *string     `json:"to_address,omitempty"`
	Amount           interface{} `json:"amount,omitempty"`
	ProposalId       *string     `json:"proposal_id,omitempty"`
	Voter            *string     `json:"voter,omitempty"`
	Option           *string     `json:"option,omitempty"`
	SourcePort       *string     `json:"source_port,omitempty"`
	SourceChannel    *string     `json:"source_channel,omitempty"`
	Token            *Token      `json:"token,omitempty"`
	Sender           *string     `json:"sender,omitempty"`
	Receiver         *string     `json:"receiver,omitempty"`
	TimeoutHeight    *Height     `json:"timeout_height,omitempty"`
	TimeoutTimestamp *string     `json:"timeout_timestamp,omitempty"`
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
	Height    string `json:"height"`
	Txhash    string `json:"txhash"`
	Codespace string `json:"codespace"`
	Code      int64  `json:"code"`
	Data      string `json:"data"`
	RawLog    string `json:"raw_log"`
	Logs      []Log  `json:"logs"`
	Info      string `json:"info"`
	GasWanted string `json:"gas_wanted"`
	GasUsed   string `json:"gas_used"`
	Tx        Tx     `json:"tx"`
	Timestamp string `json:"timestamp"`
}

type Log struct {
	MsgIndex int64   `json:"msg_index"`
	Log      string  `json:"log"`
	Events   []Event `json:"events"`
}

func (m *Message) UnmarshalJSON(data []byte) error {
	// Create a struct that mirrors Message but with Amount as a raw message.
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

	// Based on the message type, unmarshal the Amount field appropriately.
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
	}

	return nil
}

func buildAPIURL(baseURL, tx string) string {
	return fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", baseURL, tx)
}
func fetchAPIData(url string) (*Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err

	}
	defer resp.Body.Close()

	var apiData Response
	if err := json.NewDecoder(resp.Body).Decode(&apiData); err != nil {
		return nil, err
	}

	return &apiData, nil
}

type AlertData struct {
	// Common fields that are relevant to all platforms
	TxHash         string
	Height         string
	Timestamp      string
	ChainName      string
	ExplorerURL    string
	MessageDetails []MessageDetail
	Fees           string

	// Add other fields as necessary
}
type MessageDetail struct {
	Index   int
	Action  string
	Details []map[string]string
}

func transformData(apiData *Response, alerts *AlertData) {

	if apiData == nil {
		log.Println("apiData is nil")
		return
	}

	alerts.Timestamp = apiData.TxResponse.Timestamp
	alerts.Height = apiData.TxResponse.Height
	alerts.TxHash = apiData.TxResponse.Txhash
	alerts.Fees = fmt.Sprintf("%f %s ", extractNumber(apiData.Tx.AuthInfo.Fee.Amount[0].Amount)/1000000, extractDenom(apiData.Tx.AuthInfo.Fee.Amount[0].Denom))

	for i, message := range apiData.Tx.Body.Messages {
		var messageDetail MessageDetail
		messageDetail.Index = i + 1
		messageDetail.Details = make([]map[string]string, 0)

		switch message.Type {
		case "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward":
			if message.DelegatorAddress != nil {
				amount, denom := extractAmount(apiData.TxResponse.Logs, i)
				messageDetail.Action = "Get Reward"
				messageDetail.Details = append(messageDetail.Details, map[string]string{"Delegator Address": *message.DelegatorAddress})
				messageDetail.Details = append(messageDetail.Details, map[string]string{"Validator Address": message.ValidatorAddress})
				messageDetail.Details = append(messageDetail.Details, map[string]string{"Amount": fmt.Sprintf("%f %s", amount, denom)})
			}
		case "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission":
			amount, denom := extractAmount(apiData.TxResponse.Logs, i)
			messageDetail.Action = "Get Commission"
			messageDetail.Details = append(messageDetail.Details, map[string]string{"Validator Address": message.ValidatorAddress})
			messageDetail.Details = append(messageDetail.Details, map[string]string{"Amount": fmt.Sprintf("%f %s", amount, denom)})

		case "/cosmos.staking.v1beta1.MsgDelegate":
			if delegatorAddr := message.DelegatorAddress; delegatorAddr != nil {
				switch amount := message.Amount.(type) {
				case Amount:
					extractedAmount, denom := extractNumber(amount.Amount)/1000000, extractDenom(amount.Denom)
					messageDetail.Action = "Delegate"
					messageDetail.Details = append(messageDetail.Details, map[string]string{"Delegator Address": *delegatorAddr})
					messageDetail.Details = append(messageDetail.Details, map[string]string{"Validator Address": message.ValidatorAddress})
					messageDetail.Details = append(messageDetail.Details, map[string]string{"Amount": fmt.Sprintf("%f %s", extractedAmount, denom)})
				}
			}

		case "/ibc.applications.transfer.v1.MsgTransfer":
			if message.Sender != nil && message.Receiver != nil && message.Token != nil {
				var amount float64
				amount, err := strconv.ParseFloat(message.Token.Amount, 64)
				if err == nil {
					amount = amount / 1000000
				}
				messageDetail.Action = "IBC Transfer"
				messageDetail.Details = append(messageDetail.Details, map[string]string{"Sender": *message.Sender})
				messageDetail.Details = append(messageDetail.Details, map[string]string{"Receiver": *message.Receiver})
				messageDetail.Details = append(messageDetail.Details, map[string]string{"Source Channel": *message.SourceChannel})
				messageDetail.Details = append(messageDetail.Details, map[string]string{"Port": *message.SourcePort})
				messageDetail.Details = append(messageDetail.Details, map[string]string{"Amount": fmt.Sprintf("%f %s", amount, message.Token.Denom)})
			}

		case "/cosmos.gov.v1beta1.MsgVote":
			if message.ProposalId != nil && message.Voter != nil && message.Option != nil {
				messageDetail.Action = "Vote"
				messageDetail.Details = append(messageDetail.Details, map[string]string{"Proposal Id": *message.ProposalId})
				messageDetail.Details = append(messageDetail.Details, map[string]string{"Voter": *message.Voter})
				messageDetail.Details = append(messageDetail.Details, map[string]string{"Option": *message.Option})
			}

		case "/cosmos.bank.v1beta1.MsgSend":
			if fromAddr, toAddr := message.FromAddress, message.ToAddress; fromAddr != nil && toAddr != nil {
				amountSlice, ok := message.Amount.([]Amount)
				if ok && len(amountSlice) > 0 {
					extractedAmount, denom := extractNumber(amountSlice[0].Amount)/1000000, extractDenom(amountSlice[0].Denom)
					messageDetail.Action = "Send"
					messageDetail.Details = append(messageDetail.Details, map[string]string{"From": *fromAddr})
					messageDetail.Details = append(messageDetail.Details, map[string]string{"To": *toAddr})
					messageDetail.Details = append(messageDetail.Details, map[string]string{"Amount": fmt.Sprintf("%f %s", extractedAmount, denom)})
				}
			}

		}
		if len(messageDetail.Details) > 0 {
			// Append only if messageDetail contains relevant information
			alerts.MessageDetails = append(alerts.MessageDetails, messageDetail)
		}

	}

}

func extractNumber(str string) float64 {
	reg, err := regexp.Compile("^[0-9]+")
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}
	denom := reg.ReplaceAllString(str, "")

	// Remove 'u' prefix if it exists
	if strings.HasPrefix(denom, "u") {
		denom = strings.TrimPrefix(denom, "u")
	}

	return denom
}
func extractAmount(logs []Log, msgIndex int) (float64, string) {
	if len(logs) == 0 {
		log.Println("No logs found")
		return 0, ""
	}

	for _, log := range logs {
		if log.MsgIndex == int64(msgIndex) {
			for _, event := range log.Events {
				if event.Type == "withdraw_commission" || event.Type == "withdraw_rewards" {
					for _, attr := range event.Attributes {
						if attr.Key == "amount" {
							// "354817ungm" 같은 형식의 문자열에서 숫자 부분만 추출
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
