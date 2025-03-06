package cmd

type CommandRegistry struct {
	commands map[string]*CommandInfo
}

func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]*CommandInfo),
	}
}

func (cr *CommandRegistry) RegistryCommand(info *CommandInfo) {
	cr.commands[info.Name] = info
}

func (cr *CommandRegistry) GetCommand(name string) (*CommandInfo, bool) {
	info, exists := cr.commands[name]
	return info, exists
}

func (cr *CommandRegistry) GetAllCommands() []*CommandInfo {
	var allCommands []*CommandInfo

	for _, cmdInfo := range cr.commands {
		allCommands = append(allCommands, cmdInfo)
	}

	return allCommands
}
