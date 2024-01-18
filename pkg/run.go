package pkg

import (
	"fmt"
	"log"
	"time"
)

var alertChan = make(chan Alert) // Buffer size can be adjusted based on expected load
type Alert struct {
	ChainName string
	TxHash    string
}

func NewAlert(chainName, txHash string) Alert {
	return Alert{
		ChainName: chainName,
		TxHash:    txHash,
	}
}
func ProcessAlerts(cfg *Config, alertChan <-chan Alert) {
	for alert := range alertChan {
		AlertRun(cfg, alert.ChainName, alert.TxHash)
	}
}

func AlertRun(cfg *Config, chainName string, txhash string) {
	url := buildAPIURL(cfg.Chains[chainName].API, txhash)
	fmt.Println(url)
	apiData, err := fetchAPIData(url)
	if err != nil {
		log.Printf("Error fetching API data: %v", err)
		return
	}
	var alerts AlertData

	transformData(apiData, &alerts)
	alerts.ChainName = chainName
	alerts.ExplorerURL = cfg.Chains[chainName].Explorer

	if cfg.Alerting.Discord.Enable {
		if err := SendDiscordWebhook(cfg.Alerting.Discord.WebhookURL, alerts); err != nil {
			log.Printf("Error sending message to Discord: %v", err)
		} else {
			log.Println("Message sent to Discord successfully")
		}
	}

	if cfg.Alerting.Slack.Enable {
		if err := SendSlackWebhook(cfg.Alerting.Slack.WebhookURL, alerts); err != nil {
			log.Printf("Error sending message to Slack: %v", err)
		} else {
			log.Println("Message sent to Slack successfully")
		}
	}

	if cfg.Alerting.Telegram.Enable {
		if err := SendTelegramMessage(cfg.Alerting.Telegram.BotToken, cfg.Alerting.Telegram.ChatID, alerts); err != nil {
			log.Printf("Error sending message to Telegram: %v", err)
		} else {
			log.Println("Message sent to Telegram successfully")
		}
	}

}

func Run(cfg *Config) {

	for name, chain := range cfg.Chains {
		for _, walletInfo := range chain.WalletInfo {
			go SubscribeToNewBlocks(cfg, chain, name, walletInfo.WalletAddress)
		}

	}
	go ProcessAlerts(cfg, alertChan)

	for {
		time.Sleep(1 * time.Second)
	}
}
