package popgun

import (
	"fmt"
	"strconv"
	"strings"
)

type Executable interface {
	Run(c *Client, args []string) (int, error)
}

type QuitCommand struct{}

func (cmd QuitCommand) Run(c *Client, args []string) (int, error) {
	newState := c.currentState
	if c.currentState == STATE_TRANSACTION {
		err := c.backend.Update(c.user)
		if err != nil {
			return 0, fmt.Errorf("Error updating maildrop for user %s: %v", c.user, err)
		}
		err = c.backend.Unlock(c.user)
		if err != nil {
			c.printer.Err("Server was unable to unlock maildrop")
			return 0, fmt.Errorf("Error unlocking maildrop for user %s: %v", c.user, err)
		}
		newState = STATE_UPDATE
	}

	c.isAlive = false
	c.printer.Ok("Goodbye")

	return newState, nil
}

type UserCommand struct{}

func (cmd UserCommand) Run(c *Client, args []string) (int, error) {
	if c.currentState != STATE_AUTHORIZATION {
		return 0, ErrInvalidState
	}
	if len(args) != 1 {
		return 0, fmt.Errorf("Invalid arguments count: %d", len(args))
	}
	c.user = args[0]
	c.printer.Ok("")
	return STATE_AUTHORIZATION, nil
}

type PassCommand struct{}

func (cmd PassCommand) Run(c *Client, args []string) (int, error) {
	if c.currentState != STATE_AUTHORIZATION {
		return 0, ErrInvalidState
	}
	if c.lastCommand != "USER" {
		c.printer.Err("PASS can be executed only directly after USER command")
		return STATE_AUTHORIZATION, nil
	}
	if len(args) != 1 {
		return 0, fmt.Errorf("Invalid arguments count: %d", len(args))
	}
	c.pass = args[0]
	if !c.authorizator.Authorize(c.user, c.pass) {
		c.printer.Err("Invalid username or password")
		return STATE_AUTHORIZATION, nil
	}

	err := c.backend.Lock(c.user)
	if err != nil {
		c.printer.Err("Server was unable to lock maildrop")
		return 0, fmt.Errorf("Error locking maildrop for user %s: %v", c.user, err)
	}

	c.printer.Ok("User Successfully Logged on")

	return STATE_TRANSACTION, nil
}

type StatCommand struct{}

func (cmd StatCommand) Run(c *Client, args []string) (int, error) {
	if c.currentState != STATE_TRANSACTION {
		return 0, ErrInvalidState
	}

	messages, octets, err := c.backend.Stat(c.user)
	if err != nil {
		return 0, fmt.Errorf("Error calling Stat for user %s: %v", c.user, err)
	}
	c.printer.Ok("%d %d", messages, octets)
	return STATE_TRANSACTION, nil
}

type ListCommand struct{}

func (cmd ListCommand) Run(c *Client, args []string) (int, error) {
	if c.currentState != STATE_TRANSACTION {
		return 0, ErrInvalidState
	}

	if len(args) > 0 {
		msgId, err := strconv.Atoi(args[0])
		if err != nil {
			c.printer.Err("Invalid argument: %s", args[0])
			return 0, fmt.Errorf("Invalid argument for LIST given by user %s: %v", c.user, err)
		}
		exists, octets, err := c.backend.ListMessage(c.user, msgId)
		if err != nil {
			return 0, fmt.Errorf("Error calling 'LIST %d' for user %s: %v", msgId, c.user, err)
		}
		if !exists {
			c.printer.Err("no such message")
			return STATE_TRANSACTION, nil
		}
		c.printer.Ok("%d %d", msgId, octets)
	} else {
		octets, err := c.backend.List(c.user)
		if err != nil {
			return 0, fmt.Errorf("Error calling LIST for user %s: %v", c.user, err)
		}
		c.printer.Ok("%d messages", len(octets))
		messagesList := make([]string, len(octets))
		for i, octet := range octets {
			messagesList[i] = fmt.Sprintf("%d %d", i, octet)
		}
		c.printer.MultiLine(messagesList)
	}

	return STATE_TRANSACTION, nil
}

type RetrCommand struct{}

