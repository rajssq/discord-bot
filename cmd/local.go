package cmd

import (
	"bot-map/config"
	"bot-map/shared"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type LocalCommand struct {
	Config      *config.Config
	Client      shared.HTTPClient
	Localidades map[string]string
}

func NewLocalCommand(config *config.Config, client shared.HTTPClient, localidades map[string]string) *CommandInfo {
	localCmd := &LocalCommand{
		Config:      config,
		Client:      client,
		Localidades: localidades,
	}

	return &CommandInfo{
		Name:        "local",
		Description: "Mostra uma lista de locais disponíveis ou detalhes de um local específico.",
		Options: []discordgo.ApplicationCommandOption{
			{
				Name:         "nome",
				Description:  "Nome do local para obter detalhes",
				Type:         discordgo.ApplicationCommandOptionString,
				Required:     false,
				Autocomplete: true,
			},
		},
		Command: localCmd,
	}
}

func (c *LocalCommand) Execute(interaction map[string]interface{}) error {
	options, ok := interaction["data"].(map[string]interface{})["options"].([]interface{})

	fmt.Println("Locais armazenados no momento:", c.Localidades)

	var responseText string

	if !ok || len(options) == 0 {
		if len(c.Localidades) == 0 {
			responseText = "Nenhuma localidade cadastrada ainda! Use `/addlocal` para adicionar uma."
		} else {
			responseText = "**Locais disponíveis:**\n"
			for nome := range c.Localidades {
				responseText += fmt.Sprintf("- %s\n", nome)
			}
		}
	} else {
		nome := options[0].(map[string]interface{})["value"].(string)
		descricao, existe := c.Localidades[nome]
		if existe {
			responseText = fmt.Sprintf("️🧭 **%s**\n\n- ***%s***", nome, descricao)
		} else {
			responseText = fmt.Sprintf("❌ Localidade '%s' não encontrada.", nome)
		}
	}

	response := map[string]interface{}{
		"type": 4,
		"data": map[string]interface{}{
			"content": responseText,
		},
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Println("Erro ao serializar resposta:", err)
		return err
	}

	interactionID := interaction["id"].(string)
	interactionToken := interaction["token"].(string)

	url := fmt.Sprintf("%s/interactions/%s/%s/callback", c.Config.BaseURL, interactionID, interactionToken)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonResponse))
	if err != nil {
		log.Println("Erro ao criar requisição:", err)
		return err
	}

	req.Header.Set("Authorization", "Bot "+c.Config.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		log.Println("Erro ao enviar resposta:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Falha ao enviar mensagem, código: %d", resp.StatusCode)
	}

	return nil
}

func (c *LocalCommand) HandleAutocomplete(interaction map[string]interface{}) error {
	interactionID := interaction["id"].(string)
	interactionToken := interaction["token"].(string)

	data := interaction["data"].(map[string]interface{})
	options := data["options"].([]interface{})

	focusedOption := options[0].(map[string]interface{})
	inputValue := focusedOption["value"].(string)

	var suggestions []map[string]interface{}
	for nome := range c.Localidades {
		if strings.HasPrefix(strings.ToLower(nome), strings.ToLower(inputValue)) {
			suggestions = append(suggestions, map[string]interface{}{
				"name":  nome,
				"value": nome,
			})
		}
	}

	response := map[string]interface{}{
		"type": 8,
		"data": map[string]interface{}{
			"choices": suggestions,
		},
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Println("Erro ao serializar resposta de autocomplete:", err)
		return err
	}

	url := fmt.Sprintf("%s/interactions/%s/%s/callback", c.Config.BaseURL, interactionID, interactionToken)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonResponse))
	if err != nil {
		log.Println("Erro ao criar requisição HTTP:", err)
		return err
	}

	req.Header.Set("Authorization", "Bot "+c.Config.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		log.Println("Erro ao enviar resposta de autocomplete:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Erro na resposta de autocomplete:", resp.StatusCode)
		return fmt.Errorf("erro na resposta de autocomplete: %d", resp.StatusCode)
	}

	return nil
}
