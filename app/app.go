package app

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
)

var client *Client

type Client struct {
	Discord *discordgo.Session
}

func GetDiscord() *discordgo.Session {
	return GetInstance().Discord
}

func GetInstance() *Client {
	if client == nil {
		client = &Client{
			Discord: newDiscord(),
		}
	}

	return client
}

func newDiscord() *discordgo.Session {

	token := os.Getenv("DISCORD_TOKEN")
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("discord Create Error : ")
		panic(err)
	}

	discord.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates
	discord.StateEnabled = true

	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)

		err := s.UpdateCustomStatus("gaming")
		if err != nil {
			fmt.Println("Error attempting to set status to gaming")
		}
	})

	return discord
}
