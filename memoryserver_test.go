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

	t.Run("set key", func(t *testing.T) {
		v, err := ms.Set("mykey", "myvalue")
		if err != nil {
			t.Error("error setting mykey")
		}
		if string(v) != "OK" {
			t.Error("error setting mykey")
		}
	})

	t.Run("key exists", func(t *testing.T) {
		v, err := ms.Exists("mykey")
		if err != nil {
			t.Error("error setting mykey")
		}
		if string(v) != "1" {
			t.Error("mykey should exist")
		}
	})

	t.Run("get non-empty key", func(t *testing.T) {
		v, err := ms.Get("mykey")
		if err != nil {
			t.Error("error getting mykey")
		}

		if v == nil {
			t.Error("value for mykey does not exists")
		}

		if string(v) != "myvalue" {
			t.Error("value set before does not match current value")
		}
	})
}
