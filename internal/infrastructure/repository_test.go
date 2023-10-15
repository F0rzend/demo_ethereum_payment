package infrastructure_test

import (
	"math/big"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/F0rzend/demo_ethereum_payment/internal/infrastructure"
)

func TestRepository_GetID(t *testing.T) {
	t.Parallel()

	const callTimes = 1_000_000
	sut := infrastructure.NewRepository()

	callParallelAndWait(callTimes-1, func() {
		sut.GetID()
	})
	lastID := sut.GetID()

	assert.Equal(t, big.NewInt(callTimes), lastID)
}

func callParallelAndWait(times int, f func()) {
	wg := new(sync.WaitGroup)
	wg.Add(times)

	for i := 0; i < times; i++ {
		go func() {
			defer wg.Done()

			f()
		}()
	}

	wg.Wait()
}
