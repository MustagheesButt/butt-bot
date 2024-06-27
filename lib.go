package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

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

	next := <-queue
QueueLoop:
	// Load the sound file.
	err = loadSound(next)
	if err != nil {
		fmt.Println("Error loading sound: ", err)
		return err
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)
	// Start speaking.
	vc.Speaking(true)
	fmt.Println("Now playing", next)

	// Send the buffer data.
PlayLoop:
	for _, buff := range buffer {
		select {
		case _cmd := <-cmd:
			if _cmd == "stop" {
				goto ExitQueue
			} else if _cmd == "skip" {
				break PlayLoop
			}
		default:
			vc.OpusSend <- buff
		}
	}

	// check for next in queue
	select {
	case next = <-queue:
		goto QueueLoop
	default:
		goto ExitQueue
	}

ExitQueue:

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	// Disconnect from the provided voice channel.
	vc.Disconnect()

	return nil
}
