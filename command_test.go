package popgun

import (
	"io/ioutil"
	"net"
	"regexp"
	"testing"

	"github.com/DevelHell/popgun/backends"
)

type cmdTestCase struct {
	cmd            Executable
	initialState   int
	args           []string
	expectedState  int
	expectedErr    bool
	expectedOutput string
}

func commandTest(t *testing.T, tc cmdTestCase) {
	s, c := net.Pipe()
	defer c.Close()

	go func(t *testing.T) {
		backend := backends.DummyBackend{}
		authorizator := backends.DummyAuthorizator{}
		client := newClient(authorizator, backend)
		client.currentState = tc.initialState

		client.printer = NewPrinter(s)
		state, err := tc.cmd.Run(client, tc.args)
		if state != tc.expectedState {
			t.Errorf("Expected state '%d', but got '%d'", tc.expectedState, state)
		}
		if tc.expectedErr && err == nil {
			t.Error("Expected error, but got none")
		} else if !tc.expectedErr && err != nil {
			t.Error("Error not expected, but got one")
		}
		s.Close()
	}(t)

	buf, err := ioutil.ReadAll(c)
	if err != nil {
		t.Fatal(err)
	}
	response := string(buf[:])
	matched, err := regexp.MatchString(tc.expectedOutput, response)
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Errorf("Expected to match '%s', but got '%s'", tc.expectedOutput, response)
	}
}

func TestQuitCommand_Run(t *testing.T) {
	testCases := []cmdTestCase{
		{
			cmd:            QuitCommand{},
			initialState:   STATE_AUTHORIZATION,
			args:           []string{},
			expectedState:  STATE_AUTHORIZATION,
			expectedErr:    false,
			expectedOutput: "^\\+OK Goodbye",
		},
		{
			cmd:            QuitCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{},
			expectedState:  STATE_UPDATE,
			expectedErr:    false,
			expectedOutput: "^\\+OK Goodbye",
		},
	}

	for _, testCase := range testCases {
		commandTest(t, testCase)
	}
}

func TestUserCommand_Run(t *testing.T) {
	testCases := []cmdTestCase{
		{
			cmd:            UserCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{},
			expectedState:  0,
			expectedErr:    true,
			expectedOutput: "",
		},
		{
			cmd:            UserCommand{},
			initialState:   STATE_AUTHORIZATION,
			args:           []string{},
			expectedState:  0,
			expectedErr:    true,
			expectedOutput: "",
		},
		{
			cmd:            UserCommand{},
			initialState:   STATE_AUTHORIZATION,
			args:           []string{"john"},
			expectedState:  STATE_AUTHORIZATION,
			expectedErr:    false,
			expectedOutput: "^\\+OK",
		},
	}

	for _, testCase := range testCases {
		commandTest(t, testCase)
	}
}

func TestPassCommand_Run(t *testing.T) {
	testCases := []cmdTestCase{
		{
			cmd:            PassCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{},
			expectedState:  0,
			expectedErr:    true,
			expectedOutput: "",
		},
		{ // USER was not called before
			cmd:            PassCommand{},
			initialState:   STATE_AUTHORIZATION,
			args:           []string{},
			expectedState:  STATE_AUTHORIZATION,
			expectedErr:    false,
			expectedOutput: "",
		},
	}

	for _, testCase := range testCases {
		commandTest(t, testCase)
	}
}

func TestStatCommand_Run(t *testing.T) {
	testCases := []cmdTestCase{
		{
			cmd:            StatCommand{},
			initialState:   STATE_AUTHORIZATION,
			args:           []string{},
			expectedState:  0,
			expectedErr:    true,
			expectedOutput: "",
		},
		{
			cmd:            StatCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{},
			expectedState:  STATE_TRANSACTION,
			expectedErr:    false,
			expectedOutput: "^\\+OK 5 50",
		},
	}

	for _, testCase := range testCases {
		commandTest(t, testCase)
	}
}

func TestListCommand_Run(t *testing.T) {
	testCases := []cmdTestCase{
		{
			cmd:            ListCommand{},
			initialState:   STATE_AUTHORIZATION,
			args:           []string{},
			expectedState:  0,
			expectedErr:    true,
			expectedOutput: "",
		},
		{
			cmd:            ListCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{"a"},
			expectedState:  0,
			expectedErr:    true,
			expectedOutput: "^\\-ERR Invalid argument: a",
		},
		{
			cmd:            ListCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{"1"},
			expectedState:  STATE_TRANSACTION,
			expectedErr:    false,
			expectedOutput: "^\\+OK 1 10",
		},
		{
			cmd:            ListCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{"6"},
			expectedState:  STATE_TRANSACTION,
			expectedErr:    false,
			expectedOutput: "^\\-ERR no such message",
		},
		{
			cmd:            ListCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{},
			expectedState:  STATE_TRANSACTION,
			expectedErr:    false,
			expectedOutput: "^\\+OK 5 messages\r\n0 10\r\n1 10\r\n2 10\r\n3 10\r\n4 10\r\n\\.",
		},
	}

	for _, testCase := range testCases {
		commandTest(t, testCase)
	}
}

