package client

// Message contains information of the sender and the text.
// This makes it easier to send necessary information about a message from client to server
type Message struct {
	Client *Client
	Text   string
}

// NewMessage Creates a new message with the given client and text
func NewMessage(client *Client, text string) *Message {
	return &Message{
		Client: client,
		Text:   text,
	}
}

// String Stringifies Message into text
//func (message *Message) String() string {
//	return fmt.Sprintf("%s: %s\n", message.Client.Name, message.Text)
//}
