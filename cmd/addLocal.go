package cmd

import (
	"bot-map/config"
	"bot-map/shared"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

type AddLocalCommand struct {
	Config      *config.Config
	Client      shared.HTTPClient
	Localidades map[string]string
}

func NewAddLocalCommand(config *config.Config, client shared.HTTPClient, localidades map[string]string) *CommandInfo {
	addLocalCmd := &AddLocalCommand{
		Config:      config,
		Client:      client,
		Localidades: localidades,
	}

	return &CommandInfo{
		Name:        "addlocal",
		Description: "Adiciona uma nova localidade.",
		Options: []discordgo.ApplicationCommandOption{
			{
				Name:        "nome",
				Description: "Nome da localidade",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "descricao",
				Description: "Descrição da localidade",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
		Command: addLocalCmd,
	}
}

func (c *AddLocalCommand) Execute(interaction map[string]interface{}) error {

	options := interaction["data"].(map[string]interface{})["options"].([]interface{})

	if len(options) < 2 {
		return fmt.Errorf("faltam argumentos! Use: /addlocal <nome> <descrição>")
	}

	nome := options[0].(map[string]interface{})["value"].(string)
	descricao := options[1].(map[string]interface{})["value"].(string)

	c.Localidades[nome] = descricao
	fmt.Println("Localidade adicionada:", nome)

	response := map[string]interface{}{
		"type": 4,
		"data": map[string]interface{}{
			"content": fmt.Sprintf("🗺️ Localidade **%s** adicionada!\nDescrição: ***%s***", nome, descricao),
		},
	}
	jsonResponse, _ := json.Marshal(response)

	interactionID := interaction["id"].(string)
	interactionToken := interaction["token"].(string)
	url := fmt.Sprintf("%s/interactions/%s/%s/callback", c.Config.BaseURL, interactionID, interactionToken)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonResponse))
	req.Header.Set("Authorization", "Bot "+c.Config.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		fmt.Println("Erro ao enviar resposta:", err)
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("falha ao enviar mensagem, código: %d", resp.StatusCode)
	}

	return nil
}
