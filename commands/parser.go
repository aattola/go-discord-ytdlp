package commands

import "github.com/bwmarrin/discordgo"

func ParseStringArg(interaction *discordgo.Interaction, argName string) string {
	cmd := interaction.ApplicationCommandData()
	for _, option := range cmd.Options {
		if option.Type != discordgo.ApplicationCommandOptionString {
			return ""
		}

		if option.Name != argName {
			return ""
		}

		return option.StringValue()
	}

	return ""
}
