package tests

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	DockerContext = "../"
	Dockerfile    = "Dockerfile"
	ContainerName = "testcontainers_app"

	TestPrivateKeyEnv      = "TEST_PRIVATE_KEY"
	TestRPCURLEnv          = "TEST_ETHEREUM_RPC"
	ApplicationMnemonicEnv = "MNEMONIC"

	TransferGasLimit = 21000
)

type applicationContainer struct {
	testcontainers.Container

	URI         string
	close       func(ctx context.Context) error
	testAccount *Account
}

type environment struct {
	testEthereumRPCUrl  string
	testAccountPrivate  string
	applicationMnemonic string
}

func setupTestEnvironment(ctx context.Context) (*applicationContainer, error) {
	env, err := readEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to read environment: %w", err)
	}

	testAccount, err := NewAccount(ctx, env.testEthereumRPCUrl, env.testAccountPrivate)
	if err != nil {
		return nil, fmt.Errorf("failed to create test account: %w", err)
	}

	containerCreationRequest := getGenericContainerRequest(env.testEthereumRPCUrl, env.applicationMnemonic)

	container, err := testcontainers.GenericContainer(ctx, containerCreationRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	containerURL, err := getContainerURL(ctx, container)
	if err != nil {
		return nil, fmt.Errorf("failed to get container URL: %w", err)
	}

	return &applicationContainer{
		Container:   container,
		URI:         containerURL,
		close:       container.Terminate,
		testAccount: testAccount,
	}, nil
}

func readEnv() (*environment, error) {
	rpcURL, ok := os.LookupEnv(TestRPCURLEnv)
	if !ok {
		return nil, fmt.Errorf("environment variable %s is not set", TestRPCURLEnv)
	}

	testPrivate, ok := os.LookupEnv(TestPrivateKeyEnv)
	if !ok {
		return nil, fmt.Errorf("environment variable %s is not set", TestPrivateKeyEnv)
	}

	appMnemonic, ok := os.LookupEnv(ApplicationMnemonicEnv)
	if !ok {
		return nil, fmt.Errorf("environment variable %s is not set", ApplicationMnemonicEnv)
	}

	return &environment{
		testEthereumRPCUrl:  rpcURL,
		testAccountPrivate:  testPrivate,
		applicationMnemonic: appMnemonic,
	}, nil
}

func getGenericContainerRequest(testRPCUrl, appMnemonic string) testcontainers.GenericContainerRequest {
	return testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:       DockerContext,
				Dockerfile:    Dockerfile,
				PrintBuildLog: true,
			},
			ExposedPorts: []string{"8080"},
			Env: map[string]string{
				"ETHEREUM_RPC": testRPCUrl,
				"MNEMONIC":     appMnemonic,
			},
			WaitingFor: wait.ForLog("run app"),
			Name:       ContainerName,
		},
		Started: true,
		Reuse:   true,
	}
}

func getContainerURL(ctx context.Context, container testcontainers.Container) (string, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "8080")
	if err != nil {
		return "", fmt.Errorf("failed to get container port: %w", err)
	}

	addr := net.JoinHostPort(host, port.Port())

	return fmt.Sprintf("http://%s", addr), nil
}
