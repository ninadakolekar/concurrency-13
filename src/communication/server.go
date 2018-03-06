package communication

import (
	"fmt"
	"log"
	"net/http"

	constants "github.com/IITH-POPL2-Jan2018/concurrency-13/src/constants"
	message "github.com/IITH-POPL2-Jan2018/concurrency-13/src/message"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // To convert HTTP GET Request to WebSocket

// Server ... Structure: Server
type Server struct {
	Pattern      string              // URL of the server
	ClientCount  uint32              // Number of Clients Connected
	Clients      map[uint32]*Client  // User Connections
	Upgrader     *websocket.Upgrader // HTTP to WebSocket Upgrader
	CeaseUpdates chan bool
}

// NewServer ... Function to initialize new server
func NewServer(pattern string) *Server {
	return &Server{
		Pattern:     pattern,
		ClientCount: uint32(0),
		Clients:     make(map[uint32]*Client),
		Upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			}},
		CeaseUpdates: make(chan bool),
	}
}

// Listen ... Make the server listen
func (s *Server) Listen() {
	log.Println("Listening Server ...")

	handler := func(w http.ResponseWriter, r *http.Request) {

		if s.validateUser() {

			log.Println("Client Count after entry: ", s.ClientCount)

			conn, err := s.Upgrader.Upgrade(w, r, nil)

			log.Println("Client Count after upgrade: ", s.ClientCount)

			if err != nil {
				log.Println("Upgrade : ", err)
				return
			}

			log.Println("Client Count after err: ", s.ClientCount)

			cl := NewClient(s, conn, s.nextClientID())

			fmt.Println("Client Count after newC: ", s.ClientCount)

			s.Clients[cl.ID] = cl

			fmt.Println("Client Count after acc: ", s.ClientCount)
			s.ClientCount = s.ClientCount + 1

			log.Println("Added new client. Now", len(s.Clients), "clients connected.")

			cl.HandleNewUserConnected(cl.ID)
			cl.Listen()
		} else {

			log.Println("Maximum Connections Limit Reached! Rejecting any further requests.")

		}

	}

	http.HandleFunc(s.Pattern, handler)
}

//Broadcast ... Send message to all the clients
func (s *Server) Broadcast(msg message.Message) {
	for _, c := range s.Clients {
		go c.SendMessage(msg)
		// log.Println("Broadcasted now...")
	}
}

//BroadcastNewUser ... Send message to all the clients
func (s *Server) BroadcastNewUser(msg message.NewClientMessage) {
	for _, c := range s.Clients {
		c.SendNewUserMessage(msg)
		log.Println("Broadcasted New User now...")
	}
}

//SendMessageToClient ... Send message to a particular client
func (s *Server) SendMessageToClient(clientID uint32, msg message.Message) {

	client, found := s.Clients[clientID]

	if found {
		client.SendMessage(msg)
	} else {
		log.Printf("Error (SendMessageToClient) : Client %d not found!\n", clientID)
		return
	}

}

//validateUser ... Checks whether a new user is permitted or not
func (s *Server) validateUser() bool {

	return s.ClientCount < constants.MaxNumberOfClients

}

//nextClientID ... Returns the index (ID) to be assigned to the new client
func (s *Server) nextClientID() uint32 {
	return s.ClientCount
}

// TODO : Deleting Client and Reducing Client Count on Connection close
