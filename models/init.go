package models

import (
	"context"
	"log"

	"github.com/go-pg/pg/v10"
	"gitlab.bbdev.team/vh/vh-srv-profile/app"
)

//DB DB link
var DB *pg.DB

//postgresDebugger
type postgresDebugger struct {
}

//NameTranslate
type NameTranslate struct {
	LangID uint64 `json:"lang_id"`
	Value  string `json:"name"`
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

	DB = pg.Connect(&pg.Options{
		Network:         app.Config.DBNetwork,
		Addr:            app.Config.DBHost,
		User:            app.Config.DBUserName,
		Password:        app.Config.DBPassword,
		Database:        app.Config.DBName,
		ApplicationName: app.Config.DBApplicationName,
	})

	DB.AddQueryHook(postgresDebugger{})

}

//CloseDBConnection close db connection
func CloseDBConnection() {

}
