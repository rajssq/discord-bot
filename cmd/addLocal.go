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

// Estrutura que adc localidade
type AddLocalCommand struct {
	Config      *config.Config    // Configurações do bot (como a URL base e o token)
	Client      shared.HTTPClient // Cliente HTTP para fazer requisições
	Localidades map[string]string // Mapa para armazenar localidades e suas descrições
}

// Função que cria e retorna um novo comando de adicionar localidade
func NewAddLocalCommand(config *config.Config, client shared.HTTPClient, localidades map[string]string) *CommandInfo {
	// Instancia a estrutura do comando com as configurações e o cliente HTTP
	addLocalCmd := &AddLocalCommand{
		Config:      config,
		Client:      client,
		Localidades: localidades,
	}

	// Retorna as informações do comando para o Discord, incluindo nome, descrição e opções
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

// Método que executa o comando quando chamado pelo usuário
func (c *AddLocalCommand) Execute(interaction map[string]interface{}) error {
	// Obtém as opções fornecidas pelo usuário na interação do Discord
	options := interaction["data"].(map[string]interface{})["options"].([]interface{})

	// Verifica se foram passados os dois argumentos necessários (nome e descrição)
	if len(options) < 2 {
		return fmt.Errorf("faltam argumentos! Use: /addlocal <nome> <descrição>")
	}

	// Extrai os valores das opções da interação (nome e descrição)
	nome := options[0].(map[string]interface{})["value"].(string)
	descricao := options[1].(map[string]interface{})["value"].(string)

	// Adiciona a localidade
	c.Localidades[nome] = descricao
	fmt.Println("Localidade adicionada:", nome)

	// Cria a resposta para ser enviada ao Discord
	response := map[string]interface{}{
		"type": 4, // Tipo 4 significa resposta de interação imediata (CHANNEL_MESSAGE_WITH_SOURCE) (Discordgo )
		"data": map[string]interface{}{
			"content": fmt.Sprintf("🗺️ Localidade **%s** adicionada!\nDescrição: ***%s***", nome, descricao),
		},
	}

	// Converte a resposta para JSON
	jsonResponse, _ := json.Marshal(response)

	// Constrói a URL da API do Discord para responder à interação
	interactionID := interaction["id"].(string)
	interactionToken := interaction["token"].(string)
	url := fmt.Sprintf("%s/interactions/%s/%s/callback", c.Config.BaseURL, interactionID, interactionToken)

	// Cria uma requisição HTTP para enviar a resposta ao Discord
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonResponse))
	req.Header.Set("Authorization", "Bot "+c.Config.Token) // Adiciona a autenticação do bot
	req.Header.Set("Content-Type", "application/json")     // Define o tipo de conteúdo como JSON

	// Envia a requisição HTTP
	resp, err := c.Client.Do(req)
	if err != nil {
		fmt.Println("Erro ao enviar resposta:", err)
		return err
	}
	defer resp.Body.Close() // Garante que o corpo da resposta será fechado após a execução

	// Verifica se a resposta foi bem-sucedida
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("falha ao enviar mensagem, código: %d", resp.StatusCode)
	}

	return nil // Retorna nil indicando que não houve erro
}
