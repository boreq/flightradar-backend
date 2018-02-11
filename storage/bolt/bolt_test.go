package bolt

import (
	"testing"
	"time"
)

func TestTimeToKeyNoNanoseconds(t *testing.T) {
	key := timeToKey(time.Unix(0, 0))
	if len(key) != 30 {
		t.Error("Wrong length %d", len(key))
	}
}

func TestTimeToKeyWithNanoseconds(t *testing.T) {
	key := timeToKey(time.Unix(0, 42))
	if len(key) != 30 {
		t.Error("Wrong length %d", len(key))
	}
}
