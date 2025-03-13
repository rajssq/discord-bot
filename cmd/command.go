package cmd

import "github.com/bwmarrin/discordgo"

// Command é uma interface que define a estrutura de um comando do bot.
type Command interface {
	// Execute é o método que será implementado por cada comando, processando a interação recebida.
	Execute(interaction map[string]interface{}) error
}

// CommandInfo armazena informações sobre um comando registrado no bot.
type CommandInfo struct {
	Name        string                               // Nome do comando
	Description string                               // Descrição do comando
	Options     []discordgo.ApplicationCommandOption // Opções disponíveis para o comando (parâmetros)
	Command     Command                              // Instância do comando que implementa a interface Command
}
