package main

import (
	"errors"
	"fmt"
	"github.com/edgarcoime/3958-a2/client"
	"log"
	"net"
	"strings"
)

const (
	MaxClients = 10
	ConnType   = "tcp"
	ConnPort   = ":6666"

	ClientEnterPrivateMsg = "You have joined! Please create a nickname to chat with others.\n" +
		"Or type \"/help\" to get a list of commands.\n"
	ClientPrivateInvalidMsg         = "Invalid Command! Type \"/help\" to see valid commands.\n"
	ClientPrivateMsgNoNicknameError = "You must create a nickname to message anyone first\n"
	ClientPrivateNicknameTakenError = "That nickname has been taken try another one :(\n"

	CmdPrefix     = "/"
	CmdNickname   = CmdPrefix + "nick"
	CmdMessage    = CmdPrefix + "msg"
	CmdMessageAll = CmdPrefix + "msg ;"
	CmdHelp       = CmdPrefix + "help"
)

type Chat struct {
	usernames map[string]string         // Key (uid): val(nickname)
	clients   map[string]*client.Client // key (uid): val(*Client)
	incoming  chan *client.Message
	join      chan *client.Client
	leave     chan *client.Client

	messages []string
}

func NewChat() *Chat {
	chat := &Chat{
		usernames: make(map[string]string),
		clients:   make(map[string]*client.Client),
		incoming:  make(chan *client.Message),
		join:      make(chan *client.Client),
		leave:     make(chan *client.Client),
	}
	chat.Listen()
	return chat
}

// <============== SENDING MSGES AND CHANGING NAME METHODS ===================>

func (c *Chat) Broadcast(msg string) {
	c.messages = append(c.messages, msg)
	for recipientId := range c.usernames {
		c.clients[recipientId].Outgoing <- msg
	}
}

func (c *Chat) Message(senderId string, targetUsers []string, msg string) {
	for _, targetNick := range targetUsers {
		foundUser := false
		for uid, otherNick := range c.usernames {
			if strings.ToLower(targetNick) == otherNick {
				if targetClient, ok := c.clients[uid]; ok {
					targetClient.Outgoing <- msg
					foundUser = true
					break
				}
			}
		}
		if !foundUser {
			c.clients[senderId].Outgoing <- fmt.Sprintf("User with nickname (%s) does not exist.\n", targetNick)
		}
	}
}

func (c *Chat) ChangeName(message *client.Message) {
	// Ensure valid HERE
	msgTokens := strings.Split(message.Text, " ")
	requestedNick := strings.ToLower(msgTokens[1])

	// Check if username is taken by someone NOT YOU
	for k, v := range c.usernames {
		if v == requestedNick && k != message.Client.ID {
			message.Client.Outgoing <- ClientPrivateNicknameTakenError
			return
		}
	}

	oldNick, oldUser := c.usernames[message.Client.ID]

	delete(c.usernames, message.Client.ID)
	c.usernames[message.Client.ID] = requestedNick
	message.Client.Outgoing <- fmt.Sprintf("Nickname has been changed to %s\n", requestedNick)
	if oldUser {
		c.Broadcast(fmt.Sprintf("Server: User (%s) has changed to (%s)\n", oldNick, requestedNick))
	} else {
		c.Broadcast(fmt.Sprintf("Server: New user (%s) has joined chat\n", requestedNick))
	}
}

// <==================== SERVER LIFELINE METHODS FOR CONNECTIONS =================>

func (c *Chat) Listen() {
	go func() {
		for {
			select {
			case message := <-c.incoming:
				// Message received from server
				c.parse(message)
			case conn := <-c.join:
				c.Join(conn)
			case conn := <-c.leave:
				c.Leave(conn)
			}
		}
	}()
}

func (c *Chat) Join(client *client.Client) {
	if len(c.clients) >= MaxClients {
		client.Quit()
		return
	}
	c.clients[client.ID] = client
	client.Outgoing <- ClientEnterPrivateMsg
	go func() {
		for msg := range client.Incoming {
			c.incoming <- msg
		}
		c.leave <- client
	}()
}

