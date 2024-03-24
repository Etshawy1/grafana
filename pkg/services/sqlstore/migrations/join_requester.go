package migrations

import (
	. "github.com/grafana/grafana/pkg/services/sqlstore/migrator"
)

func addJoinRequesterMigrations(mg *Migrator) {
	JoinRequester := Table{
		Name: "join_requester",
		Columns: []*Column{
			{Name: "id", Type: DB_BigInt, IsPrimaryKey: true, IsAutoIncrement: true},
			{Name: "org_id", Type: DB_BigInt, Nullable: false},
			{Name: "email", Type: DB_NVarchar, Length: 190},
			{Name: "role", Type: DB_NVarchar, Length: 20, Nullable: false},
			{Name: "justification", Type: DB_Text, Nullable: false},
			{Name: "created", Type: DB_DateTime},
			{Name: "updated", Type: DB_DateTime},
		},
		Indices: []*Index{
			{Cols: []string{"email"}, Type: IndexType},
			{Cols: []string{"org_id"}, Type: IndexType},
		},
	}

	// create table
	mg.AddMigration("create join requester table", NewAddTableMigration(JoinRequester))
	addTableIndicesMigrations(mg, "v1", JoinRequester)
}
