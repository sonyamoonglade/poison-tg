package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/sonyamoonglade/poison-tg/config"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/repositories"
	"github.com/sonyamoonglade/poison-tg/internal/services"
	"github.com/sonyamoonglade/poison-tg/internal/telegram"
	"github.com/sonyamoonglade/poison-tg/internal/telegram/catalog"
	"github.com/sonyamoonglade/poison-tg/pkg/database"
	"github.com/sonyamoonglade/poison-tg/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	configPath, logsPath, production, strict := readCmdArgs()

	if err := logger.NewLogger(logger.Config{
		Out:              []string{logsPath},
		Strict:           strict,
		Production:       production,
		EnableStacktrace: false,
	}); err != nil {
		return fmt.Errorf("error instantiating logger: %w", err)
	}

	cfg, err := config.ReadConfig(configPath)
	if err != nil {
		return fmt.Errorf("can't read config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	mongo, err := database.Connect(ctx, cfg.Database.URI, cfg.Database.Name)
	if err != nil {
		return fmt.Errorf("error connecting to mongo: %w", err)
	}

	catalogProvider := catalog.NewCatalogProvider()
	updateOnChange := func(items []domain.CatalogItem) {
		catalogProvider.Load(items)
	}

	customerRepo := repositories.NewCustomerRepo(mongo.Collection("customers"))
	orderRepo := repositories.NewOrderRepo(mongo.Collection("orders"))
	businessRepo := repositories.NewBusinessRepo(mongo.Collection("business"))
	catalogRepo := repositories.NewCatalogRepo(mongo.Collection("catalog"), updateOnChange)

	initialCatalog, err := catalogRepo.GetCatalog(ctx)
	if err != nil {
		return fmt.Errorf("error getting initial catalog: %w", err)
	}

	catalogProvider.Load(initialCatalog)

	bot, err := telegram.NewBot(telegram.Config{
		Token: cfg.Bot.Token,
	})
	if err != nil {
		return fmt.Errorf("error creating telegram bot: %w", err)
	}

	if err := telegram.LoadTemplates("templates.json"); err != nil {
		return fmt.Errorf("can't load templates: %w", err)
	}

	yuanService := services.NewYuanService(new(rateProvider))

	handler := telegram.NewHandler(bot,
		customerRepo,
		businessRepo,
		orderRepo,
		yuanService,
		catalogProvider)

	router := telegram.NewRouter(bot.GetUpdates(),
		handler,
		customerRepo,
		cfg.Bot.HandlerTimeout)

	return router.Bootstrap()
}

type rateProvider struct{}

func (r *rateProvider) GetYuanRate() (float64, error) {
	return 11.6, nil
}

func readCmdArgs() (string, string, bool, bool) {
	production := flag.Bool("production", false, "if logger should write to file")
	logsPath := flag.String("logs-path", "", "where log file is")
	strict := flag.Bool("strict", false, "if logger should log only warn+ logs")
	configPath := flag.String("config-path", "", "where config file is")

	flag.Parse()

	// Critical for app if not specified
	if *configPath == "" {
		panic("config path is not provided")
	}

	// Naked return, see return variable names
	return *configPath, *logsPath, *production, *strict
}