func (cmd RetrCommand) Run(c *Client, args []string) (int, error) {
	if c.currentState != STATE_TRANSACTION {
		return 0, ErrInvalidState
	}
	if len(args) == 0 {
		c.printer.Err("Missing argument for RETR command")
		return 0, fmt.Errorf("Missing argument for RETR called by user %s", c.user)
	}

	msgId, err := strconv.Atoi(args[0])
	if err != nil {
		c.printer.Err("Invalid argument: %s", args[0])
		return 0, fmt.Errorf("Invalid argument for RETR given by user %s: %v", c.user, err)
	}

	message, err := c.backend.Retr(c.user, msgId)
	if err != nil {
		return 0, fmt.Errorf("Error calling 'RETR %d' for user %s: %v", msgId, c.user, err)
	}
	lines := strings.Split(message, "\r\n")
	c.printer.Ok("")
	c.printer.MultiLine(lines)
	return STATE_TRANSACTION, nil
}

type DeleCommand struct{}

func (cmd DeleCommand) Run(c *Client, args []string) (int, error) {
	if c.currentState != STATE_TRANSACTION {
		return 0, ErrInvalidState
	}
	if len(args) == 0 {
		c.printer.Err("Missing argument for DELE command")
		return 0, fmt.Errorf("Missing argument for DELE called by user %s", c.user)
	}

	msgId, err := strconv.Atoi(args[0])
	if err != nil {
		c.printer.Err("Invalid argument: %s", args[0])
		return 0, fmt.Errorf("Invalid argument for DELE given by user %s: %v", c.user, err)
	}
	err = c.backend.Dele(c.user, msgId)
	if err != nil {
		return 0, fmt.Errorf("Error calling 'DELE %d' for user %s: %v", msgId, c.user, err)
	}

	c.printer.Ok("Message %d deleted", msgId)

	return STATE_TRANSACTION, nil
}

type NoopCommand struct{}

func (cmd NoopCommand) Run(c *Client, args []string) (int, error) {
	if c.currentState != STATE_TRANSACTION {
		return 0, ErrInvalidState
	}
	c.printer.Ok("")
	return STATE_TRANSACTION, nil
}

type RsetCommand struct{}

func (cmd RsetCommand) Run(c *Client, args []string) (int, error) {
	if c.currentState != STATE_TRANSACTION {
		return 0, ErrInvalidState
	}
	err := c.backend.Rset(c.user)
	if err != nil {
		return 0, fmt.Errorf("Error calling 'RSET' for user %s: %v", c.user, err)
	}

	c.printer.Ok("")

	return STATE_TRANSACTION, nil
}

type UidlCommand struct{}

func (cmd UidlCommand) Run(c *Client, args []string) (int, error) {
	if c.currentState != STATE_TRANSACTION {
		return 0, ErrInvalidState
	}

	if len(args) > 0 {
		msgId, err := strconv.Atoi(args[0])
		if err != nil {
			c.printer.Err("Invalid argument: %s", args[0])
			return 0, fmt.Errorf("Invalid argument for UIDL given by user %s: %v", c.user, err)
		}
		exists, uid, err := c.backend.UidlMessage(c.user, msgId)
		if err != nil {
			return 0, fmt.Errorf("Error calling 'UIDL %d' for user %s: %v", msgId, c.user, err)
		}
		if !exists {
			c.printer.Err("no such message")
			return STATE_TRANSACTION, nil
		}
		c.printer.Ok("%d %s", msgId, uid)
	} else {
		uids, err := c.backend.Uidl(c.user)
		if err != nil {
			return 0, fmt.Errorf("Error calling UIDL for user %s: %v", c.user, err)
		}
		c.printer.Ok("%d messages", len(uids))
		uidsList := make([]string, len(uids))
		for i, uid := range uids {
			uidsList[i] = fmt.Sprintf("%d %s", i, uid)
		}
		c.printer.MultiLine(uidsList)
	}

	return STATE_TRANSACTION, nil
}

type CapaCommand struct{}

func (cmd CapaCommand) Run(c *Client, args []string) (int, error) {
	c.printer.Ok("")
	var commands []string
	commands = []string{"USER", "UIDL"}

	c.printer.MultiLine(commands)

	return c.currentState, nil
}
