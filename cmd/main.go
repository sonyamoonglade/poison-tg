package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/sonyamoonglade/poison-tg/config"
	"github.com/sonyamoonglade/poison-tg/internal/telegram"
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

	if err := loadEnvs(); err != nil {
		logger.Get().Sugar().Warn(err)
	}

	cfg, err := config.ReadConfig(configPath)
	if err != nil {
		return fmt.Errorf("can't read config: %w", err)
	}

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	// defer cancel()

	// mongo, err := database.Connect(ctx, cfg.Database.URI, cfg.Database.Name)
	// if err != nil {
	// 	return fmt.Errorf("error connecting to mongo: %w", err)
	// }

	bot, err := telegram.NewBot(telegram.Config{
		Token: cfg.Bot.Token,
	})
	if err != nil {
		return fmt.Errorf("error creating telegram bot: %w", err)
	}

	handler := telegram.NewHandler(bot)
	router := telegram.NewRouter(bot.GetUpdates(), handler, cfg.Bot.HandlerTimeout)

	// _ = mongo

	router.Bootstrap()
	return nil
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

func loadEnvs() error {
	if err := godotenv.Load(".env"); err != nil {
		return fmt.Errorf("can't load environment variables from .env: %w", err)
	}
	return nil
}
