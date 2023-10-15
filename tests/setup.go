package tests

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/docker/go-connections/nat"
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
	ServerAddressEnv       = "SERVER_ADDRESS"

	TransferGasLimit = 21000
)

type applicationContainer struct {
	testcontainers.Container

	URI         string
	close       func(ctx context.Context) error
	testAccount *EthereumGateway
}

type environment struct {
	testEthereumRPCUrl  string
	testAccountPrivate  string
	applicationMnemonic string
	serverAddress       string
}

func setupTestEnvironment(ctx context.Context) (*applicationContainer, error) {
	env, err := readEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to read environment: %w", err)
	}

	testAccount, err := NewEthereumGateway(ctx, env.testEthereumRPCUrl, env.testAccountPrivate)
	if err != nil {
		return nil, fmt.Errorf("failed to create test account: %w", err)
	}

	containerCreationRequest, err := getGenericContainerRequest(
		env.serverAddress,
		env.testEthereumRPCUrl,
		env.applicationMnemonic,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get container creation request: %w", err)
	}

	container, err := testcontainers.GenericContainer(ctx, containerCreationRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	containerURL, err := getContainerURL(ctx, container, env.serverAddress)
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

	serverAddress, ok := os.LookupEnv(ServerAddressEnv)
	if !ok {
		return nil, fmt.Errorf("environment variable SERVER_ADDRESS is not set")
	}

	return &environment{
		testEthereumRPCUrl:  rpcURL,
		testAccountPrivate:  testPrivate,
		applicationMnemonic: appMnemonic,
		serverAddress:       serverAddress,
	}, nil
}

func getGenericContainerRequest(
	serverAddress, testRPCUrl, appMnemonic string,
) (testcontainers.GenericContainerRequest, error) {
	_, port, err := net.SplitHostPort(serverAddress)
	if err != nil {
		return testcontainers.GenericContainerRequest{}, fmt.Errorf("failed to split server address: %w", err)
	}

	return testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:       DockerContext,
				Dockerfile:    Dockerfile,
				PrintBuildLog: true,
			},
			ExposedPorts: []string{port},
			Env: map[string]string{
				"ETHEREUM_RPC":   testRPCUrl,
				"MNEMONIC":       appMnemonic,
				"SERVER_ADDRESS": serverAddress,
			},
			WaitingFor: wait.ForLog("server started"),
			Name:       ContainerName,
		},
		Started: true,
		Reuse:   true,
	}, nil
}

func getContainerURL(
	ctx context.Context,
	container testcontainers.Container,
	serverAddress string,
) (string, error) {
	_, rawServerPort, err := net.SplitHostPort(serverAddress)
	if err != nil {
		return "", fmt.Errorf("failed to split server address: %w", err)
	}

	serverPort := nat.Port(rawServerPort)

	host, err := container.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, serverPort)
	if err != nil {
		return "", fmt.Errorf("failed to get container port: %w", err)
	}

	addr := net.JoinHostPort(host, port.Port())

	return fmt.Sprintf("http://%s", addr), nil
}
