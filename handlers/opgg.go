package handlers

import (
	"github.com/RegaliaXYZ/opgg-discord/types"
	"github.com/bwmarrin/discordgo"
)

var OPGGCommand = types.Command{
	Name:        "ping",
	Description: "Replies with pong",
	Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "pong",
			},
		})
	},
}
