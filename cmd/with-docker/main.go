package main

import (
	"context"
	"github.com/neo4j-drivers/go-driver-level-api-experiment/pkg/container"
	"github.com/neo4j-drivers/go-driver-level-api-experiment/pkg/neo4j_alpha"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
	ctx := context.Background()
	username := "neo4j"
	password := "s3cr3t"
	server, err := container.Start(ctx, containerConfiguration(username, password))
	panicInTheErr(err)
	defer func() {
		panicInTheErr(container.Stop(ctx, server))
	}()

	driver, err := neo4j_alpha.NewDriver(container.Uri(ctx, server), basicAuth(username, password))
	panicInTheErr(err)
	defer func() {
		panicInTheErr(driver.Close(ctx))
	}()
	panicInTheErr(driver.VerifyConnectivity(ctx))

	// TODO 1:
	//  - use the above driver and run the Cypher query "RETURN 42"
	//  - call panicInTheErr on the error, if the API you call returns one
	//  - print the result with fmt.Println

	// YOUR CODE GOES HERE

	// TODO 2:
	//  - use the above driver and run the Cypher query "UNWIND [1, 2] AS value RETURN value" against database "neo4j"
	//  - call panicInTheErr on the error, if the API you call returns one
	//  - sum each record value
	//  - print the sum with fmt.Println

	// YOUR CODE GOES HERE
}

func containerConfiguration(username string, password string) container.ContainerConfiguration {
	return container.ContainerConfiguration{
		Neo4jVersion: "4.4-enterprise",
		Username:     username,
		Password:     password,
		Scheme:       "bolt",
	}
}

func basicAuth(username string, password string) neo4j.AuthToken {
	return neo4j.BasicAuth(username, password, "")
}

func panicInTheErr(err error) {
	if err != nil {
		panic(err)
	}
}
