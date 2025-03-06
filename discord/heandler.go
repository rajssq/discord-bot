package discord

import (
	"encoding/json"
	"fmt"
)

func (dc *DiscordClient) HandleEvents() error {
	for {
		var payload GatewayPayload
		if err := dc.WsConn.ReadJSON(&payload); err != nil {
			dc.HandleError(err)
			return err
		}

		switch payload.Op {
		case 0:
			dc.Sequence = *payload.S

			if payload.T != nil && *payload.T == "INTERACTION_CREATE" {
				data, _ := json.Marshal(payload.D)
				var interactionEvent map[string]interface{}
				json.Unmarshal(data, &interactionEvent)

				interactionType := interactionEvent["type"].(float64)

				if interactionType == 4 { // Autocomplete
					commandData := interactionEvent["data"].(map[string]interface{})
					commandName := commandData["name"].(string)

					cmdInfo, exists := dc.Registry.GetCommand(commandName)
					if exists {
						if autoCmd, ok := cmdInfo.Command.(interface {
							HandleAutocomplete(map[string]interface{}) error
						}); ok {
							autoCmd.HandleAutocomplete(interactionEvent)
						}
					}
				} else if interactionType == 2 { // Comando normal
					commandData := interactionEvent["data"].(map[string]interface{})
					commandName := commandData["name"].(string)

					cmdInfo, exists := dc.Registry.GetCommand(commandName)
					if exists {
						cmdInfo.Command.Execute(interactionEvent)
					}
				}
			}
		case 11:
			fmt.Println("Heartbeat recognized: ", payload.Op)
		}
	}

}
