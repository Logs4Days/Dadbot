
#  Dadbot

Remaking my favorite Discord bot in Golang

## What is this abomination?

 - A fun an engaging way to bother everyone in a Discord server.
 ![Dadbot Example](https://i.imgur.com/uHBZMyH.png)

## Usage:

1. Configure Discord's end for your bot <br>
https://discordnet.dev/guides/getting_started/first-bot.html	
	 
2. Compile Dadbot.go<br>
`go build Dadbot.go` 
	
3. Get your DISCORD_BOT_TOKEN<br>
![Bot token](https://i.imgur.com/yCfOMFV.png)

4. Run it!<br>
`./Dadbot -t DISCORD_BOT_TOKEN`<br>
Or set environment variable: `export DISCORD_BOT_TOKEN=your_token_here && ./Dadbot`

5. Wait for people to tell you to remove it<br>
![profit](https://i.imgur.com/Ij2h3dW.png)

## Systemd Service:
For production deployment, use the included systemd service file:
1. Copy `dadbot.service` to `/etc/systemd/system/`
2. Update the token in the service file or use systemd environment files
3. `sudo systemctl enable dadbot && sudo systemctl start dadbot`

## Features:
- **Structured Logging**: JSON output for systemd/Elastic ingestion
- **Metrics Tracking**: Event logging with metrics for dashboards
- **Environment Config**: Supports both CLI flags and environment variables

## Note:
I haven't tested this on Windows. Should work on all *.nix platforms. We push to prod without testing here 🎉
