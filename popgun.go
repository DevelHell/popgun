/*
- implementation of POP3 server according to rfc1939
- inspired a bit by https://github.com/r0stig/golang-pop3
*/

package popgun

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

const (
	STATE_UNAUTHORIZED = iota
	STATE_TRANSACTION
	STATE_UPDATE
)

type Config struct {
	ListenInterface string `json:"listen_interface"`
}

//---------------CLIENT

type Client struct {
	commands     map[string]Executable
	isAlive      bool
	currentState int
}

func newClient() *Client {
	commands := make(map[string]Executable)

	commands["QUIT"] = QuitCommand{}

	return &Client{
		commands: commands,
	}
}

func (c Client) handle(conn net.Conn) {
	defer conn.Close()

	c.isAlive = true
	reader := bufio.NewReader(conn)

	fmt.Fprintln(conn, "+OK POPgun POP3 server ready")

	for c.isAlive {
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Print("Connection closed by client")
			} else {
				log.Print("Error reading input: ", err)
			}
			break
		}

		cmd, args := c.parseInput(input)
		exec, ok := c.commands[cmd]
		if !ok {
			log.Printf("Invalid command %s", cmd)
			continue
		}
		state, err := exec.Run(&c, args)
		if err != nil {
			log.Print("Error executing command: ", err)
			continue
		}
		c.currentState = state
	}
}

func (c Client) parseInput(input string) (string, []string) {
	input = strings.Trim(input, "\r \n")
	cmd := strings.Split(input, " ")
	return cmd[0], cmd[1:]
}

//---------------SERVER

type Server struct {
	listener net.Listener
}

func (s Server) Run(cfg Config) error {

	var err error
	s.listener, err = net.Listen("tcp", cfg.ListenInterface)
	if err != nil {
		log.Printf("Error: could not listen on %s", cfg.ListenInterface)
		return err
	}

	go func() {
		fmt.Printf("Server listening on: %s\n", cfg.ListenInterface)
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				log.Print("Error: could not accept connection: ", err)
				continue
			}

			c := newClient()
			go c.handle(conn)
		}
	}()

	return nil
}
