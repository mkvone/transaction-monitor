# Transaction Monitor

## Overview

Transaction Monitor is a comprehensive tool designed to track blockchain transactions across Tendermint chains. It monitors specific wallet addresses, captures transaction details, and sends alerts to various communication platforms, including Discord, Slack, and Telegram.

## Features

-   Tendermint chains Support: Tracks transactions on multiple blockchains including Odin-protocol, E-money, Kava, Konstellation, and Osmosis.
-   Custom Alerts: Sends transaction notifications to Discord, Slack, and Telegram based on user configuration.
-   Flexible Configuration: Users can specify which wallets to monitor and configure settings for each supported communication platform.
-   Real-Time Monitoring: Utilizes WebSocket connections for real-time transaction tracking.

## Installation

1. Clone the repository or download the source code.
2. Ensure that Go is installed on your system.
3. Navigate to the project directory and install dependencies if needed.

## Configuration

Edit the config.yml file to set up the application:

-   Alerting: Configure the communication platforms to send alerts (Discord, Slack, Telegram).
-   Chains: Define the blockchain networks to monitor, including RPC, API endpoints, explorer URLs, and wallet addresses.
    Example Configuration

```yaml
alerting:
    slack:
        enable: false
        webhook_url: https://hooks.slack.com/services/AAAAAAAAAAAAAAAAAAAAAAA/bbbbbbbbbbbbbbbbbbbbbbbb
    telegram:
        enable: false
        bot_token: 5555555555:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
        chat_id: -666666666
    discord:
        enable: false
        webhook_url: https://discord.com/api/webhooks/999999999999999999/zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz

chains:
    'chain name':
        rpc: https://rpc-odin.mkv.one
        api: https://api-odin.mkv.one
        explorerURL: https://ping.pub/odin/tx/
        wallet_Info:
            - wallet_address: odin~$~#$~@#%~@#%~@#%@#%
        # Other chain configurations...
```

## Usage

Run the application with a specified configuration file path:

```bash
git clone https://github.com/mkvone/transaction-monitor
cd transaction-monitor
go run ./main.go --config-path "./config.yml"
# or
go run main.go # default "./config.yml"
```
