package driver

import (
	"database/sql"

	"github.com/La-Perla-App/backend-core-laperla/pkg/db/postgres/sqlhooks"
	"github.com/La-Perla-App/backend-core-laperla/pkg/db/postgres/sqlhooks/hooks"
	"github.com/lib/pq"
)

var CoreSQLDriver = "laperla-psql-driver"

func init() {
	sql.Register(CoreSQLDriver, sqlhooks.Wrap(&pq.Driver{}, getHooks()))
}

func getHooks() sqlhooks.Hooks {
	return sqlhooks.Compose(
		hooks.NewLogHook(),
	)
}
