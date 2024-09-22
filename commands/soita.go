package commands

import (
	"context"
	"github.com/aattola/sleier-go/app"
	"github.com/aattola/sleier-go/music"
	"github.com/bwmarrin/discordgo"
	"log"
)

func Soita(interaction *discordgo.Interaction) {
	discord := app.GetDiscord()

	hakusana := ParseStringArg(interaction, "hakusana")

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

	song, err := music.SearchYoutube(hakusana)
	if err != nil {
		log.Println("Error searching: ", err)
		return
	}

	//TODO: hae musa ja aseta nytsoivaksi

	if queue.PlayingCtxCancel != nil {
		(*queue.PlayingCtxCancel)()
	}

	ctx, cancel := context.WithCancel(context.Background())

	queue.PlayingCtxCancel = &cancel
	queue.State = music.StatePlaying
	go queue.Play(*song, ctx)

}
