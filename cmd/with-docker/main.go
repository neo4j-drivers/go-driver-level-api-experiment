package main

import (
	"context"
	"github.com/neo4j-drivers/go-driver-level-api-experiment/pkg/container"
	"github.com/neo4j-drivers/go-driver-level-api-experiment/pkg/neo4j_alpha"
	"github.com/neo4j-drivers/go-driver-level-api-experiment/pkg/todo"
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

	// go there first ;)
	todo.Todo1(driver)
	// and when done with the first, go to that one :)
	todo.Todo2(driver)
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
