package cmd

// CommandRegistry é uma estrutura que gerencia o registro de comandos do bot.
type CommandRegistry struct {
	commands map[string]*CommandInfo // Mapa que armazena os comandos pelo nome
}

// NewCommandRegistry cria e retorna uma nova instância de CommandRegistry.
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]*CommandInfo), // Inicializa o mapa de comandos
	}
}

// RegistryCommand registra um novo comando no registro.
func (cr *CommandRegistry) RegistryCommand(info *CommandInfo) {
	cr.commands[info.Name] = info // Adiciona o comando ao mapa, usando seu nome como chave
}

// GetCommand retorna um comando pelo nome, se existir.
func (cr *CommandRegistry) GetCommand(name string) (*CommandInfo, bool) {
	info, exists := cr.commands[name] // Busca o comando no mapa
	return info, exists               // Retorna o comando e um booleano indicando se ele existe
}

// GetAllCommands retorna uma lista com todos os comandos registrados.
func (cr *CommandRegistry) GetAllCommands() []*CommandInfo {
	var allCommands []*CommandInfo

	// Itera sobre os comandos armazenados e adiciona à lista
	for _, cmdInfo := range cr.commands {
		allCommands = append(allCommands, cmdInfo)
	}

	return allCommands // Retorna a lista de todos os comandos registrados
}
