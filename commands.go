package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var cmd = make(chan string)

func Download(videoUrl string) error {
	// check for playlist
	if strings.Contains(videoUrl, "&list=") {
		return errors.New("playlists are not supported yet")
	}

	fmt.Println("Started download", time.Now(), videoUrl)

	// TODO check if already exists (by url or meta info)
	cmd := exec.Command(
		"yt-dlp", "--format", "140",
		// "--rm-cache-dir",
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

var queue = make(chan string, 2)
var isPlaying = false

func Play(
	s *discordgo.Session,
	voiceStates []*discordgo.VoiceState,
	guildId string,
	authorId string,
	query string,
) {
	var filename = cm.Closest(query)
	// TODO what if nothing is found
	queue <- filename

	fmt.Println("Queued", filename)

	if !isPlaying {
		// Look for the message sender in that guild's current voice states.
		for _, vs := range voiceStates {
			if vs.UserID == authorId {
				isPlaying = true
				err := playSound(s, guildId, vs.ChannelID)
				if err != nil {
					fmt.Println("Error playing sound:", err)
				}
				isPlaying = false

				return
			}
		}

		fmt.Println("Looks like the message author was not in any VC")
	}
}

func QueueAction(action string) {
	cmd <- action
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

	output := fmt.Sprintf(`ButtBot v0.5-alpha-release-candidate
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
		" - !butt skip -- skip current playing\n" +
		" - !butt stop -- exit\n" +
		"4. System info `!butt info`"
}
