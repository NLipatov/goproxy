package app

import (
	"context"
	"fmt"
	"goproxy/internal/config"
	"goproxy/internal/modules"
)

type App struct {
	config config.Config
}

func NewApp(config config.Config) *App {
	return &App{
		config: config,
	}
}

func (app *App) Run(ctx context.Context) error {
	switch app.config.Mode {
	case "migrator":
		modules.NewMigrator().MigrateDb()
	case "proxy":
		modules.NewProxy().Start()
	case "rest-api":
		modules.NewUsersApi().Start()
	case "plan-controller":
		modules.NewPlanController().Start()
	case "google-auth":
		modules.NewGoogleAuthAPI().Start()
	case "plans-api":
		modules.NewPlansAPI().Start()
	case "billing-api": // ToDo: rename to billing-prices-api ?
		modules.NewBillingAPI().Start()
	case "crypto-cloud-billing-api":
		modules.NewCryptoCloudBillingAPI().Start()
	case "free-plan-billing-api":
		modules.NewFreePlanBillingAPI().Start()
	default:
		return fmt.Errorf("unsupported mode: %s", app.config.Mode)
	}

	return nil
}