func (c *Chat) Leave(client *client.Client) {
	// Remove client from client map
	targetClientId := client.ID
	clientNick, hadUsername := c.usernames[targetClientId]

	// Send messages indicating leaving
	client.Outgoing <- "Leaving chat room..."
	if hadUsername {
		c.Broadcast(fmt.Sprintf("User (%s) just left the chat :(. Say Goodbye!\n", clientNick))
	}

	delete(c.clients, targetClientId)
	delete(c.usernames, targetClientId)

	// Quietly leave
	close(client.Outgoing)
	log.Println("Closed client's outgoing channel\n")
}

// <================== UTILITY FUNCTIONS TO PARSE AND VALIDATE =================>

func (c *Chat) parse(message *client.Message) {
	msgTokens := strings.Split(message.Text, " ")
	switch {
	case strings.HasPrefix(message.Text, CmdNickname):
		err := c.validateChangeNickCommand(message, msgTokens)
		if err != nil {
			return
		}
		c.ChangeName(message)
	case strings.HasPrefix(message.Text, CmdMessageAll):
		err := c.validateMessageCommand(message, msgTokens)
		if err != nil {
			return
		}
		nick, _ := c.usernames[message.Client.ID]
		finalMsg := fmt.Sprintf("%s: %s\n", nick, strings.Join(msgTokens[2:], " "))
		c.Broadcast(finalMsg)
	case strings.HasPrefix(message.Text, CmdMessage):
		err := c.validateMessageCommand(message, msgTokens)
		if err != nil {
			return
		}
		nick, _ := c.usernames[message.Client.ID]
		finalMsg := fmt.Sprintf("%s (private): %s\n", nick, strings.Join(msgTokens[2:], " "))
		c.Message(message.Client.ID, strings.Split(msgTokens[1], ","), finalMsg)
	case strings.HasPrefix(message.Text, CmdHelp):
		c.help(message.Client)
	default:
		message.Client.Outgoing <- ClientPrivateInvalidMsg
	}
}

func (c *Chat) help(client *client.Client) {
	client.Outgoing <- "\n"
	client.Outgoing <- "/help - lists all commands\n"
	client.Outgoing <- "/nick - change nickname\n"
	client.Outgoing <- "/msg ; <msg> - Message all users connected\n"
	client.Outgoing <- "/msg <nickname> <msg> - Directly message a user\n"
	client.Outgoing <- "/msg <nickname1>,<nickname2> <msg> - Directly message multiple users\n"
	client.Outgoing <- "\n"
}

func (c *Chat) validateChangeNickCommand(message *client.Message, msgTokens []string) error {
	if len(msgTokens) != 2 {
		message.Client.Outgoing <- "Invalid change nickname command try \"/nick <desired Nickname>\"\n"
		return errors.New("no Nickname found for user")
	}
	return nil
}

func (c *Chat) validateMessageCommand(message *client.Message, msgTokens []string) error {
	uid := message.Client.ID

	log.Println("Broadcasting message")

	if len(msgTokens) < 3 {
		message.Client.Outgoing <- "Invalid Message command. Try the following:\n"
		message.Client.Outgoing <- "/msg <nickname> <msg> - Directly message a user\n"
		message.Client.Outgoing <- "/msg <nickname1>,<nickname2> <msg> - Directly message multiple users\n"
		return errors.New("client invalid message command")
	}

	_, ok := c.usernames[uid]
	if !ok {
		message.Client.Outgoing <- ClientPrivateMsgNoNicknameError
		return errors.New("client no nickname")
	}

	return nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	chat := NewChat()

	listener, err := net.Listen(ConnType, ConnPort)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Println("Error closing the connection.")
		}
	}(listener)
	log.Println("Listening on " + ConnPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Error: ", err)
		}
		chat.Join(client.NewClient(conn))
	}
}
