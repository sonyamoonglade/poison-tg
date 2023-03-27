package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/sonyamoonglade/poison-tg/config"
	"github.com/sonyamoonglade/poison-tg/internal/api"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/repositories"
	"github.com/sonyamoonglade/poison-tg/internal/services"
	"github.com/sonyamoonglade/poison-tg/internal/telegram"
	"github.com/sonyamoonglade/poison-tg/internal/telegram/catalog"
	"github.com/sonyamoonglade/poison-tg/pkg/database"
	"github.com/sonyamoonglade/poison-tg/pkg/logger"
	"go.uber.org/zap"
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

	repos := repositories.NewRepositories(mongo, updateOnChange)
	initialCatalog, err := repos.Catalog.GetCatalog(ctx)
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
	rateProvider := api.NewRateProvider()
	yuanService := services.NewYuanService(rateProvider)

	handler := telegram.NewHandler(bot,
		repos,
		yuanService,
		catalogProvider)

	router := telegram.NewRouter(bot.GetUpdates(),
		handler,
		repos.Customer,
		cfg.Bot.HandlerTimeout)

	// HTTP api
	app := fiber.New(fiber.Config{
		Immutable: true,
		Prefork:   false,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			logger.Get().Error("error in api endpoint", zap.Error(err))
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "*",
	}))

	apiController := api.NewHandler(repos.Catalog, repos.Order, repos.Customer, rateProvider)
	apiController.RegisterRoutes(app)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		if err := app.Listen(":" + cfg.App.Port); err != nil {
			logger.Get().Error("http server error", zap.Error(err))
		}
		wg.Done()
	}()
	logger.Get().Info("http api server is up")

	if err := router.Bootstrap(); err != nil {
		return err
	}

	exitChan := make(chan os.Signal)
	signal.Notify(exitChan, os.Interrupt, syscall.SIGINT)

	// Graceful shutdown
	<-exitChan
	if err := app.Shutdown(); err != nil {
		return fmt.Errorf("api shutdown: %w", err)
	}

	wg.Wait()
	return mongo.Close(context.Background())
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
