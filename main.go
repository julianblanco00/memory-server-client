package memoryserver

import (
	"encoding/hex"
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

func (ms *memoryServer) Get(cmd string) ([]byte, error) {
	return ms.handleRequest(buildRESPCommand("GET", cmd))
}
