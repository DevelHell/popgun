package popgun

type Executable interface {
	Run(c *Client, args []string) (int, error)
}

type QuitCommand struct{}

func (cmd QuitCommand) Run(c *Client, args []string) (int, error) {
	c.isAlive = false
	c.printer.Ok("Goodbye")
	return STATE_UPDATE, nil
}

type UserCommand struct{}

func (cmd UserCommand) Run(c *Client, args []string) (int, error) {
	if c.currentState != STATE_AUTHORIZATION {
		return 0, ErrInvalidState
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
	c.pass = args[0]
	if !c.authorizator.Authorize(c.user, c.pass) {
		c.printer.Err("Invalid username or password")
		return STATE_AUTHORIZATION, nil
	}
	c.printer.Ok("User Successfully Logged on")
	return STATE_TRANSACTION, nil
}

type ListCommand struct{}

func (cmd ListCommand) Run(c *Client, args []string) (int, error) {
	return 0, nil
}

type RetrCommand struct{}

func (cmd RetrCommand) Run(c *Client, args []string) (int, error) {
	return 0, nil
}

type DeleCommand struct{}

func (cmd DeleCommand) Run(c *Client, args []string) (int, error) {
	return 0, nil
}

type NoopCommand struct{}

func (cmd NoopCommand) Run(c *Client, args []string) (int, error) {
	return 0, nil
}

type RsetCommand struct{}

func (cmd RsetCommand) Run(c *Client, args []string) (int, error) {
	return 0, nil
}
