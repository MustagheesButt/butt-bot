package main

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/schollz/closestmatch"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Token string `yaml:"token"`
	Name  string `yaml:"name"`
}

var config Config
var bagSizes = []int{2, 3, 4, 5}
var cm *closestmatch.ClosestMatch

func init() {

	confFile, err := os.ReadFile("conf.yml")
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(confFile, &config)
	if err != nil {
		panic(err)
	}

	// Create downloads dir
	// _, err = os.Stat("downloads/")
	// if errors.Is(err, fs.ErrNotExist) {
	os.Mkdir("downloads/", 0755)
	// }

	// Create a closestmatch object
	files, _ := List()
	cm = closestmatch.New(files, bagSizes)
}

func main() {
	if config.Token == "" {
		fmt.Printf("No token provided. Please run: %s -t <bot token>\n", os.Args[0])
		return
	}

	dcSession, err := discordgo.New("Bot " + config.Token)

	if err != nil {
		fmt.Println("an unknown error occured", err)
		return
	}

	dcSession.AddHandler(ready)

	// Register messageCreate as a callback for the messageCreate events.
	dcSession.AddHandler(messageCreate)

	// dcSession.AddHandler(guildCreate)

	// Declare intents
	dcSession.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates | discordgo.IntentMessageContent

	// Open the websocket and begin listening.
	err = dcSession.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Printf("%s is now running. Press CTRL-C to exit.\n", config.Name)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	dcSession.Close()

}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {

	// Set the playing status.
	s.UpdateGameStatus(0, "!butt")
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is "!butt"
	if strings.HasPrefix(m.Content, "!butt") {

		var tokens = strings.Split(m.Content, " ")
		if len(tokens) < 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, Help())
			if err != nil {
				fmt.Println(err)
			}

			return
		}

		// Find the channel that the message came from.
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			// Could not find channel.
			return
		}

		// Find the guild for that channel.
		g, err := s.State.Guild(c.GuildID)
		if err != nil {
			// Could not find guild.
			return
		}

		if tokens[1] == "list" {
			files, err := List()
			if err != nil {
				fmt.Println(err)
				return
			}
			s.ChannelMessageSend(m.ChannelID, strings.Join(files, "\n"))
		} else if tokens[1] == "play" {
			tokens = slices.Delete(tokens, 0, 1)
			keywords := strings.Join(tokens, "")
			Play(s, g.VoiceStates, g.ID, m.Author.ID, keywords)

			fmt.Println("done playing (TODO make async)")
		} else if tokens[1] == "info" {
			_, err := s.ChannelMessageSend(m.ChannelID, Info())
			if err != nil {
				fmt.Println(err)
			}
		} else {
			var vidUrl = tokens[1]
			_, err := url.ParseRequestURI(vidUrl)
			if err != nil {
				return
			}

			_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s> Your request has been queued for download", m.Author.ID))
			if err != nil {
				fmt.Println(err)
			}

			go func() {
				err := Download(vidUrl)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, err.Error())
				}

				// Update search index
				files, _ := List()
				cm = closestmatch.New(files, bagSizes)
			}()
		}

	}
}

// func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {

// 	if event.Guild.Unavailable {
// 		return
// 	}

// 	for _, channel := range event.Guild.Channels {
// 		if channel.ID == event.Guild.ID {
// 			_, _ = s.ChannelMessageSend(channel.ID, "ButtBot is ready! Type !butt while in a voice channel to play a sound.")
// 			return
// 		}
// 	}
// }
