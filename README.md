
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
For production deployment, use the included systemd service file with secure environment configuration:

### Setup:
1. **Configure environment file:**
   ```bash
   sudo cp etc-dadbot.env.example /etc/dadbot.env
   sudo nano /etc/dadbot.env  # Add your DISCORD_BOT_TOKEN
   sudo chown root:dadbot /etc/dadbot.env
   sudo chmod 640 /etc/dadbot.env
   ```

2. **Deploy dad:**
   ```bash
   sudo cp DadBot /usr/local/bin/
   sudo cp dadbot.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable dadbot
   sudo systemctl start dadbot
   ```

3. **Monitor logs:**
   ```bash
   # View structured JSON logs
   journalctl -u dadbot -f --output=json | jq .
   
   # Simple log viewing
   journalctl -u dadbot -f
   ```

## Note:
I haven't tested this on Windows. Should work on all *.nix platforms. We push to prod without testing here 🎉
