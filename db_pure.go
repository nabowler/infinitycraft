//go:build !cgo

// this uses a pure go sqlite driver

package main

import (
	_ "github.com/glebarez/go-sqlite"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.Info("Using glebarez/go-sqlite")
}

func dbDriver() string {
	return "sqlite"
}
