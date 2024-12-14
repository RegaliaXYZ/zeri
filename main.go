package main

import (
	"github.com/RegaliaXYZ/opgg-discord/bot"
	"github.com/RegaliaXYZ/opgg-discord/handlers"
	"github.com/RegaliaXYZ/opgg-discord/types"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

var (
	logger   *zap.Logger
	commands = []types.Command{
		handlers.PingCommand,
		handlers.OPGGCommand,
		handlers.RiotCommand,
		handlers.PongCommand,
	}
)

func main() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("Starting bot...")
	cfg, err := bot.ReadConfig()
	if err != nil {
		logger.Fatal("Error reading configuration", zap.Error(err))
		return
	}

	/*
		patch, err := bot.FetchCurrentPatch()
		if err != nil {
			logger.Fatal("Error getting current patch version", zap.Error(err))
			return
		}
		logger.Info("Current patch version", zap.String("patch", patch))
		champions, err := bot.FetchChampionData(patch)
		if err != nil {
			logger.Fatal("Error getting champions on current patch", zap.Error(err))
			return
		}
	*/
	//logger.Info("All champions", zap.Any("champions", champions))
	session, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		logger.Fatal("Error creating Discord session", zap.Error(err))
		return
	}

	defer session.Close()

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		for _, cmd := range commands {
			if i.ApplicationCommandData().Name == cmd.Name {
				cmd.Handler(s, i)
				return
			}
		}
	})

	err = session.Open()
	if err != nil {
		logger.Fatal("Error opening connection", zap.Error(err))
		return
	}

	// Register commands
	appID := cfg.AppID // Your application ID (optional but recommended)
	guildID := ""      // Your test guild ID (for guild-specific commands during testing)

	for _, cmd := range commands {
		command := &discordgo.ApplicationCommand{
			Name:        cmd.Name,
			Description: cmd.Description,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "riot_id",
					Description: "The Riot ID of the player (Player#test)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "region",
					Description: "Region of the player",
					Type:        discordgo.ApplicationCommandOptionString,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "North America (NA)",
							Value: "NA1",
						},
						{
							Name:  "Europe West (EUW)",
							Value: "EUW1",
						},
						{
							Name:  "Europe Nordic & East (EUNE)",
							Value: "EUN1",
						},
						{
							Name:  "Oceania (OCE)",
							Value: "OC1",
						},
						{
							Name:  "Korea (KR)",
							Value: "KR",
						},
						{
							Name:  "Middle East (ME)",
							Value: "ME1",
						},
						{
							Name:  "Japan (JP)",
							Value: "JP1",
						},
						{
							Name:  "Brazil (BR)",
							Value: "BR1",
						},
						{
							Name:  "Latin America South (LAS)",
							Value: "LA1",
						},
						{
							Name:  "Latin America North (LAN)",
							Value: "LA2",
						},
						{
							Name:  "Russia (RU)",
							Value: "RU",
						},
						{
							Name:  "Türkiye (TR)",
							Value: "TR1",
						},
						{
							Name:  "Singapore (SG)",
							Value: "SG2",
						},
						{
							Name:  "Philippines (PH)",
							Value: "PH2",
						},
						{
							Name:  "Taiwan (TW)",
							Value: "TW2",
						},
						{
							Name:  "Vietnam (VN)",
							Value: "VN2",
						},
						{
							Name:  "Thailand (TH)",
							Value: "TH2",
						},
					},
					Required: true,
				},
			},
		}
		_, err := session.ApplicationCommandCreate(appID, guildID, command)

		if err != nil {
			logger.Error("Cannot create command", zap.String("command", cmd.Name), zap.Error(err))
		} else {
			logger.Info("Command registered", zap.String("command", cmd.Name))
		}
	}
	logger.Info("Bot is now running. Press CTRL+C to exit.")
	select {}
}
