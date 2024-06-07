package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Token string `yaml:"token"`
	Name  string `yaml:"name"`
}

var config Config

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

	// Declare intents
	// discord.Identify.Intents = discordgo.IntentsGuildMessages

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
		if len(tokens) != 2 {
			fmt.Println("Wrong number of command args")
			return
		}

		var vidUrl = tokens[1]
		_, err := url.ParseRequestURI(vidUrl)
		if err != nil {
			return
		}

		fmt.Println("Started download", time.Now(), vidUrl)
		go func() {
			cmd := exec.Command("youtube-dl", "--format", "140", "--rm-cache-dir", "-o", "downloads/%(title)s.%(ext)s", vidUrl)
			out, err := cmd.Output()
			if err != nil {
				fmt.Println(string(out))
				fmt.Println(err)
				s.ChannelMessageSend(m.ChannelID, "Your requested audio couldnt be downloaded")
				return
			}
			fmt.Println("Finished? download", time.Now())
		}()

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

		_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s> Your request has been queued for download", m.Author.ID))
		if err != nil {
			fmt.Println(err)
		}

		// Look for the message sender in that guild's current voice states.
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
				// err = playSound(s, g.ID, vs.ChannelID)
				// if err != nil {
				// 	fmt.Println("Error playing sound:", err)
				// }

				return
			}
		}

		fmt.Println("Looks like the message author was not in any VC")
	}
}
