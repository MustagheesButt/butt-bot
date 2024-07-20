package main

import "github.com/bwmarrin/discordgo"

var slashCommands = []*discordgo.ApplicationCommand{
	{
		Name:        "download",
		Description: "Download requested song from a URL",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "url",
				Description: "URL to download audio from",
				Required:    true,
			},
		},
	},
	{
		Name:        "play",
		Description: "Adds a song from downloads to now playing queue",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "query",
				Description: "Search for a song in downloads",
				Required:    true,
			},
		},
	},
	{
		Name:        "skip",
		Description: "Skip current playing track",
	},
	{
		Name:        "stop",
		Description: "Stop whatever is playing and leave voice channel",
	},
	{
		Name:        "sysinfo",
		Description: "Prints server information",
	},
}
