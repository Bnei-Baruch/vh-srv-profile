package membership

import (
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
)

type doer interface {
	String() string
	init() error
	do() error
}

func Migrate() {
	do(NewMigrator())
}

func Invalidate() {
	do(NewInvalidator())
}

func BulkEval() {
	do(NewBulkEvaluator())
}

func FullEval() {
	do(NewFullEvaluator())
}

func do(doer doer) {
	slog.Info("running doer", slog.String("doer", doer.String()))

	// Setup sentry
	sentryTransport := sentry.NewHTTPSyncTransport()
	sentryTransport.Timeout = 3 * time.Second
	err := sentry.Init(sentry.ClientOptions{
		Release:     common.GitSHA,
		Environment: common.Config.Env,
		Transport:   sentryTransport,
		Tags: map[string]string{
			"command": "membership " + doer.String(),
		},
	})
	if err != nil {
		utils.LogFatal("sentry.Init", slog.Any("err", err))
	}
	defer sentry.Flush(2 * time.Second)

	// do the thing
	if err := doer.init(); err != nil {
		sentry.CaptureException(err)
		utils.LogFatal("doer.init", slog.Any("err", err))
	}

	if err := doer.do(); err != nil {
		sentry.CaptureException(err)
		utils.LogFatal("doer.do", slog.Any("err", err))
	}

	slog.Info("doer completed", slog.String("doer", doer.String()))
}
