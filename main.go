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
		return "", err
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
			// TODO: handle this error better
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

	if len(r) == 0 {
		return nil, nil
	}

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
	// params can be: [field, value, field, value, ...] or map[string]string
	cmd := []string{key}

	switch param := params[0].(type) {
	case string:
		if len(params)%2 == 1 {
			return []byte{}, errors.New("missing values in input for hSet command")
		}
		for _, p := range params {
			str, ok := p.(string)
			if !ok {
				return []byte{}, errors.New("invalid format")
			}
			cmd = append(cmd, str)
		}
		return []byte{}, nil
	case map[string]string:
		for k, v := range param {
			cmd = append(cmd, k, v)
		}
	default:
		return []byte{}, errors.New("invalid format")
	}

	return ms.handleRequest(buildRESPCommand(cmd...))
}

func (ms *memoryServer) hGet(key, field string) ([]byte, error) {
	if key == "" || field == "" {
		return nil, errors.New("missing key or field for hGet command")
	}
	return ms.handleRequest(buildRESPCommand("HGET", key, field))
}

func (ms *memoryServer) hGetAll(key string, fields ...string) ([]byte, error) {
	if key == "" || len(fields) == 0 {
		return nil, errors.New("missing key or fields for hGetAll command")
	}
	cmd := []string{"HGET", key}

	for _, v := range fields {
		cmd = append(cmd, v)
	}
	return ms.handleRequest(buildRESPCommand(cmd...))
}

func (ms *memoryServer) hDel(key string, fields ...string) ([]byte, error) {
	if key == "" || len(fields) == 0 {
		return nil, errors.New("missing key or fields for hDel command")
	}
	cmd := []string{"HDEL", key}

	for _, v := range fields {
		cmd = append(cmd, v)
	}
	return ms.handleRequest(buildRESPCommand(cmd...))
}
