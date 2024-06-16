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
	"slices"
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

func Download(videoUrl string) error {
	// check for playlist
	if strings.Contains(videoUrl, "&list=") {
		return errors.New("playlists are not supported yet")
	}

	fmt.Println("Started download", time.Now(), videoUrl)

	// TODO check if already exists (by url or meta info)
	cmd := exec.Command(
		"yt-dlp", "--format", "140",
		"--rm-cache-dir",
		"-x", "--audio-format", "opus",
		"--audio-quality", "0",
		"-o", "downloads/%(title)s.%(ext)s",
		videoUrl,
	)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(string(out))
		fmt.Println(err)

		return errors.New("your requested audio couldnt be downloaded")
	}

	// extract filename from the output
	extractNameCmd := "echo '" + string(out) + "' | grep opus | sed 's/.*\\///'"
	cmd = exec.Command("bash", "-c", extractNameCmd)
	out, err = cmd.Output()
	filename := strings.TrimSpace(string(out))

	if err != nil {
		fmt.Println(string(out))
		fmt.Println(err)

		return errors.New("failed to extract filename")
	}

	// encode to DCA format
	convertToDcaCmd := "ffmpeg -i 'downloads/" + filename + "' -f s16le -ar 48000 -ac 2 pipe:1 | dca > 'downloads/" + filename + ".dca'; rm 'downloads/" + filename + "'"
	cmd = exec.Command("bash", "-c", convertToDcaCmd)
	_, err = cmd.Output()

	if err != nil {
		fmt.Println("couldnt encode to DCA", err)

		return errors.New("couldnt encode to DCA")
	}

	fmt.Println("Finished download", time.Now())
	return nil
}

func Play(
	s *discordgo.Session,
	voiceStates []*discordgo.VoiceState,
	guildId string,
	authorId string,
	query string,
) {
	var filename = cm.Closest(query)

	fmt.Println("Now playing", filename)

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

var buffer [][]byte

// loadSound attempts to load an encoded sound file from disk.
func loadSound(filename string) error {

	file, err := os.Open("downloads/" + filename)
	if err != nil {
		fmt.Println("Error opening dca file :", err)
		return err
	}

	var opuslen int16
	buffer = make([][]byte, 0)

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

		if opuslen < 0 {
			return errors.New("frame size is negative, possibly corrupted")
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

// TODO check character limit
func List() ([]string, error) {
	entries, err := os.ReadDir("downloads/")
	if err != nil {
		return nil, err
	}

	var files []string
	var total = len(entries)
	for i := 0; i < total && i < 10; i++ {
		entry := entries[i]
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

func Info() string {
	// neofetch without colors
	cmd := "neofetch | sed 's/\x1B\\[[0-9;]*m//g'"
	stats, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return err.Error()
	}

	_, os, _ := strings.Cut(string(stats), "OS: ")
	os, _, _ = strings.Cut(os, "\n")
	_, cpu, _ := strings.Cut(string(stats), "CPU: ")
	cpu, _, _ = strings.Cut(cpu, "\n")
	_, ram, _ := strings.Cut(string(stats), "Memory: ")
	ram, _, _ = strings.Cut(ram, "\n")
	_, host, _ := strings.Cut(string(stats), "Host: ")
	host, _, _ = strings.Cut(host, "\n")
	_, kernel, _ := strings.Cut(string(stats), "Kernel: ")
	kernel, _, _ = strings.Cut(kernel, "\n")
	_, uptime, _ := strings.Cut(string(stats), "Uptime: ")
	uptime, _, _ = strings.Cut(uptime, "\n")

	output := fmt.Sprintf(`ButtBot v0.4-alpha-release-candidate
	OS: %s
	CPU: %s
	RAM: %s
	Host: %s
	Kernel: %s
	Uptime: %s
	`, os, cpu, ram, host, kernel, uptime)

	return output
}

func Help() string {
	return "Usage:\n" +
		"1. Download `!butt <youtube-url>`\n" +
		"2. List all media `!butt list`\n" +
		"3. Play media `!butt play <keyword>`\n" +
		"4. System info `!butt info`"
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
