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

// Estrutura que representa o comando para listar e buscar localidades
type LocalCommand struct {
	Config      *config.Config    // Configurações do bot (como a URL base e o token)
	Client      shared.HTTPClient // Cliente HTTP para fazer requisições
	Localidades map[string]string // Mapa que armazena localidades e suas descrições
}

// Função que cria e retorna um novo comando para listar/buscar localidades
func NewLocalCommand(config *config.Config, client shared.HTTPClient, localidades map[string]string) *CommandInfo {
	// Instancia a estrutura do comando com as configurações e o cliente HTTP
	localCmd := &LocalCommand{
		Config:      config,
		Client:      client,
		Localidades: localidades,
	}

	// Retorna as informações do comando para o Discord, incluindo nome, descrição e opções
	return &CommandInfo{
		Name:        "local",
		Description: "Mostra uma lista de locais disponíveis ou detalhes de um local específico.",
		Options: []discordgo.ApplicationCommandOption{
			{
				Name:         "nome",
				Description:  "Nome do local para obter detalhes",
				Type:         discordgo.ApplicationCommandOptionString,
				Required:     false,
				Autocomplete: true, // Habilita sugestões ao digitar o nome do local
			},
		},
		Command: localCmd,
	}
}

// Método que executa o comando quando chamado pelo usuário
func (c *LocalCommand) Execute(interaction map[string]interface{}) error {
	// Obtém as opções fornecidas pelo usuário na interação do Discord
	options, ok := interaction["data"].(map[string]interface{})["options"].([]interface{})

	// Exibe no console as localidades armazenadas no momento (apenas para depuração)
	fmt.Println("Locais armazenados no momento:", c.Localidades)

	var responseText string // Variável que armazenará a resposta do bot

	// Se o usuário não especificou um local, lista todas as localidades disponíveis
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
		// Se o usuário forneceu um nome, busca a descrição correspondente
		nome := options[0].(map[string]interface{})["value"].(string)
		descricao, existe := c.Localidades[nome]

		// Se a localidade existe, exibe suas informações; caso contrário, informa que não foi encontrada
		if existe {
			responseText = fmt.Sprintf("️🧭 **%s**\n\n- ***%s***", nome, descricao)
		} else {
			responseText = fmt.Sprintf("❌ Localidade '%s' não encontrada.", nome)
		}
	}

	// Monta a resposta que será enviada ao Discord
	response := map[string]interface{}{
		"type": 4, // Tipo 4 significa resposta de interação imediata (Discordgo)
		"data": map[string]interface{}{
			"content": responseText,
		},
	}

	// Converte a resposta para JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Println("Erro ao serializar resposta:", err)
		return err
	}

	// Obtém o ID e o token da interação para construir a URL de resposta
	interactionID := interaction["id"].(string)
	interactionToken := interaction["token"].(string)

	url := fmt.Sprintf("%s/interactions/%s/%s/callback", c.Config.BaseURL, interactionID, interactionToken)

	// Cria uma requisição HTTP para enviar a resposta ao Discord
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonResponse))
	if err != nil {
		log.Println("Erro ao criar requisição:", err)
		return err
	}

	req.Header.Set("Authorization", "Bot "+c.Config.Token) // Adiciona a autenticação do bot
	req.Header.Set("Content-Type", "application/json")     // Define o tipo de conteúdo como JSON

	// Envia a requisição HTTP
	resp, err := c.Client.Do(req)
	if err != nil {
		log.Println("Erro ao enviar resposta:", err)
		return err
	}
	defer resp.Body.Close() // Fecha o corpo da resposta para liberar recursos

	// Verifica se a resposta foi bem-sucedida
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Falha ao enviar mensagem, código: %d", resp.StatusCode)
	}

	return nil // Retorna nil indicando que não houve erro
}

// Método que trata o autocomplete de nomes de localidades no Discord
func (c *LocalCommand) HandleAutocomplete(interaction map[string]interface{}) error {
	// Obtém o ID e o token da interação para construir a URL de resposta
	interactionID := interaction["id"].(string)
	interactionToken := interaction["token"].(string)

	// Obtém os dados da interação, incluindo as opções digitadas pelo usuário
	data := interaction["data"].(map[string]interface{})
	options := data["options"].([]interface{})

	// Obtém a opção que está sendo preenchida e o valor digitado pelo usuário
	focusedOption := options[0].(map[string]interface{})
	inputValue := focusedOption["value"].(string)

	var suggestions []map[string]interface{} // Lista de sugestões a serem enviadas ao usuário

	// Filtra as localidades para sugerir apenas aquelas que começam com o que foi digitado
	for nome := range c.Localidades {
		if strings.HasPrefix(strings.ToLower(nome), strings.ToLower(inputValue)) {
			suggestions = append(suggestions, map[string]interface{}{
				"name":  nome,
				"value": nome,
			})
		}
	}

	// Monta a resposta do autocomplete
	response := map[string]interface{}{
		"type": 8, // Tipo 8 significa resposta de autocomplete (Discordgo)
		"data": map[string]interface{}{
			"choices": suggestions,
		},
	}

	// Converte a resposta para JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Println("Local não encontrado no autocomplete:", err)
		return err
	}

	// Cria a URL da API do Discord para responder ao autocomplete
	url := fmt.Sprintf("%s/interactions/%s/%s/callback", c.Config.BaseURL, interactionID, interactionToken)

	// Cria uma requisição HTTP para enviar as sugestões ao Discord
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonResponse))
	if err != nil {
		log.Println("Erro ao criar requisição HTTP:", err)
		return err
	}

	req.Header.Set("Authorization", "Bot "+c.Config.Token) // Adiciona a autenticação do bot
	req.Header.Set("Content-Type", "application/json")     // Define o tipo de conteúdo como JSON

	// Envia a requisição HTTP
	resp, err := c.Client.Do(req)
	if err != nil {
		log.Println("Erro ao enviar resposta de autocomplete:", err)
		return err
	}
	defer resp.Body.Close() // Fecha o corpo da resposta para liberar recursos

	// Verifica se a resposta foi bem-sucedida
	if resp.StatusCode != http.StatusOK {
		log.Println("Erro na resposta de autocomplete:", resp.StatusCode)
		return fmt.Errorf("erro na resposta de autocomplete: %d", resp.StatusCode)
	}

	return nil // Retorna nil indicando que não houve erro
}
