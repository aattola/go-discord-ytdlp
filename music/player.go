package music

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io"
	"layeh.com/gopus"
	"log"
	"os"
	"os/exec"
	"strconv"
)

const (
	CHANNELS  int = 2
	FrameRate int = 48000
	FrameSize int = 960
	MaxBytes  int = (FrameSize * 2) * 2
)

const (
	MIRELLA = "https://www.youtube.com/watch?v=JzzLF2An83s"
	BRUH    = "https://www.youtube.com/watch?v=kpwNjdEPz7E"
	WEEZER  = "https://www.youtube.com/watch?v=6a_9HQW1VmY"
	BOOM    = "https://www.youtube.com/watch?v=eUy6sAXeDF8"
)

func (q *Queue) Play(song Song) {
	err := q.vc.Speaking(true)
	if err != nil {
		log.Panic("Error setting speaking: ", err)
	}

	log.Println("Started speaking & connected to voice channel")

	r, w, err := os.Pipe()
	if err != nil {
		log.Panic("Error creating pipe: ", err)
	}

	// hard coded youtube video id change as needed
	ytdlCmd := exec.Command("yt-dlp", "-x", "--audio-format", "opus", "--audio-quality", "0", "-o", "-", song.Link)
	ytdlCmd.Stdout = w
	err = ytdlCmd.Start()
	if err != nil {
		log.Panic("Error starting yt-dlp: ", err)
	}

	log.Println("Audio stream started")

	ffmpegCmd := exec.Command("ffmpeg", "-i", "pipe:0", "-f", "s16le", "-ar", strconv.Itoa(FrameRate), "-ac", strconv.Itoa(CHANNELS), "-")
	ffmpegCmd.Stdin = r

	log.Println("ffmpeg conversion started")

	//cmd := exec.Command("yt-dlp", "-x", "--audio-format", "opus", "--audio-quality", "0", "-o", "-", "https://www.youtube.com/watch?v=I4hrPA2a9RI")
	stdout, _ := ffmpegCmd.StdoutPipe()
	err = ffmpegCmd.Start()
	if err != nil {
		log.Panic("Error starting ffmpeg: ", err)
	}

	go func() {
		err := ffmpegCmd.Wait()
		if err != nil {
			log.Panic("Error waiting for ffmpeg: ", err)
		}

		q.songEnd()
	}()

	defer func(vc *discordgo.VoiceConnection) {
		err := vc.Disconnect()
		if err != nil {
			log.Println("Error disconnecting from voice channel: ", err)
		}
	}(q.vc)

	encoder, err := gopus.NewEncoder(FrameRate, CHANNELS, gopus.Audio)

	buffer := bufio.NewReaderSize(stdout, 16384)

	for {

		if q.Stop {
			log.Println("Stopping song")
			break
		}

		audioBuffer := make([]int16, FrameSize*CHANNELS)
		err = binary.Read(buffer, binary.LittleEndian, &audioBuffer)
		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			log.Printf("EOF: error reading from ffmpeg stdout: %v", err)
			break
		}
		if err != nil {
			log.Printf("error reading from ffmpeg stdout: %v", err)
			break
		}

		opus, err := encoder.Encode(audioBuffer, FrameSize, MaxBytes)
		if err != nil {
			fmt.Println("Encoding error,", err)
			return
		}
		if !q.vc.Ready || q.vc.OpusSend == nil {
			fmt.Printf("Discordgo not ready for opus packets. %+v : %+v", q.vc.Ready, q.vc.OpusSend)
			return
		}

		q.vc.OpusSend <- opus

	}
}
