package mudserver

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/service"
)

// MUDServer ... server application connecting the websocket clients with the game instance, providing utility functions etc.
type MUDServer interface {
	Run()
	HandleConnections(*gin.Context)
}
type conn struct {
	User   *entities.User
	ws     *websocket.Conn
	active bool
}

/*CheckOrigin:
 */
type server struct {
	Facade service.Facade
	port   string

	Game *game.Game

	Clients   map[string]conn
	Broadcast chan interface{}
	Upgrader  websocket.Upgrader

	//	MessageHandler *MessageHandler
}

//New creates a new mud server
func New(facade service.Facade) MUDServer {

	game := game.New(facade)

	srv := &server{
		Facade: facade,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		Clients:   make(map[string]conn),
		Broadcast: make(chan interface{}),
		Game:      game,
	}

	game.Subscribe(srv)

	return srv
}

func (server *server) Run() {

	log.WithTime(time.Now()).Info("MUD Server starting ...")

	go server.Game.Run()
	go server.handleBroadcastMessages()

	log.WithTime(time.Now()).Info("MUD Server running")
}

//HandleConnections asd
func (server *server) HandleConnections(c *gin.Context) {

	var user *entities.User

	if usr, exists := c.Get("user"); exists {

		log.WithField("User", usr.(*entities.User).Nickname).Info("User logged in")
		user = usr.(*entities.User)
	}

	// Upgrade initial GET request to a websocket
	ws, err := server.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	log.Info("Upgraded client connection")

	// Register our new client
	server.Clients[user.ID.Hex()] = conn{
		User:   user,
		ws:     ws,
		active: true,
	}

	// Send Welcome message
	server.sendMessage(user.ID.Hex(), messages.NewOutgoingMessage("", "Connected to [Tales of the Red Dragon's Lair] ..."))

	server.Game.OnUserJoined <- &messages.UserJoined{
		User: user,
	}

	for {
		// Read in a new message as JSON and map it to a Message object
		var msg messages.IncomingMessage
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(server.Clients, user.ID.Hex())
			break
		}

		if msg.Message != "" {
			server.Game.OnMessageReceived() <- messages.NewMessage(user, msg.Message)
		}
	}
}
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
func (server *server) sendMessage(id string, msg interface{}) {

	if client, ok := server.Clients[id]; ok {
		err := client.ws.WriteJSON(msg)
		if err != nil {

			// tell the game that the user quit as the websocket closes/closed...
			server.Game.OnUserQuit <- &messages.UserQuit{
				User: client.User,
			}

			log.Printf("error: %v", err)
			client.ws.Close()
			delete(server.Clients, id)
		}
	}
}

func (server *server) sendToRoom(room *rooms.Room, msg interface{}) {
	server.sendToRoomWithout("", room, msg)
}

func (server *server) sendToRoomWithout(id string, room *rooms.Room, msg interface{}) {

	if id != "" {
		log.WithField("origin", id).Info("Sending to room without origin")
	}

	usersInRoom := []string{}

	//TODO build service that reads all users from
	allUsers, _ := server.Facade.UsersService().FindAll()

	for _, usr := range allUsers {
		if usr.LastCharacter != id && contains(room.Characters, usr.LastCharacter) {
			usersInRoom = append(usersInRoom, usr.ID.Hex())
		}
	}

	for _, usr := range usersInRoom {
		server.sendMessage(usr, msg)
	}
}

func (server *server) handleBroadcastMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-server.Broadcast

		// Send it out to every client that is currently connected
		for _, client := range server.Clients {
			err := client.ws.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)

				server.Game.OnUserQuit <- &messages.UserQuit{
					User: client.User,
				}

				client.ws.Close()
				delete(server.Clients, client.User.ID.Hex())

			}
		}
	}
}

// OnMessage .. broadcast receiver
func (server *server) OnMessage(message interface{}) {

	if msg, ok := message.(messages.MessageResponder); ok {
		switch msg.GetAudience() {
		case messages.MessageAudienceOrigin:
			server.sendMessage(msg.GetAudienceID(), msg)
			break
		case messages.MessageAudienceUser:
			server.sendMessage(msg.GetAudienceID(), msg)
			break
		case messages.MessageAudienceRoom:
			room, _ := server.Facade.RoomsService().FindByID(msg.GetAudienceID())
			server.sendToRoom(room, msg)
			break

		case messages.MessageAudienceRoomWithoutOrigin:
			room, _ := server.Facade.RoomsService().FindByID(msg.GetAudienceID())
			server.sendToRoomWithout(msg.GetOriginID(), room, msg)
			break

		case messages.MessageAudienceGlobal:
			server.Broadcast <- msg
			break
		case messages.MessageAudienceSystem:

			server.Broadcast <- messages.MessageResponse{
				Username: "#SYSTEM",
				Message:  msg.GetMessage(),
			}
			break
		}
	}
}

// OnSystemMessage .. broadcast receiver
func (server *server) OnSystemMessage(message *messages.Message) {

	server.Broadcast <- messages.MessageResponse{
		Username: "#SYSTEM",
		Message:  message.Data,
	}
}
