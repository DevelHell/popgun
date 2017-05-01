package popgun

import (
	"fmt"
	"log"
	"net"
)

type Config struct {
	ListenInterface string `json:"listen_interface"`
}

type Client struct {
}

func (c Client) handle(conn net.Conn) {
	defer conn.Close()
}

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
		fmt.Printf("Server listening no: %s\n", cfg.ListenInterface)
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				log.Print("Error: could not accept connection: ", err)
				continue
			}

			c := Client{}
			go c.handle(conn)
		}
	}()

	return nil
}
