package ids

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_Returns32CharHex(t *testing.T) {
	t.Parallel()

	id, err := New()
	assert.NoError(t, err)
	assert.Len(t, id, 32)

	_, decErr := hex.DecodeString(id)
	assert.NoError(t, decErr)
}

func TestNew_Unique(t *testing.T) {
	t.Parallel()

	seen := map[string]struct{}{}
	for i := 0; i < 100; i++ {
		id, err := New()
		assert.NoError(t, err)
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate id generated: %s", id)
		}
		seen[id] = struct{}{}
	}
}

