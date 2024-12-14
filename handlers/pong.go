package handlers

import (
	"github.com/RegaliaXYZ/opgg-discord/types"
	"github.com/bwmarrin/discordgo"
)

var PongCommand = types.Command{
	Name:        "pong",
	Description: "Replies with pong",
	Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		embed := &discordgo.MessageEmbed{
			Title:       "Regalia#xyz",
			URL:         "https://www.op.gg/summoners/euw/Regalia-xyz",
			Description: "LV.1090",
			Color:       0x0000ff, // Green color for the embed
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Ranked Solo/Duo",
					Value:  "Master 1\n10LP\n46G 29W 17L 63%",
					Inline: false,
				},
				{
					Name:   "Most Champion 3",
					Value:  "Zeri 5.0 KDA 80%\nKalista 1.8 KDA 100%\nJinx 10.0KDA 100%",
					Inline: false,
				},
				{
					Name:   "Match History",
					Value:  "Recent 10 Games",
					Inline: false,
				},
				{
					Name:   "10G 10W 0L",
					Value:  ":regional_indicator_w: " + types.Zeri + " Ranked Solo/Duo 10/0/0 10.0KDA",
					Inline: false,
				},
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://ddragon.leagueoflegends.com/cdn/14.24.1/img/champion/Ezreal.png",
			},
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "pong",                           // Optional text content
				Embeds:  []*discordgo.MessageEmbed{embed}, // Attach the embed here
			},
		})
	},
}
