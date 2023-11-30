package logging

import (
	"context"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/monetrapp/rest-api/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"strings"
)

var (
	_ pg.QueryHook = &PostgresHooks{}
)

type PostgresHooks struct {
	log   *logrus.Entry
	stats *metrics.Stats
}

func NewPostgresHooks(log *logrus.Entry, stats *metrics.Stats) pg.QueryHook {
	return &PostgresHooks{
		log:   log,
		stats: stats,
	}
}

func (h *PostgresHooks) BeforeQuery(ctx context.Context, event *pg.QueryEvent) (context.Context, error) {
	query, err := event.FormattedQuery()
	if err != nil {
		return ctx, nil
	}
	if strings.TrimSpace(strings.ToLower(string(query))) != "select 1" {
		h.log.Trace(strings.TrimSpace(string(query)))
	}

	return ctx, nil
}

func (h *PostgresHooks) AfterQuery(ctx context.Context, event *pg.QueryEvent) error {
	var queryType string
	switch query := event.Query.(type) {
	case string:
		query = strings.TrimSpace(query)
		query = strings.ReplaceAll(query, "\n", " ")

		// Don't do anything with health check queries.
		if strings.ToLower(query) == "select 1" {
			return nil
		}

		switch strings.ToUpper(query) {
		case "BEGIN", "COMMIT", "ROLLBACK":
			// Do nothing we don't want to count these.
			return nil
		default:
			firstSpace := strings.IndexRune(query, ' ')
			queryType = strings.ToUpper(query[:firstSpace])
		}
	case *orm.SelectQuery:
		queryType = "SELECT"
	case *orm.InsertQuery:
		queryType = "INSERT"
	case *orm.UpdateQuery:
		queryType = "UPDATE"
	case *orm.DeleteQuery:
		queryType = "DELETE"
	default:
		queryType = "UNKNOWN"
	}
	h.stats.Queries.With(prometheus.Labels{
		"stmt": queryType,
	}).Inc()

	//unformattedQuery, err := event.UnformattedQuery()
	//if err == nil && len(unformattedQuery) > 0 {
	//	span := sentry.StartSpan(ctx, queryType, sentry.TransactionName(string(unformattedQuery)))
	//	span.StartTime = event.StartTime
	//	span.Data = map[string]interface{}{
	//		"query": string(unformattedQuery),
	//	}
	//	defer span.Finish()
	//}

	return nil
}
