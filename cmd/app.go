/*
Go-micro-SAAS is a web application skeleton to quickly build new SAAS applications.

Usage:

	app [flags]

The flags are:

	    --migrate [=true/false]
	        When true a database migration is executed before starting
			the scheduler web application. Default value is false.
	    --application [=true/false]
	        Starts the web application for the scheduler. Default value
			is true.
	    --config [path]
		    Path of the configuration folder where the app.env config file
			is present. Default value is "."
*/
package main

import (
	"flag"

	"github.com/inokone/go-micro-saas/internal/app"
	"github.com/inokone/go-micro-saas/internal/common"
	"github.com/inokone/go-micro-saas/internal/db"
)

// @title                     Go-micro-SAAS API
// @version                   0.1
// @description               Go-micro-SAAS is a web application skeleton to quickly build new SAAS.
// @BasePath                  /api/v1
// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	var (
		isMigration = flag.Bool("migrate", false, "Start migration of the database. Default: [false]")
		application = flag.Bool("application", true, "Start the web application on the provided port. Default: [true].")
		config      = flag.String("config", ".", "Path of the configuration folder where the app.env file is. Default: [.]")
	)
	flag.Parse()
	c := common.InitApp(*config)
	if *isMigration {
		go db.ForceMigration(c.DB)
	}
	if *application {
		app.App(c)
	}
}
