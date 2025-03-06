package main

import (
	"bot-map/cmd"
	"bot-map/config"
	"bot-map/discord"
	"fmt"
	"net/http"
)

func main() {
	configInstance := config.LoadConfig()

	// ✅ Criação do mapa compartilhado de localidades
	localidades := make(map[string]string)

	// Registra o comando /addlocal
	addLocalCmd := cmd.NewAddLocalCommand(configInstance, &http.Client{}, localidades)
	registry := cmd.NewCommandRegistry()
	registry.RegistryCommand(addLocalCmd)

	// Registra o comando /local
	localCmd := cmd.NewLocalCommand(configInstance, &http.Client{}, localidades)
	registry.RegistryCommand(localCmd)

	// 🔍 Verificar se ambos os comandos estão usando o mesmo mapa
	fmt.Println("Mapa de localidades antes do bot iniciar:", localidades)

	// Inicializa o cliente do Discord
	discordClient := discord.GetDiscordClient(configInstance, registry)
	discordClient.RegisterSlashCommands()

	// Conecta ao gateway do Discord
	if err := discordClient.ConnectToGateway(); err != nil {
		discordClient.HandleError(err)
	}

	// Mantém o bot rodando
	select {}
}
