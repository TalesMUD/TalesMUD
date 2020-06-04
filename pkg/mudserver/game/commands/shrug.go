package commands

import (
	"strings"

	"github.com/atla/owndnd/pkg/mudserver/game/def"
	"github.com/atla/owndnd/pkg/mudserver/game/messages"
)

// ShrugCommand ... foo
type ShrugCommand struct {
}

// Execute ... executes scream command
func (shrugCommand *ShrugCommand) Execute(game def.GameCtrl, message *messages.Message) bool {

	parts := strings.Fields(message.Data)
	newMsg := strings.Join(parts[1:], " ")
	msg := ""

	if newMsg != "" {
		msg = "'" + newMsg + "'"
	}

	newMessage := "-- " + message.Character.Name + " shrugs " + msg + " ¯\\_(ツ)_/¯"
	out := messages.NewOutgoingMessage("", newMessage)
	out.AudienceID = message.Character.CurrentRoomID
	game.SendMessage(out)
	return true
}
