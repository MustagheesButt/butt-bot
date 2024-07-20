package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/schollz/closestmatch"
)

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"download": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Access options in the order provided by the user.
		options := i.ApplicationCommandData().Options
		// Convert to map
		optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
		for _, opt := range options {
			optionMap[opt.Name] = opt
		}

		if option, ok := optionMap["url"]; ok {
			go func() {
				var reply = "Downloaded successfully"
				err := Download(option.StringValue())

				if err != nil {
					reply = err.Error()
				}
				s.ChannelMessageSend(i.ChannelID, reply)

				// Update search index
				files, _ := List()
				cm = closestmatch.New(files, bagSizes)
			}()
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Your request has been queued for download",
			},
		})
	},
	"play": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		options := i.ApplicationCommandData().Options
		optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
		for _, opt := range options {
			optionMap[opt.Name] = opt
		}

		// Find the guild for that channel.
		g, err := s.State.Guild(i.GuildID)
		if err != nil {
			// Could not find guild.
			return
		}

		if option, ok := optionMap["query"]; ok {
			go Play(s, g.VoiceStates, g.ID, i.Member.User.ID, option.StringValue())
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Queued for playing",
			},
		})
	},
	"skip": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		QueueAction("skip")

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Skipping",
			},
		})
	},
	"stop": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		QueueAction("stop")

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Stopeed",
			},
		})
	},
	"sysinfo": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: Info(),
			},
		})
	},
}
