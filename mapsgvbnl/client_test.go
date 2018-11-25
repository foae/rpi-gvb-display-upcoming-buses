package mapsgvbnl

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFromClockToTime(t *testing.T) {
	t1 := "10:31:58"
	t2 := "01:39:56"
	t3 := "23:41:38"
	t4 := "00:54:39"

	t11, err := FromClockToTime(t1)
	assert.NoError(t, err)
	assert.IsTypef(t, time.Now(), *t11, "time.Time")

	t22, err := FromClockToTime(t2)
	assert.NoError(t, err)
	assert.IsTypef(t, time.Now(), *t22, "time.Time")

	t33, err := FromClockToTime(t3)
	assert.NoError(t, err)
	assert.IsTypef(t, time.Now(), *t33, "time.Time")

	t44, err := FromClockToTime(t4)
	assert.NoError(t, err)
	assert.IsTypef(t, time.Now(), *t44, "time.Time")
}
