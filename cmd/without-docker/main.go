package main

import (
	"context"
	"github.com/neo4j-drivers/go-driver-level-api-experiment/pkg/neo4j_alpha"
	"github.com/neo4j-drivers/go-driver-level-api-experiment/pkg/todo"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
	ctx := context.Background()
	driver, err := neo4j_alpha.NewDriver("bolt://18.206.88.187:7687", basicAuth("neo4j", "legends-wonder-bulkhead"))
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

func basicAuth(username string, password string) neo4j.AuthToken {
	return neo4j.BasicAuth(username, password, "")
}

func panicInTheErr(err error) {
	if err != nil {
		panic(err)
	}
}
