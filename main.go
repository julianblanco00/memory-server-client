package memoryserver

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/google/uuid"
)

type memoryServer struct {
	client          net.Conn
	pendingRequests map[string]chan ([]byte)
	host            string
	port            string
}

func buildRESPCommand(vals ...string) string {
	str := ""

	for _, v := range vals {
		str += fmt.Sprintf("$%d\n%s\n", len(v), v)
	}

	return str
}

func buildRequestId() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", nil
	}

	clean := strings.ReplaceAll(id.String(), "-", "")

	return clean, nil
}

func New(host, port string) *memoryServer {
	return &memoryServer{
		host:            host,
		port:            port,
		pendingRequests: make(map[string]chan ([]byte)),
	}
}

func (ms *memoryServer) listenConnectionEvents() (string, error) {
	for {
		buf := make([]byte, 1024)
		n, err := ms.client.Read(buf)
		if err != nil {
			return "", err
		}

		id := hex.EncodeToString(buf[:16])
		c := ms.pendingRequests[id]
		c <- buf[16:n]
	}
}

func (ms *memoryServer) Connect() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", ms.host, ms.port))
	if err != nil {
		return err
	}

	ms.client = conn

	go ms.listenConnectionEvents()

	return nil
}

func (ms *memoryServer) handleRequest(cmd string) ([]byte, error) {
	c := make(chan []byte)
	id, err := buildRequestId()
	if err != nil {
		return []byte{}, err
	}

	ms.pendingRequests[id] = c

	hexId, err := hex.DecodeString(id)
	if err != nil {
		return []byte{}, err
	}

	command := fmt.Sprintf("%s%s", hexId, cmd)

	_, err = ms.client.Write([]byte(command))
	if err != nil {
		return []byte{}, err
	}

	r := <-c

	return r, nil
}

func (ms *memoryServer) Get(key string) ([]byte, error) {
	return ms.handleRequest(buildRESPCommand("GET", key))
}

func (ms *memoryServer) Del(key ...string) ([]byte, error) {
	keys := []string{"DEL"}

	for _, k := range key {
		keys = append(keys, k)
	}

	return ms.handleRequest(buildRESPCommand(keys...))
}

func (ms *memoryServer) Set(key, val string) ([]byte, error) {
	return ms.handleRequest(buildRESPCommand("SET", key, val))
}

func (ms *memoryServer) SetWithOpts(key, val string, opts [][]string) ([]byte, error) {
	cmd := []string{"SET", key, val}
	for _, opt := range opts {
		cmd = append(cmd, opt...)
	}

	return ms.handleRequest(buildRESPCommand(cmd...))
}

func (ms *memoryServer) mSet(params ...string) ([]byte, error) {
	if len(params)%2 == 1 {
		return []byte{}, errors.New("missing values in input for mSet command")
	}

	kvs := []string{"MSET"}

	for _, p := range params {
		kvs = append(kvs, p)
	}

	return ms.handleRequest(buildRESPCommand(kvs...))
}

func (ms *memoryServer) hSet(key string, params ...interface{}) ([]byte, error) {
	return []byte{}, nil
}
