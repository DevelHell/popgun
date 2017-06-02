package backends

import "fmt"

// DummyAuthorizator is a fake authorizator interface implementation used for test
type DummyAuthorizator struct {
}

func (a DummyAuthorizator) Authorize(user, pass string) bool {
	return true
}

// DummyBackend is a fake backend interface implementation used for test
type DummyBackend struct {
}

func (b DummyBackend) Stat(user string) (messages, octets int, err error) {
	return 5, 50, nil
}
func (b DummyBackend) List(user string) (octets []int, err error) {
	return []int{10, 10, 10, 10, 10}, nil
}
func (b DummyBackend) ListMessage(user string, msgId int) (exists bool, octets int, err error) {
	if msgId > 4 {
		return false, 0, nil
	}
	return true, 10, nil
}
func (b DummyBackend) Retr(user string, msgId int) (message string, err error) {
	return "this is dummy message", nil
}
func (b DummyBackend) Dele(user string, msgId int) error {
	return nil
}
func (b DummyBackend) Rset(user string) error {
	return nil
}
func (b DummyBackend) Uidl(user string) (uids []string, err error) {
	return []string{"1", "2", "3", "4", "5"}, nil
}
func (b DummyBackend) UidlMessage(user string, msgId int) (exists bool, uid string, err error) {
	if msgId > 4 {
		return false, "", nil
	}
	return true, fmt.Sprintf("%d", msgId+1), nil
}
func (b DummyBackend) Update(user string) error {
	return nil
}
func (b DummyBackend) Lock(user string) error {
	return nil
}
func (b DummyBackend) Unlock(user string) error {
	return nil
}
