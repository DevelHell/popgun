/*
- implementation of POP3 server according to rfc1939, rfc2449 in progress
*/

package popgun

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

const (
	STATE_AUTHORIZATION = iota + 1
	STATE_TRANSACTION
	STATE_UPDATE
)

type Config struct {
	ListenInterface string `json:"listen_interface"`
}

type Authorizator interface {
	Authorize(user, pass string) bool
}

type Backend interface {
	Stat(user string) (messages, octets int, err error)
	List(user string) (octets []int, err error)
	ListMessage(user string, msgId int) (exists bool, octets int, err error)
	Retr(user string, msgId int) (message string, err error)
	Dele(user string, msgId int) error
	Rset(user string) error
	Uidl(user string) (uids []string, err error)
	UidlMessage(user string, msgId int) (exists bool, uid string, err error)
	Update(user string) error
	Lock(user string) error
	Unlock(user string) error
}

var (
	ErrInvalidState = fmt.Errorf("Invalid state")
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
	commands["UIDL"] = UidlCommand{}
	commands["CAPA"] = CapaCommand{}

	return &Client{
		commands:     commands,
		currentState: STATE_AUTHORIZATION,
		authorizator: authorizator,
		backend:      backend,
	}
}

func (c Client) handle(conn net.Conn) {
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(1 * time.Minute))
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
			if len(c.user) > 0 {
				log.Printf("Unlocking user %s due to connection error ", c.user)
				c.backend.Unlock(c.user)
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
	config   Config
	auth     Authorizator
	backend  Backend
}

func NewServer(cfg Config, auth Authorizator, backend Backend) *Server {
	return &Server{
		config:  cfg,
		auth:    auth,
		backend: backend,
	}
}

func (s Server) Start() error {

	var err error
	s.listener, err = net.Listen("tcp", s.config.ListenInterface)
	if err != nil {
		log.Printf("Error: could not listen on %s", s.config.ListenInterface)
		return err
	}

	go func() {
		log.Printf("Server listening on: %s\n", s.config.ListenInterface)
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				log.Println("Error: could not accept connection: ", err)
				continue
			}

			c := newClient(s.auth, s.backend)
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
	fmt.Fprintf(p.conn, "+OK %s\r\n", fmt.Sprintf(msg, a...))
}

func (p Printer) Err(msg string, a ...interface{}) {
	fmt.Fprintf(p.conn, "-ERR %s\r\n", fmt.Sprintf(msg, a...))
}

func (p Printer) MultiLine(msgs []string) {
	for _, line := range msgs {
		line := strings.Trim(line, "\r")
		if strings.HasPrefix(line, ".") {
			fmt.Fprintf(p.conn, ".%s\r\n", line)
		} else {
			fmt.Fprintf(p.conn, "%s\r\n", line)
		}
	}
	fmt.Fprint(p.conn, ".\r\n")
}
