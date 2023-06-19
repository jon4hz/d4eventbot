package d4armory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRecentEvents(t *testing.T) {
	c := New()
	r, err := c.GetRecent(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, r)
}
