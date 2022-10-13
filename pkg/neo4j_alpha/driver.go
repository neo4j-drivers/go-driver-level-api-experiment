package neo4j_alpha

import (
	"context"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"net/url"
	"strings"
)

func NewDriver(target string, auth neo4j.AuthToken, configurers ...func(*neo4j.Config)) (Driver, error) {
	delegate, err := neo4j.NewDriverWithContext(target, auth, configurers...)
	if err != nil {
		return nil, err
	}
	return &driver{delegate: delegate}, nil
}

type Driver interface {
	// ExecuteQuery runs the specified query and optional parameters in a retryable transaction function
	// The query will be re-executed until it is successful or a number of attempts has been reached
	ExecuteQuery(ctx context.Context, query string, params map[string]any, settings ...QueryConfigOption) (*EagerResult, error)

	// Target returns the url this driver is bootstrapped
	Target() url.URL
	// NewSession creates a new session based on the specified session configuration.
	NewSession(ctx context.Context, config neo4j.SessionConfig) neo4j.SessionWithContext
	// VerifyConnectivity checks that the driver can connect to a remote server or cluster by
	// establishing a network connection with the remote. Returns nil if succesful
	// or error describing the problem.
	VerifyConnectivity(ctx context.Context) error
	// Close the driver and all underlying connections
	Close(ctx context.Context) error
	// IsEncrypted determines whether the driver communication with the server
	// is encrypted. This is a static check. The function can also be called on
	// a closed Driver.
	IsEncrypted() bool
}

type driver struct {
	delegate neo4j.DriverWithContext
}

func (d *driver) ExecuteQuery(ctx context.Context, query string, params map[string]any, options ...QueryConfigOption) (_ *EagerResult, err error) {
	queryConfig := QueryConfig{}
	for _, option := range options {
		option(&queryConfig)
	}
	session := d.delegate.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName:     queryConfig.Database,
		ImpersonatedUser: queryConfig.ImpersonatedUser,
		BookmarkManager:  queryConfig.BookmarkManager,
	})
	defer func() {
		err = session.Close(ctx)
	}()
	txFuncApi := queryConfig.RoutingControl.resolveTxFuncApi(session)
	result, err := txFuncApi(ctx, RunQuery(ctx, query, params))
	if err != nil {
		return nil, err
	}
	return result.(*EagerResult), nil
}

func (d *driver) Target() url.URL {
	return d.delegate.Target()
}

func (d *driver) NewSession(ctx context.Context, config neo4j.SessionConfig) neo4j.SessionWithContext {
	return d.delegate.NewSession(ctx, config)
}

func (d *driver) VerifyConnectivity(ctx context.Context) error {
	return d.delegate.VerifyConnectivity(ctx)
}

func (d *driver) Close(ctx context.Context) error {
	return d.delegate.Close(ctx)
}

func (d *driver) IsEncrypted() bool {
	return d.delegate.IsEncrypted()
}

type QueryConfig struct {
	RoutingControl   RoutingControl
	Database         string
	ImpersonatedUser string
	BookmarkManager  neo4j.BookmarkManager
}

type EagerResult struct {
	Keys    []string
	Records []*neo4j.Record
	Summary neo4j.ResultSummary
}

func (e *EagerResult) String() string {
	return fmt.Sprintf("keys: %v, records: %s, summary: %s", e.Keys, stringifyRecords(e.Records), stringifySummary(e.Summary))
}

func stringifyRecords(records []*neo4j.Record) string {
	builder := strings.Builder{}
	for _, record := range records {
		builder.WriteString("{")
		for i, key := range record.Keys {
			value, found := record.Get(key)
			if !found {
				value = "<N/A>"
			}
			separator := ","
			if i == len(record.Keys)-1 {
				separator = ""
			}
			builder.WriteString(fmt.Sprintf("%q: %v%s", key, value, separator))
		}
		builder.WriteString("}")
	}
	return builder.String()
}

func stringifySummary(summary neo4j.ResultSummary) string {
	serverInfo := summary.Server()
	protocolVersion := serverInfo.ProtocolVersion()
	return fmt.Sprintf(
		`{"db": %q, "address": %q, "protocol_version": "%d.%d", "agent": %q}`,
		summary.Database().Name(),
		serverInfo.Address(),
		protocolVersion.Major, protocolVersion.Minor,
		serverInfo.Agent(),
	)
}

type QueryConfigOption func(*QueryConfig)

func WithReadersRoutingControl() QueryConfigOption {
	return func(config *QueryConfig) {
		config.RoutingControl = Readers
	}
}

func WithWritersRoutingControl() QueryConfigOption {
	return func(config *QueryConfig) {
		config.RoutingControl = Writers
	}
}

func WithDatabase(db string) QueryConfigOption {
	return func(config *QueryConfig) {
		config.Database = db
	}
}

func WithImpersonatedUser(user string) QueryConfigOption {
	return func(config *QueryConfig) {
		config.ImpersonatedUser = user
	}
}

func WithBookmarkManager(bookmarkManager neo4j.BookmarkManager) QueryConfigOption {
	return func(config *QueryConfig) {
		config.BookmarkManager = bookmarkManager
	}
}

type RoutingControl uint8

func (rc RoutingControl) resolveTxFuncApi(session neo4j.SessionWithContext) TransactionFunctionApi {
	switch rc {
	case Writers:
		return session.ExecuteWrite
	case Readers:
		return session.ExecuteRead
	default:
		panic(fmt.Sprintf("unknown routing control: %d", rc))
	}
}

const (
	Writers RoutingControl = iota
	Readers
)

type TransactionFunctionApi func(ctx context.Context, work neo4j.ManagedTransactionWork, configurers ...func(*neo4j.TransactionConfig)) (any, error)

func RunQuery(ctx context.Context, query string, params map[string]any) neo4j.ManagedTransactionWork {
	return func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		keys, err := result.Keys()
		if err != nil {
			return nil, err
		}
		records, err := result.Collect(ctx)
		if err != nil {
			return nil, err
		}
		summary, err := result.Consume(ctx)
		if err != nil {
			return nil, err
		}
		return &EagerResult{
			Keys:    keys,
			Records: records,
			Summary: summary,
		}, nil
	}
}
