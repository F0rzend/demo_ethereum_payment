package tests

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
)

const (
	DockerContext = "../"
	Dockerfile    = "Dockerfile"
	ContainerName = "testcontainers_app"

	TestPrivateKeyEnv = "TEST_PRIVATE_KEY"
	TestRPCURLEnv     = "TEST_RPC_URL"

	TransferGasLimit = 21000
)

type applicationContainer struct {
	testcontainers.Container

	URI         string
	close       func(ctx context.Context) error
	testAccount *Account
}

func setupTestEnvironment(ctx context.Context) (*applicationContainer, error) {
	rpcURL, ok := os.LookupEnv(TestRPCURLEnv)
	if !ok {
		return nil, fmt.Errorf("environment variable %s is not set", TestRPCURLEnv)
	}

	rawPrivate, ok := os.LookupEnv(TestPrivateKeyEnv)
	if !ok {
		return nil, fmt.Errorf("environment variable %s is not set", TestPrivateKeyEnv)
	}

	testAccount, err := NewAccount(rpcURL, rawPrivate)
	if err != nil {
		return nil, err
	}

	appMnemonic, ok := os.LookupEnv("MNEMONIC")
	if !ok {
		return nil, fmt.Errorf("environment variable %s is not set", "MNEMONIC")
	}

	container, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				FromDockerfile: testcontainers.FromDockerfile{
					Context:       DockerContext,
					Dockerfile:    Dockerfile,
					PrintBuildLog: true,
				},
				ExposedPorts: []string{"8080"},
				Env: map[string]string{
					"ETHEREUM_RPC": rpcURL,
					"MNEMONIC":     appMnemonic,
				},
				WaitingFor: wait.ForLog("run app"),
				Name:       ContainerName,
			},
			Started: true,
			Reuse:   true,
		},
	)
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := container.MappedPort(ctx, "8080")
	if err != nil {
		return nil, err
	}

	return &applicationContainer{
		Container:   container,
		URI:         fmt.Sprintf("http://%s:%s", host, port.Port()),
		close:       container.Terminate,
		testAccount: testAccount,
	}, nil
}