func TestRetrCommand_Run(t *testing.T) {
	testCases := []cmdTestCase{
		{
			cmd:            RetrCommand{},
			initialState:   STATE_AUTHORIZATION,
			args:           []string{},
			expectedState:  0,
			expectedErr:    true,
			expectedOutput: "",
		},
		{
			cmd:            RetrCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{},
			expectedState:  0,
			expectedErr:    true,
			expectedOutput: "^\\-ERR Missing argument for RETR command",
		},
		{
			cmd:            RetrCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{"a"},
			expectedState:  0,
			expectedErr:    true,
			expectedOutput: "^\\-ERR Invalid argument: a",
		},
		{
			cmd:            RetrCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{"1"},
			expectedState:  STATE_TRANSACTION,
			expectedErr:    false,
			expectedOutput: "^\\+OK \r\nthis is dummy message\r\n\\.",
		},
	}

	for _, testCase := range testCases {
		commandTest(t, testCase)
	}
}

func TestDeleCommand_Run(t *testing.T) {
	testCases := []cmdTestCase{
		{
			cmd:            DeleCommand{},
			initialState:   STATE_AUTHORIZATION,
			args:           []string{},
			expectedState:  0,
			expectedErr:    true,
			expectedOutput: "",
		},
		{
			cmd:            DeleCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{},
			expectedState:  0,
			expectedErr:    true,
			expectedOutput: "^\\-ERR Missing argument for DELE command",
		},
		{
			cmd:            DeleCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{"foo"},
			expectedState:  0,
			expectedErr:    true,
			expectedOutput: "^\\-ERR Invalid argument: foo",
		},
		{
			cmd:            DeleCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{"1"},
			expectedState:  STATE_TRANSACTION,
			expectedErr:    false,
			expectedOutput: "^\\+OK Message 1 deleted",
		},
	}

	for _, testCase := range testCases {
		commandTest(t, testCase)
	}
}

func TestNoopCommand_Run(t *testing.T) {
	testCases := []cmdTestCase{
		{
			cmd:            NoopCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{},
			expectedState:  STATE_TRANSACTION,
			expectedErr:    false,
			expectedOutput: "^\\+OK",
		},
		{
			cmd:            NoopCommand{},
			initialState:   STATE_AUTHORIZATION,
			args:           []string{},
			expectedState:  0,
			expectedErr:    true,
			expectedOutput: "",
		},
	}

	for _, testCase := range testCases {
		commandTest(t, testCase)
	}
}

func TestRsetCommand_Run(t *testing.T) {
	testCases := []cmdTestCase{
		{
			cmd:            RsetCommand{},
			initialState:   STATE_AUTHORIZATION,
			args:           []string{},
			expectedState:  0,
			expectedErr:    true,
			expectedOutput: "",
		},
		{
			cmd:            RsetCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{},
			expectedState:  STATE_TRANSACTION,
			expectedErr:    false,
			expectedOutput: "^\\+OK ",
		},
	}

	for _, testCase := range testCases {
		commandTest(t, testCase)
	}
}

func TestUidlCommand_Run(t *testing.T) {
	testCases := []cmdTestCase{
		{
			cmd:            UidlCommand{},
			initialState:   STATE_AUTHORIZATION,
			args:           []string{},
			expectedState:  0,
			expectedErr:    true,
			expectedOutput: "",
		},
		{
			cmd:            UidlCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{"a"},
			expectedState:  0,
			expectedErr:    true,
			expectedOutput: "^\\-ERR Invalid argument: a",
		},
		{
			cmd:            UidlCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{"6"},
			expectedState:  STATE_TRANSACTION,
			expectedErr:    false,
			expectedOutput: "^\\-ERR no such message",
		},
		{
			cmd:            UidlCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{"1"},
			expectedState:  STATE_TRANSACTION,
			expectedErr:    false,
			expectedOutput: "^\\+OK 1 2",
		},
		{
			cmd:            UidlCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{},
			expectedState:  STATE_TRANSACTION,
			expectedErr:    false,
			expectedOutput: "^\\+OK 5 messages\r\n0 1\r\n1 2\r\n2 3\r\n3 4\r\n4 5\r\n\\.",
		},
	}

	for _, testCase := range testCases {
		commandTest(t, testCase)
	}
}

func TestCapaCommand_Run(t *testing.T) {
	testCases := []cmdTestCase{
		{
			cmd:            CapaCommand{},
			initialState:   STATE_TRANSACTION,
			args:           []string{},
			expectedState:  STATE_TRANSACTION,
			expectedErr:    false,
			expectedOutput: "^\\+OK \r\nUSER\r\nUIDL\r\n\\.",
		},
		{
			cmd:            CapaCommand{},
			initialState:   STATE_AUTHORIZATION,
			args:           []string{},
			expectedState:  STATE_AUTHORIZATION,
			expectedErr:    false,
			expectedOutput: "^\\+OK \r\nUSER\r\nUIDL\r\n\\.",
		},
	}

	for _, testCase := range testCases {
		commandTest(t, testCase)
	}
}
