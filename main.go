package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

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
		if len(tokens) < 2 || len(tokens) > 3 {
			fmt.Println("Wrong number of command args")
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
				return
			}
			s.ChannelMessageSend(m.ChannelID, strings.Join(files, "\n"))
		} else if tokens[1] == "play" {
			Play(s, g.VoiceStates, g.ID, m.Author.ID, tokens[2])
			fmt.Println("done playing (TODO make async)")
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

				// Update files list for search
				files, _ := List()
				cm = closestmatch.New(files, bagSizes)
			}()
		}

	}
}

func Download(videoUrl string) error {
	fmt.Println("Started download", time.Now(), videoUrl)

	cmd := exec.Command("youtube-dl", "--format", "140", "--rm-cache-dir", "-o", "downloads/%(title)s.%(ext)s", videoUrl)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(string(out))
		fmt.Println(err)

		return errors.New("your requested audio couldnt be downloaded")
	}

	fmt.Println("Finished? download", time.Now())
	return nil
}

func Play(
	s *discordgo.Session,
	voiceStates []*discordgo.VoiceState,
	guildId string,
	authorId string,
	query string,
) {
	filename := cm.Closest(query)

	// Load the sound file.
	err := loadSound(filename)
	if err != nil {
		fmt.Println("Error loading sound: ", err)
		return
	}

	// Look for the message sender in that guild's current voice states.
	for _, vs := range voiceStates {
		if vs.UserID == authorId {
			err := playSound(s, guildId, vs.ChannelID)
			if err != nil {
				fmt.Println("Error playing sound:", err)
			}

			return
		}
	}

	fmt.Println("Looks like the message author was not in any VC")
}

var buffer = make([][]byte, 0)

// loadSound attempts to load an encoded sound file from disk.
func loadSound(filename string) error {

	file, err := os.Open("downloads/" + filename)
	if err != nil {
		fmt.Println("Error opening dca file :", err)
		return err
	}

	var opuslen int16

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				return err
			}
			return nil
		}

		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return err
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return err
		}

		// Append encoded pcm data to the buffer.
		buffer = append(buffer, InBuf)
	}
}

// playSound plays the current buffer to the provided channel.
func playSound(s *discordgo.Session, guildID, channelID string) (err error) {

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.
	vc.Speaking(true)

	// Send the buffer data.
	for _, buff := range buffer {
		vc.OpusSend <- buff
	}

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	// Disconnect from the provided voice channel.
	vc.Disconnect()

	return nil
}

func List() ([]string, error) {
	entries, err := os.ReadDir("downloads/")
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}
