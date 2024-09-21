package music

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

type Song struct {
	Title     string
	Link      string
	Thumbnail string
}

type QueueState int

const (
	StateIdle QueueState = iota
	StatePlaying
	StateError
)

var queue = make(map[string]*Queue)

type Queue struct {
	Queue      []Song
	NowPlaying Song
	vc         *discordgo.VoiceConnection
	Requester  *discordgo.Member
	State      QueueState
	Guild      *discordgo.Guild

	Stop bool
}

func GetQueue(guildId string) (*Queue, bool) {
	if q, ok := queue[guildId]; ok {
		return q, true
	}

	return nil, false
}

func NewQueue(guild *discordgo.Guild, vc *discordgo.VoiceConnection, requester *discordgo.Member) *Queue {
	q := &Queue{
		Queue:      []Song{},
		NowPlaying: Song{},
		vc:         vc,
		Requester:  requester,
		State:      StateIdle,
		Guild:      guild,
	}

	queue[guild.ID] = q

	return q
}

func (q *Queue) songEnd() {
	log.Println("Song ended mita tehda?")
}

func (q *Queue) AddSong(song Song) {
	q.Queue = append(q.Queue, song)
}

func (q *Queue) RemoveSong(index int) {
	q.Queue = append(q.Queue[:index], q.Queue[index+1:]...)
}

func (q *Queue) NextSong() {
	if len(q.Queue) > 0 {
		q.NowPlaying = q.Queue[0]
		q.Queue = q.Queue[1:]
	}
}
