package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.WithContext(ctx).WithField("signal", sig).Infof("Received signal. Cancelling context")
			cancel()
		}
	}()

	db, err := initDB(ctx)
	if err != nil {
		log.WithError(err).Fatal("Unable to initialize database")
	}
	defer db.Close()

	log.Info("Database intialized")

	findPairs(ctx, db)

	log.Info("All done")
}

func findPairs(ctx context.Context, db *sqlx.DB) {
	errorsInARow := 0
	for ctx.Err() == nil {
		// since getTwoUnpairedItems is implemented as an infinite loop, force a timeout
		withTimeout, timeoutCancel := context.WithTimeout(ctx, 30*time.Second)
		first, second, err := getTwoUnpairedItems(withTimeout, db)
		timeoutCancel()
		if err != nil {
			log.WithError(err).Fatal("Unable to get two unpaired items")
		}

		toLog := log.WithFields(log.Fields{"first": first.Name, "second": second.Name})

		result, err := checkForPair(ctx, http.DefaultClient, first.Name, second.Name)
		if err != nil {
			toLog.WithError(err).Warn("Unable to check for pair")

			errorsInARow++
			if errorsInARow > 4 {
				// bail out so we're not potentially failing forever
				log.Error("Too many errors in a row. Bailing out")
				return
			}

			if ctx.Err() != nil {
				return
			}

			// back off just in case we're causing the issue
			time.Sleep(time.Duration(errorsInARow*2) * time.Second)
			continue
		}
		errorsInARow = 0

		toLog = toLog.WithFields(log.Fields{"result": result.Result, "emoji": result.Emoji, "new": result.IsNew})
		toLog.Info("Checked for pair")

		if strings.EqualFold(result.Result, nothing) {
			continue
		}

		err = insertNewPair(ctx, db, first, second, result)
		if err != nil {
			toLog.WithError(err).Fatal("Unable to insert pair")
		}
		time.Sleep(1 * time.Second)
	}
}
