package models

import (
	"context"
	"log"

	"github.com/go-pg/pg"
	"gitlab.bbdev.team/vh/vh-srv-profile/app"
)

//DB DB link
var DB *pg.DB

//postgresDebugger
type postgresDebugger struct {
}

//BeforeQuery hook before query
func (pd postgresDebugger) BeforeQuery(c context.Context, q *pg.QueryEvent) (context.Context, error) {

	return c, nil
}

//AfterQuery hook after query
func (pd postgresDebugger) AfterQuery(c context.Context, q *pg.QueryEvent) error {
	if app.Config.AppMode == "dev" {
		sql, _ := q.FormattedQuery()

		log.Println(string(sql))
	}

	return nil
}

//OpenDBConnection open db connection
func OpenDBConnection() {

}

//CloseDBConnection close db connection
func CloseDBConnection() {

}
