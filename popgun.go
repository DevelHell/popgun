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
	STATE_AUTHORIZATION = iota
	STATE_TRANSACTION
	STATE_UPDATE
)

type Config struct {
	ListenInterface string `json:"listen_interface"`
}

type Authorizator interface {
	Authorize(user, pass string) bool
	IsAuthorized() bool
}

type Backend interface {
	Stat(user string) (messages, octets int, err error)
	List(user string) ([][2]int, error)
	ListMessage(user string, msgId int) (exists bool, octets int, err error)
}

var (
	ErrInvalidState = fmt.Errorf("Invalid state")
	ErrUnauthorized = fmt.Errorf("Unauthorized")
)

//---------------CLIENT

type Client struct {
	commands     map[string]Executable
	printer      *Printer
	isAlive      bool
	currentState int
	authorizator Authorizator
	backend      Backend
	user         string
	pass         string
	lastCommand  string
}

func newClient(authorizator Authorizator, backend Backend) *Client {
	commands := make(map[string]Executable)

	commands["QUIT"] = QuitCommand{}
	commands["USER"] = UserCommand{}
	commands["PASS"] = PassCommand{}
	commands["STAT"] = StatCommand{}
	commands["LIST"] = ListCommand{}
	commands["RETR"] = RetrCommand{}
	commands["DELE"] = DeleCommand{}
	commands["NOOP"] = NoopCommand{}
	commands["RSET"] = RsetCommand{}

	return &Client{
		commands:     commands,
		currentState: STATE_AUTHORIZATION,
		authorizator: authorizator,
		backend:      backend,
	}
}

func (c Client) handle(conn net.Conn) {
	defer conn.Close()
	c.printer = NewPrinter(conn)

	c.isAlive = true
	reader := bufio.NewReader(conn)

	c.printer.Welcome()

	for c.isAlive {
		// according to RFC commands are terminated by CRLF, but we are removing \r in parseInput
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
			c.printer.Err("Invalid command %s", cmd)
			log.Printf("Invalid command: %s", cmd)
			continue
		}
		state, err := exec.Run(&c, args)
		if err != nil {
			c.printer.Err("Error executing command %s", cmd)
			log.Print("Error executing command: ", err)
			continue
		}
		c.lastCommand = cmd
		c.currentState = state
	}
}

func (c Client) parseInput(input string) (string, []string) {
	input = strings.Trim(input, "\r \n")
	cmd := strings.Split(input, " ")
	return strings.ToUpper(cmd[0]), cmd[1:]
}

//---------------SERVER

type Server struct {
	listener net.Listener
}

func (s Server) Run(cfg Config, auth Authorizator, backend Backend) error {

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

			c := newClient(auth, backend)
			go c.handle(conn)
		}
	}()

	return nil
}

//---------------PRINTER

type Printer struct {
	conn net.Conn
}

func NewPrinter(conn net.Conn) *Printer {
	return &Printer{conn}
}

func (p Printer) Welcome() {
	fmt.Fprintf(p.conn, "+OK POPgun POP3 server ready\r\n")
}

func (p Printer) Ok(msg string, a ...interface{}) {
	msg = strings.Replace(msg, "\n", "\r\n", -1)
	fmt.Fprintf(p.conn, "+OK %s\r\n", fmt.Sprintf(msg, a...))
}

func (p Printer) Err(msg string, a ...interface{}) {
	msg = strings.Replace(msg, "\n", "\r\n", -1)
	fmt.Fprintf(p.conn, "-ERR %s\r\n", fmt.Sprintf(msg, a...))
}
