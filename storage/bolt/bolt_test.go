package bolt

import (
	"testing"
	"time"
)

func TestTimeToKeyNoNanoseconds(t *testing.T) {
	key := timeToKey(time.Unix(0, 0))
	if len(key) != 30 {
		t.Errorf("Wrong length %d", len(key))
	}
}

func TestTimeToKeyWithNanoseconds(t *testing.T) {
	key := timeToKey(time.Unix(0, 42))
	if len(key) != 30 {
		t.Errorf("Wrong length %d", len(key))
	}
}

func BenchmarkReadTimerange(b *testing.B) {
	for i := 0; i < b.N; i++ {
		blt, err := New("/home/filip/repositories/goboreq/flightradar-backend/database.bolt")
		if err != nil {
			b.Fatal(err)
		}
		from := time.Unix(1, 0)
		to := time.Unix(1518113999, 0)
		data, err := blt.RetrieveTimerange(from, to)
		if err != nil {
			b.Fatal(err)
		}
		b.Logf("Retrieved data points: %d", len(data))
	}
}
