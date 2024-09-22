package main

import (
	"github.com/aattola/sleier-go/app"
	"github.com/aattola/sleier-go/commands"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	c := make(chan bool)

	client := app.GetInstance()
	discord := client.Discord

	err = discord.Open()
	if err != nil {
		log.Panic("Error opening Discord session: ", err)
	}

	cmd := &discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        "testi",
		Description: "testaa jotain",
	}

	_, _ = discord.ApplicationCommandCreate(discord.State.User.ID, "214761475422683136", cmd)

	cmd = &discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        "soitadev",
		Description: "Soittaa jotain",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "hakusana",
				Description: "yt-dlp soitettava linkki / hakusana youtubeen",
				Required:    true,
			},
		},
	}

	createdCommand, err := discord.ApplicationCommandCreate(discord.State.User.ID, "214761475422683136", cmd)
	if err != nil {
		panic("Error creating command: " + err.Error())
	}

	log.Println("Created command: ", createdCommand.Name)

	discord.AddHandler(func(s *discordgo.Session, r *discordgo.InteractionCreate) {
		commands.InteractionHandler(r)
	})

	defer discord.Close()
	<-c
}
