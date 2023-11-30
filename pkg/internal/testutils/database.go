package testutils

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/metrics"
	"github.com/monetr/monetr/pkg/migrations"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var (
	_ pg.QueryHook = &queryHook{}
)

type queryHook struct {
	log   *logrus.Entry
	stats *metrics.Stats
}

func (q *queryHook) BeforeQuery(ctx context.Context, event *pg.QueryEvent) (context.Context, error) {
	queryId := gofakeit.UUID()[0:8]
	if event.Stash != nil {
		event.Stash["queryId"] = queryId
	} else {
		event.Stash = map[interface{}]interface{}{
			"queryId": queryId,
		}
	}

	query, err := event.FormattedQuery()
	if err != nil {
		return ctx, nil
	}

	q.log.WithContext(ctx).WithField("queryId", queryId).Trace(string(query))

	return ctx, nil
}

func (q *queryHook) AfterQuery(ctx context.Context, event *pg.QueryEvent) error {
	if q.stats != nil {
		q.stats.Queries.With(prometheus.Labels{}).Inc()
	}

	if event.Err != nil {
		log := q.log.WithContext(ctx)
		if event.Stash != nil {
			if queryId, ok := event.Stash["queryId"].(string); ok {
				log = log.WithField("queryId", queryId)
			}
		}
		log.WithError(event.Err).Warn("query failed")
	}

	return nil
}

func GetPgDatabaseTxn(t *testing.T) *pg.Tx {
	db := GetPgDatabase(t)

	txn, err := db.Begin()
	require.NoError(t, err, "must begin transaction")

	t.Cleanup(func() {
		require.NoError(t, txn.Rollback(), "must rollback database transaction")
	})

	return txn
}

var testDatabases struct {
	lock      sync.Mutex
	databases map[string]*pg.DB
}

func init() {
	testDatabases = struct {
		lock      sync.Mutex
		databases map[string]*pg.DB
	}{
		lock:      sync.Mutex{},
		databases: map[string]*pg.DB{},
	}
}

func GetPgOptions(t *testing.T) *pg.Options {
	portString := os.Getenv("POSTGRES_PORT")
	if portString == "" {
		portString = "5432"
	}

	port, err := strconv.ParseInt(portString, 10, 64)
	require.NoError(t, err, "must be able to parse the Postgres port as a number")

	address := fmt.Sprintf("%s:%d", os.Getenv("POSTGRES_HOST"), port)

	options := &pg.Options{
		Network:         "tcp",
		Addr:            address,
		User:            os.Getenv("POSTGRES_USER"),
		Password:        os.Getenv("POSTGRES_PASSWORD"),
		Database:        os.Getenv("POSTGRES_DB"),
		ApplicationName: "monetr - api - tests",
	}

	return options
}

type DatabaseOption uint8

const (
	IsolatedDatabase DatabaseOption = 1
)

func GetPgDatabase(t *testing.T, databaseOptions ...DatabaseOption) *pg.DB {
	testDatabases.lock.Lock()
	defer testDatabases.lock.Unlock()

	if db, ok := testDatabases.databases[t.Name()]; ok {
		return db
	}

	options := GetPgOptions(t)
	db := pg.Connect(options)

	require.NoError(t, db.Ping(context.Background()), "must ping database")

	log := GetLog(t)

	db.AddQueryHook(&queryHook{
		log: log,
	})

	var databaseToReturn *pg.DB
	databaseToReturn = db
	if len(databaseOptions) > 0 {
		for _, option := range databaseOptions {
			switch option {
			case IsolatedDatabase:
				log.Debug("creating isolated database for test")
				databaseName := fmt.Sprintf("%x", md5.Sum([]byte(t.Name())))

				_, err := db.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS "%s";`, databaseName))
				require.NoError(t, err, "must be able to drop an isolated database if it exists")

				_, err = db.Exec(fmt.Sprintf(`CREATE DATABASE "%s";`, databaseName))
				require.NoError(t, err, "must be able to create the isolated database")

				isolatedOptions := *options
				isolatedOptions.Database = databaseName
				databaseToReturn = pg.Connect(&isolatedOptions)
				databaseToReturn.AddQueryHook(&queryHook{
					log: log,
				})

				migrations.RunMigrations(log, databaseToReturn)

				t.Cleanup(func() {
					require.NoError(t, databaseToReturn.Close(), "must close the isolated database once we are done")
					_, err := db.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS "%s";`, databaseName))
					require.NoError(t, err, "must be able to drop an isolated database if it exists")
					require.NoError(t, db.Close(), "must close database connection")
				})
			}
		}
	} else {
		t.Cleanup(func() {
			require.NoError(t, db.Close(), "must close database connection")
		})
	}

	testDatabases.databases[t.Name()] = databaseToReturn

	return databaseToReturn
}
