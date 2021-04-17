package main

import (
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/monetrapp/rest-api/pkg/logging"
	"github.com/monetrapp/rest-api/pkg/migrations"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(databaseCommand)
	databaseCommand.AddCommand(migrateCommand)
	databaseCommand.AddCommand(databaseVersionCommand)

	databaseCommand.PersistentFlags().StringVarP(&postgresAddress, "host", "h", "localhost", "PostgreSQL host address.")
	databaseCommand.PersistentFlags().IntVarP(&postgresPort, "port", "P", 5432, "PostgreSQL port.")
	databaseCommand.PersistentFlags().StringVarP(&postgresUsername, "username", "U", "postgres", "PostgreSQL user.")
	databaseCommand.PersistentFlags().StringVarP(&postgresPassword, "password", "W", "", "PostgreSQL password.")
	databaseCommand.PersistentFlags().StringVarP(&postgresDatabase, "database", "d", "postgres", "PostgreSQL database.")
}

var (
	postgresAddress  = ""
	postgresPort     = 0
	postgresUsername = ""
	postgresPassword = ""
	postgresDatabase = ""
)

var (
	migrateCommand = &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations against your PostgreSQL.",
		Long:  "Updates your PostgreSQL database to the latest schema version for monetr.",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logging.NewLogger()

			options := getDatabaseCommandConfiguration()

			db := pg.Connect(options)

			migrator, err := migrations.NewMigrationsManager(log, db)
			if err != nil {
				log.WithError(err).Fatalf("failed to create migration manager")
				return err
			}

			oldVersion, newVersion, err := migrator.Up()
			if err != nil {
				log.WithError(err).Fatalf("failed to run schema migrations")
				return err
			}

			if oldVersion != newVersion {
				log.Infof("successfully upgraded database from %d to %d", oldVersion, newVersion)
			} else {
				log.Info("database is up to date, no migrations were run")
			}

			return nil
		},
	}

	databaseVersionCommand = &cobra.Command{
		Use:   "version",
		Short: "Prints version information about your database.",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logging.NewLogger()

			options := getDatabaseCommandConfiguration()

			db := pg.Connect(options)

			migrator, err := migrations.NewMigrationsManager(log, db)
			if err != nil {
				log.WithError(err).Fatalf("failed to create migration manager")
				return err
			}

			latestVersion, err := migrator.LatestVersion()
			if err != nil {
				log.WithError(err).Fatalf("failed to determine latest database version")
				return err
			}

			fmt.Println("Latest:", latestVersion)

			version, err := migrator.CurrentVersion()
			if err != nil {
				log.WithError(err).Fatalf("failed to determine current database version")
				return err
			}

			// No logging frills, just print the version to STDOUT
			fmt.Println("Current:", version)

			return nil
		},
	}

	databaseCommand = &cobra.Command{
		Use:   "database",
		Short: "Manages the PostgreSQL database used by monetr.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)

func getDatabaseCommandConfiguration() *pg.Options {
	options := &pg.Options{
		Addr:            fmt.Sprintf("%s:%d", postgresAddress, postgresPort),
		User:            postgresUsername,
		Password:        postgresPassword,
		Database:        postgresDatabase,
		ApplicationName: "monetr",
	}

	return options
}
