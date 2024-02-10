//go:build cgo

// this uses a cgo-required sqlite driver

package main

import (
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.Info("Using mattn/go-sqlite3")
}

func dbDriver() string {
	return "sqlite3"
}
