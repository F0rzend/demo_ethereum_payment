package tests

import (
	"context"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite

	e             func() *httpexpect.Expect
	testAccount   *Account
	tearDownSuite func(*testing.T)
}

func (s *TestSuite) SetupSuite() {
	ctx := context.Background()

	app, err := setupTestEnvironment(ctx)
	require.NoError(s.T(), err)

	s.e = func() *httpexpect.Expect {
		return httpexpect.New(s.T(), app.URI)
	}
	s.testAccount = app.testAccount
	s.tearDownSuite = func(t *testing.T) {
		if err := app.close(ctx); err != nil {
			t.Fatal(err)
		}
	}
}

func (s *TestSuite) TearDownSuite() {
	s.tearDownSuite(s.T())
}

func TestApplication(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
