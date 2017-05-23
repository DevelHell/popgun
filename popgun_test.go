package popgun

import (
	"io/ioutil"
	"net"
	"testing"

	"fmt"

	"bufio"

	"reflect"

	"github.com/DevelHell/popgun/backends"
)

func TestClient_handle(t *testing.T) {
	backend := backends.DummyBackend{}
	authorizator := backends.DummyAuthorizator{}
	client := newClient(authorizator, backend)

	go func() {
		conn, err := net.Dial("tcp", ":3000")
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		//read welcome message
		_, err = reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		//invalid command
		expected := "-ERR Invalid command INVALID\r\n"
		fmt.Fprintf(conn, "INVALID\n")
		response, err := reader.ReadString('\n')
		if response != expected {
			t.Errorf("Expected '%s', but got '%s'", expected, response)
		}

		//error executing command - rset cannot be executed in current state
		expected = "-ERR Error executing command RSET\r\n"
		fmt.Fprintf(conn, "RSET\n")
		response, err = reader.ReadString('\n')
		if response != expected {
			t.Errorf("Expected '%s', but got '%s'", expected, response)
		}

		//successful command
		expected = "+OK Goodbye\r\n"
		fmt.Fprintf(conn, "QUIT\n")
		response, err = reader.ReadString('\n')
		if response != expected {
			t.Errorf("Expected '%s', but got '%s'", expected, response)
		}
	}()

	l, err := net.Listen("tcp", ":3000")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client.handle(conn)
}

func TestClient_parseInput(t *testing.T) {
	backend := backends.DummyBackend{}
	authorizator := backends.DummyAuthorizator{}
	client := newClient(authorizator, backend)

	tables := [][][]string{
		{{"COMMAND1"}, {"COMMAND1"}},
		{{"COMMAND1   "}, {"COMMAND1"}},
		{{"COMMAND1 \r \n "}, {"COMMAND1"}},
		{{"comm ARG"}, {"COMM", "ARG"}},
		{{"COMM arg"}, {"COMM", "arg"}},
		{{"COMM ARG1 ARG2"}, {"COMM", "ARG1", "ARG2"}},
	}
	for _, testCase := range tables {
		inputCmd := testCase[0][0]
		cmd, args := client.parseInput(inputCmd)
		expectedCmd := testCase[1][0]
		if cmd != expectedCmd {
			t.Errorf("Expected '%s', but got '%s'", expectedCmd, cmd)
		}
		expectedArgs := testCase[1][1:]
		if !reflect.DeepEqual(args, expectedArgs) {
			t.Errorf("Expected '%s', but got '%s'", expectedArgs, args)
		}
	}
}

func TestServer_Start(t *testing.T) {

}

type printerFunc func(conn net.Conn)

func printerTest(t *testing.T, f printerFunc) string {
	go func() {
		conn, err := net.Dial("tcp", ":3000")
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		f(conn)
	}()

	l, err := net.Listen("tcp", ":3000")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	buf, err := ioutil.ReadAll(conn)
	if err != nil {
		t.Fatal(err)
	}

	return string(buf[:])
}

func TestPrinter_Welcome(t *testing.T) {
	expected := "+OK POPgun POP3 server ready\r\n"

	msg := printerTest(t, func(conn net.Conn) {
		p := NewPrinter(conn)
		p.Welcome()
	})

	if msg != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, msg)
	}
}

func TestPrinter_Ok(t *testing.T) {
	expected := "+OK 2 foxes jumping over lazy dog\r\n"

	msg := printerTest(t, func(conn net.Conn) {
		p := NewPrinter(conn)
		p.Ok("%d foxes jumping over lazy dog", 2)
	})

	if msg != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, msg)
	}
}

func TestPrinter_Err(t *testing.T) {
	expected := "-ERR everything wrong in 10 seconds\r\n"

	msg := printerTest(t, func(conn net.Conn) {
		p := NewPrinter(conn)
		p.Err("everything wrong in %d seconds", 10)
	})

	if msg != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, msg)
	}
}

func TestPrinter_MultiLine(t *testing.T) {
	expected := "multi\r\nline\r\n.\r\n"

	msg := printerTest(t, func(conn net.Conn) {
		p := NewPrinter(conn)
		p.MultiLine([]string{"multi", "line"})
	})

	if msg != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, msg)
	}
}
