package pkg

import (
	"log"
	"os"

	"github.com/go-yaml/yaml"
)

type Config struct {
	Alerting Alerting               `yaml:"alerting"`
	Chains   map[string]ChainConfig `yaml:"chains"`
}
type Alerting struct {
	Slack struct {
		Enable     bool   `yaml:"enable"`
		WebhookURL string `yaml:"webhook_url"`
	} `yaml:"slack"`
	Telegram struct {
		Enable   bool   `yaml:"enable"`
		BotToken string `yaml:"bot_token"`
		ChatID   string `yaml:"chat_id"`
	} `yaml:"telegram"`
	Discord struct {
		Enable     bool   `yaml:"enable"`
		WebhookURL string `yaml:"webhook_url"`
	} `yaml:"discord"`
}
type ChainConfig struct {
	RPC        string `yaml:"rpc"`
	API        string `yaml:"api"`
	GRPC       string `yaml:"grpc"`
	Explorer   string `yaml:"explorerURL"`
	WalletInfo []struct {
		WalletAddress string `yaml:"wallet_address"`
	} `yaml:"wallet_Info"`
}

func LoadConfig(path string) (*Config, error) {
	// 환경 변수에서 경로를 가져오거나, 없을 경우 기본 경로 사용
	// fullPath := os.Getenv("CONFIG_PATH")
	// if fullPath == "" {
	// 	fullPath = "/.mkvBackend/config.yml" // 기본 경로
	// }

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Error reading config file from %s: %v", path, err)
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Printf("Error parsing config file: %v", err)
		return nil, err
	}

	return &config, nil
}
