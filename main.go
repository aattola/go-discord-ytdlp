package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"io"
	"layeh.com/gopus"
	"log"
	"os"
	"os/exec"
	"strconv"
)

const (
	CHANNELS   int = 2
	FRAME_RATE int = 48000
	FRAME_SIZE int = 960
	MAX_BYTES  int = (FRAME_SIZE * 2) * 2
)

func ready(s *discordgo.Session, event *discordgo.Ready) {

	fmt.Println("Ready")
	// Set the playing status.
	err := s.UpdateCustomStatus("gaming")
	if err != nil {
		fmt.Println("Error attempting to set status to gaming")
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("DISCORD_TOKEN")
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	dg.AddHandler(ready)

	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates

	err = dg.Open()
	if err != nil {
		log.Panic("Error opening Discord session: ", err)
	}

	//hardcoded channel ids change as needed
	vc, err := dg.ChannelVoiceJoin("214761475422683136", "214761475422683137", false, true)
	if err != nil {
		log.Panic("Error joining voice channel: ", err)
	}

	err = vc.Speaking(true)
	if err != nil {
		log.Panic("Error setting speaking: ", err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		log.Panic("Error creating pipe: ", err)
	}

	// hard coded youtube video id change as needed
	ytdlCmd := exec.Command("yt-dlp", "-x", "--audio-format", "opus", "--audio-quality", "0", "-o", "-", "https://www.youtube.com/watch?v=kpwNjdEPz7E")
	ytdlCmd.Stdout = w
	err = ytdlCmd.Start()
	if err != nil {
		log.Panic("Error starting yt-dlp: ", err)
	}

	ffmpegCmd := exec.Command("ffmpeg", "-i", "pipe:0", "-f", "s16le", "-ar", strconv.Itoa(FRAME_RATE), "-ac", strconv.Itoa(CHANNELS), "-")
	//ffmpegCmd := exec.Command("ffmpeg", "-i", "pipe:0", "-ar", "48000", "-ac", "2", "-b:a", "64k", "-f", "ogg", "-c:a", "libopus", "-application", "audio", "-frame_size", "960", "-")
	ffmpegCmd.Stdin = r

	//cmd := exec.Command("yt-dlp", "-x", "--audio-format", "opus", "--audio-quality", "0", "-o", "-", "https://www.youtube.com/watch?v=I4hrPA2a9RI")
	stdout, _ := ffmpegCmd.StdoutPipe()
	err = ffmpegCmd.Start()
	if err != nil {
		log.Panic("Error starting ffmpeg: ", err)
	}

	defer func(vc *discordgo.VoiceConnection) {
		err := vc.Disconnect()
		if err != nil {
			fmt.Println("Error disconnecting from voice channel: ", err)
		}
	}(vc)

	encoder, err := gopus.NewEncoder(FRAME_RATE, CHANNELS, gopus.Audio)

	buffer := bufio.NewReaderSize(stdout, 16384)

	for {
		audioBuffer := make([]int16, FRAME_SIZE*CHANNELS)
		err = binary.Read(buffer, binary.LittleEndian, &audioBuffer)
		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			log.Printf("EOF: error reading from ffmpeg stdout: %v", err)
			break
		}
		if err != nil {
			log.Printf("error reading from ffmpeg stdout: %v", err)
			break
		}

		opus, err := encoder.Encode(audioBuffer, FRAME_SIZE, MAX_BYTES)
		if err != nil {
			fmt.Println("Encoding error,", err)
			return
		}
		if !vc.Ready || vc.OpusSend == nil {
			fmt.Printf("Discordgo not ready for opus packets. %+v : %+v", vc.Ready, vc.OpusSend)
			return
		}

		vc.OpusSend <- opus
	}
}
