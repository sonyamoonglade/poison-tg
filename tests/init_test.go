package tests

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gofiber/fiber/v2"
	"github.com/sonyamoonglade/poison-tg/internal/api"
	"github.com/sonyamoonglade/poison-tg/internal/repositories"
	"github.com/sonyamoonglade/poison-tg/internal/telegram"
	"github.com/sonyamoonglade/poison-tg/internal/telegram/catalog"
	"github.com/sonyamoonglade/poison-tg/pkg/database"
	"github.com/sonyamoonglade/poison-tg/pkg/logger"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

var mongoURI, dbName string

type MockBot struct {
	mock.Mock
}

func (mb *MockBot) CleanRequest(c tg.Chattable) error {
	args := mb.Called(c)
	return args.Error(0)
}

func (mb *MockBot) SendMediaGroup(c tg.MediaGroupConfig) ([]tg.Message, error) {
	args := mb.Called(c)
	return args.Get(0).([]tg.Message), args.Error(1)
}

func (mb *MockBot) Send(c tg.Chattable) (tg.Message, error) {
	args := mb.Called(c)
	return args.Get(0).(tg.Message), args.Error(1)
}

func init() {
	mongoURI = os.Getenv("MONGO_URI")
	dbName = os.Getenv("DB_NAME")
}

type AppTestSuite struct {
	suite.Suite

	db           *database.Mongo
	tgrouter     *telegram.Router
	tghandler    telegram.RouteHandler
	api          *api.Handler
	repositories *repositories.Repositories
	mockBot      *MockBot
	updatesChan  <-chan tg.Update
	app          *fiber.App
}

func TestAPISuite(t *testing.T) {
	if testing.Short() {
		t.Skipf("skip e2e test")
	}

	suite.Run(t, new(AppTestSuite))
}

func (s *AppTestSuite) SetupSuite() {
	s.setupDeps()
}

func (s *AppTestSuite) TearDownSuite() {
	s.db.Close(context.Background()) //nolint:errcheck
}

func (s *AppTestSuite) TearDownSubTest(suiteName, testName string) {
	logger.Get().Sugar().Infof("running: %s", testName)
}

func (s *AppTestSuite) setupDeps() {

	logger.NewLogger(logger.Config{
		EnableStacktrace: true,
	})
	logger.Get().Info("Booting e2e test")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	mongo, err := database.Connect(ctx, mongoURI, dbName)
	if err != nil {
		s.FailNow("failed to connect to mongodb", err)
		return
	}

	catalogProvider := catalog.NewCatalogProvider()
	repos := repositories.NewRepositories(mongo, catalog.MakeUpdateOnChangeFunc(catalogProvider))

	rateProvider := api.NewRateProvider()
	apiHandler := api.NewHandler(repos.Catalog, repos.Order, repos.Customer, rateProvider)

	updates := make(chan tg.Update)
	mockBot := new(MockBot)
	tgHandler := telegram.NewHandler(mockBot, repos, rateProvider, catalogProvider)
	tgRouter := telegram.NewRouter(updates, tgHandler, repos.Customer, time.Second*5)

	mockBot.On("Send", mock.Anything).Return(tg.Message{}, nil)

	app := fiber.New(fiber.Config{
		Immutable:    true,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			logger.Get().Error("error in e2e test", zap.Error(err))
			return c.SendStatus(http.StatusInternalServerError)
		},
	})

	apiHandler.RegisterRoutes(app)

	s.app = app
	s.db = mongo
	s.updatesChan = updates
	s.tgrouter = tgRouter
	s.tghandler = tgHandler
	s.api = apiHandler
	s.repositories = &repos
	s.mockBot = mockBot
}
