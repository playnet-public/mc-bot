module github.com/playnet-public/mc-bot

go 1.17

require (
	github.com/bwmarrin/discordgo v0.23.2
	github.com/willroberts/minecraft-client v1.1.0
)

require (
	github.com/gorilla/websocket v1.4.2 // indirect
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b // indirect
	golang.org/x/sys v0.0.0-20201119102817-f84b799fce68 // indirect
)

replace (
	github.com/bwmarrin/discordgo => github.com/bwmarrin/discordgo v0.23.3-0.20210821175000-0fad116c6c2a
	github.com/willroberts/minecraft-client => github.com/playnet-public/minecraft-client v1.1.1-0.20210906020136-31450cadf1fb
)
