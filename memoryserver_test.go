package memoryserver_test

import (
	"testing"

	memoryserver "github.com/julianblanco00/memory-server-client"
)

func TestMemoryServer(t *testing.T) {
	ms := memoryserver.New("localhost", "4444")
	err := ms.Connect()
	if err != nil {
		t.Error("error connecting to remote server")
		return
	}

	t.Run("get empty key", func(t *testing.T) {
		v, err := ms.Get("mykey")
		if err != nil {
			t.Error("error getting mykey")
		}

		if len(v) > 0 {
			t.Error("value for mykey exists")
		}
	})
}
