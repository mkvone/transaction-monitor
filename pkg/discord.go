package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type DiscordWebhook struct {
	Username  string  `json:"username"`
	AvatarURL string  `json:"avatar_url"`
	Embeds    []Embed `json:"embeds"`
}

type Embed struct {
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Color       int          `json:"color"`
	Fields      []EmbedField `json:"fields"`
}

type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

func SendDiscordWebhook(webhookURL string, alertData AlertData) error {
	url := fmt.Sprintf("%s%s", alertData.ExplorerURL, alertData.TxHash)

	fields := []EmbedField{}

	for _, detail := range alertData.MessageDetails {
		fields = append(fields, EmbedField{Name: "\u200B", Value: "\u200B", Inline: false})
		fields = append(fields, EmbedField{
			Name:   fmt.Sprintf("#%d %s", detail.Index, detail.Action),
			Value:  "_ _", // Empty value to just show the action and index
			Inline: false,
		})
		for _, d := range detail.Details {
			for k, v := range d {
				fields = append(fields, EmbedField{
					Name:   k,
					Value:  fmt.Sprintf("`%s`", v),
					Inline: true,
				})
			}
		}

	}

	embed := Embed{
		Title:       fmt.Sprintf("%s New Transaction (<t:%d>)", alertData.ChainName, convertToUnixTimestamp(alertData.Timestamp)),
		Description: fmt.Sprintf("[Txs Hash](%s) : *`%s`*\nHeight : `%s`\nFees : `%s`\n Memo : `%s`", url, alertData.TxHash, alertData.Height, alertData.Fees, alertData.Memo),

		Fields: fields,
		Color:  15258703, // Sample color code
	}

	webhook := DiscordWebhook{
		Username: "Transaction Bot",
		Embeds:   []Embed{embed},
	}

	jsonBytes, err := json.Marshal(webhook)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func convertToUnixTimestamp(isoTimestamp string) int64 {
	parsedTime, err := time.Parse(time.RFC3339, isoTimestamp)
	if err != nil {
		log.Printf("Error parsing time: %v", err)
		return 0
	}
	return parsedTime.Unix()
}
