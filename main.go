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
	"os/signal"
	"strconv"
	"syscall"
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
	s.UpdateCustomStatus("gaming")
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

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	vc, err := dg.ChannelVoiceJoin("214761475422683136", "214761475422683137", false, true)
	if err != nil {
		fmt.Errorf("error joining voice channel: %v", err)
		os.Exit(1)
	}

	vc.Speaking(true)

	r, w, err := os.Pipe()
	if err != nil {
		fmt.Errorf("error creating pipe: %v", err)
		return
	}

	ytdlCmd := exec.Command("yt-dlp", "-x", "--audio-format", "opus", "--audio-quality", "0", "-o", "-", "https://www.youtube.com/watch?v=I4hrPA2a9RI")
	ytdlCmd.Stdout = w
	err = ytdlCmd.Start()
	if err != nil {
		fmt.Errorf("error starting yt-dlp: %v", err)
		return
	}
	defer ytdlCmd.Wait()

	ffmpegCmd := exec.Command("ffmpeg", "-i", "pipe:0", "-f", "s16le", "-ar", strconv.Itoa(FRAME_RATE), "-ac", strconv.Itoa(CHANNELS), "pipe:1")
	//ffmpegCmd := exec.Command("ffmpeg", "-i", "pipe:0", "-ar", "48000", "-ac", "2", "-b:a", "64k", "-f", "ogg", "-c:a", "libopus", "-application", "audio", "-frame_size", "960", "-")
	ffmpegCmd.Stdin = r

	//cmd := exec.Command("yt-dlp", "-x", "--audio-format", "opus", "--audio-quality", "0", "-o", "-", "https://www.youtube.com/watch?v=I4hrPA2a9RI")
	stdout, _ := ffmpegCmd.StdoutPipe()
	ffmpegCmd.Start()

	defer vc.Disconnect()

	encoder, err := gopus.NewEncoder(FRAME_RATE, CHANNELS, gopus.Audio)

	buffer := bufio.NewReaderSize(stdout, 16384)

	for {
		audioBuffer := make([]int16, FRAME_SIZE*CHANNELS)
		err = binary.Read(buffer, binary.LittleEndian, &audioBuffer)
		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			fmt.Errorf("error reading from ffmpeg stdout: %v", err)
			return
		}
		if err != nil {
			fmt.Errorf("error reading from ffmpeg stdout: %v", err)
			return
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

	//	defer vc.Disconnect()
	//	scanner := bufio.NewScanner(stdout)
	//	for scanner.Scan() {
	//		m := scanner.Bytes()
	//
	//		vc.OpusSend <- m
	//	}
	ffmpegCmd.Wait()

	fmt.Println("Pyörii.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}
