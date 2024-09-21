package commands

import (
	"github.com/aattola/sleier-go/app"
	"github.com/aattola/sleier-go/music"
	"github.com/bwmarrin/discordgo"
	"log"
)

func Soita(interaction *discordgo.Interaction) {
	discord := app.GetDiscord()

	var queue *music.Queue

	guild, err := discord.State.Guild(interaction.GuildID)
	if err != nil {
		log.Println("Error getting guild: ", err)
		return
	}

	queue, ok := music.GetQueue(interaction.GuildID)

	if !ok {
		// ei yhteyttä yms luodaan uusi

		member, err := discord.GuildMember(guild.ID, interaction.Member.User.ID)
		if err != nil {
			log.Println("Error getting member: ", err)
			return
		}

		var voiceChannelID string

		for _, voiceState := range guild.VoiceStates {
			if voiceState.UserID == interaction.Member.User.ID {
				voiceChannelID = voiceState.ChannelID
			}
		}

		if voiceChannelID == "" {
			log.Println("Not in voice channel")
			return
		}

		vc, err := discord.ChannelVoiceJoin(guild.ID, voiceChannelID, false, true)

		queue = music.NewQueue(guild, vc, member)
	}

	log.Println("Queue: ", queue)

	song := music.Song{
		Title:     "Mirella - Bängeri",
		Link:      "https://www.youtube.com/watch?v=6a_9HQW1VmY",
		Thumbnail: "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/MattiParkkonen_Orava.jpg/275px-MattiParkkonen_Orava.jpg",
	}

	//TODO: hae musa ja aseta nytsoivaksi

	if queue.State == music.StatePlaying {
		queue.Stop = true
	}

	go queue.Play(song)
	queue.State = music.StatePlaying

}
