package infrastructure

import (
	"math/big"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepository_GetID(t *testing.T) {
	t.Parallel()

	const callTimes = 1_000_000
	sut := NewRepository()

	callParallelAndWait(callTimes-1, func() {
		sut.GetID()
	})
	lastID := sut.GetID()

	assert.Equal(t, big.NewInt(callTimes), lastID)
}

func callParallelAndWait(times int, f func()) {
	wg := &sync.WaitGroup{}
	wg.Add(times)
	for i := 0; i < times; i++ {
		go func() {
			defer wg.Done()

			f()
		}()
	}
	wg.Wait()
}
