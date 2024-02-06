package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type TelegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"` // "Markdown" or "HTML"
}

func SendTelegramMessage(botToken string, chatID string, alertData AlertData) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	var messageText string
	messageText += fmt.Sprintf("*%s New Transaction*\n[View on Explorer](%s%s)\n", alertData.ChainName, alertData.ExplorerURL, alertData.TxHash)
	if alertData.Error != "" {
		messageText += fmt.Sprintf("Error: ```%s```\n", alertData.Error)
	}
	messageText += fmt.Sprintf("Transaction: `%s`\n", alertData.TxHash)
	messageText += fmt.Sprintf("Height: `%s`\nFees: `%s`\nMemo: `%s`", alertData.Height, alertData.Fees, alertData.Memo)
	for _, detail := range alertData.MessageDetails {
		messageText += fmt.Sprintf("\n*#%d %s*\n", detail.Index, detail.Action)
		for _, d := range detail.Details {
			for k, v := range d {
				messageText += fmt.Sprintf("*%s:* `%s`\n", k, v)
			}
		}
	}

	telegramMessage := TelegramMessage{
		ChatID:    chatID,
		Text:      messageText,
		ParseMode: "Markdown",
	}

	jsonBytes, err := json.Marshal(telegramMessage)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
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
