package commands

import (
	"github.com/aattola/sleier-go/app"
	"github.com/bwmarrin/discordgo"
	"log"
)

func InteractionHandler(e *discordgo.InteractionCreate) {

	discord := app.GetDiscord()

	if e.Type != discordgo.InteractionApplicationCommand {
		log.Println("Jotain muuta....")
		return
	}

	// chattikomento 100%
	cmd := e.Interaction.ApplicationCommandData()

	log.Println("Komento: ", cmd.Name)

	switch cmd.Name {
	case "soitadev":
		Soita(e.Interaction)
	}

	_ = discord.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "moro",
		},
	})

}
