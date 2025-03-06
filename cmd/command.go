package cmd

import "github.com/bwmarrin/discordgo"

type Command interface {
	Execute(interaction map[string]interface{}) error
}

type CommandInfo struct {
	Name        string
	Description string
	Options     []discordgo.ApplicationCommandOption
	Command     Command
}
