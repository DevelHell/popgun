package popgun

import (
	"io/ioutil"
	"net"
	"testing"
)

func TestClient_handle(t *testing.T) {
	//TODO

	//backend := backends.DummyBackend{}
	//authorizator := backends.DummyAuthorizator{}
	//client := newClient(authorizator, backend)
	//
	//go func() {
	//	conn, err := net.Dial("tcp", ":3000")
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	defer conn.Close()
	//
	//	//invalid command
	//	fmt.Fprintf(conn, "INVALID")
	//
	//	buf, err := ioutil.ReadAll(conn)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	fmt.Println(string(buf[:]))
	//	//error executing command
	//	//successful command
	//}()
	//
	//l, err := net.Listen("tcp", ":3000")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//defer l.Close()
	//
	//conn, err := l.Accept()
	//if err != nil {
	//	t.Fatal(err)
	//}
	//defer conn.Close()
	//
	//client.handle(conn)
}

func TestClient_parseInput(t *testing.T) {

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
