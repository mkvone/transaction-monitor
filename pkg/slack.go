package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type SlackWebhook struct {
	Blocks []Block `json:"blocks"`
}
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}
type Block struct {
	Type    string      `json:"type"`
	Text    *BlockText  `json:"text,omitempty"`
	Fields  []BlockText `json:"fields,omitempty"`
	Divider bool        `json:"divider,omitempty"`
	Ts      int64       `json:"ts,omitempty"`
}

type BlockText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func SendSlackWebhook(webhookURL string, alertData AlertData) error {
	var blocks []Block

	// Title Block
	titleText := fmt.Sprintf("*%s New Transaction*\n<%s%s|View on Explorer>", alertData.ChainName, alertData.ExplorerURL, alertData.TxHash)
	blocks = append(blocks, Block{
		Type: "section",
		Text: &BlockText{Type: "mrkdwn", Text: titleText},
	})
	transactionText := fmt.Sprintf("Transaction: `%s`", alertData.TxHash)
	blocks = append(blocks, Block{
		Type: "section",
		Text: &BlockText{Type: "mrkdwn", Text: transactionText},
	})

	heightText := fmt.Sprintf("Height: `%s`\nFees: `%s`\nMemo : `%s`", alertData.Height, alertData.Fees, alertData.Memo)
	blocks = append(blocks, Block{
		Type: "section",
		Text: &BlockText{Type: "mrkdwn", Text: heightText},
	})
	// Divider Block

	for _, detail := range alertData.MessageDetails {
		// Detail Header Block
		blocks = append(blocks, Block{Type: "divider"})
		headerText := fmt.Sprintf("*#%d %s*", detail.Index, detail.Action)
		blocks = append(blocks, Block{
			Type: "section",
			Text: &BlockText{Type: "mrkdwn", Text: headerText},
		})

		// Fields Block
		var fields []BlockText
		for _, d := range detail.Details {
			for k, v := range d {
				fieldText := fmt.Sprintf("*%s:*\n`%s`", k, v)
				fields = append(fields, BlockText{Type: "mrkdwn", Text: fieldText})
			}
		}
		if len(fields) > 0 {
			blocks = append(blocks, Block{
				Type:   "section",
				Fields: fields,
			})
		}

		// Divider Block after each detail
	}

	webhook := SlackWebhook{
		Blocks: blocks,
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
