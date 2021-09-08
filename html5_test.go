package golibinjection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Html5white(t *testing.T) {
	assert.Equal(t, h5_is_white(32), true)
	assert.Equal(t, h5_is_white('\r'), true)
	assert.Equal(t, h5_is_white('\n'), true)
	assert.Equal(t, h5_is_white('\t'), true)
}
