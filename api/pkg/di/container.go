package di

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/messaging"
	"github.com/NdoleStudio/httpsms/pkg/adapter"
	"github.com/NdoleStudio/httpsms/pkg/auth0"
	"github.com/NdoleStudio/httpsms/pkg/billing"
	"github.com/NdoleStudio/httpsms/pkg/config"
	"github.com/NdoleStudio/httpsms/pkg/database"
	"github.com/NdoleStudio/httpsms/pkg/events"
	"github.com/NdoleStudio/httpsms/pkg/heartbeat"
	"github.com/NdoleStudio/httpsms/pkg/message"
	"github.com/NdoleStudio/httpsms/pkg/phone"
	"github.com/NdoleStudio/httpsms/pkg/phoneapikey"
	"github.com/NdoleStudio/httpsms/pkg/pubsubclient"
	"github.com/NdoleStudio/httpsms/pkg/telemetry"
	"github.com/NdoleStudio/httpsms/pkg/turnstile"
	"github.com/NdoleStudio/httpsms/pkg/user"
	"github.com/NdoleStudio/httpsms/pkg/webhook"
	axiom "github.com/axiomhq/axiom-go/axiom"
	adapter "github.com/axiomhq/axiom-go/adapter/zerolog"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/api/option"
)

// Container is the dependency injection container
type Container struct {
	Config                 *config.Config
	Logger                 *zerolog.Logger
	DB                     *sql.DB
	RedisClient            *redis.Client
	FirestoreClient        *firestore.Client
	FirebaseAuthClient     *auth.Client
	FirebaseMessaging      *messaging.Client
	PubSubClient           *pubsub.Client
	UserRepo               user.Repository
	UserService            user.Service
	PhoneRepo              phone.Repository
	PhoneService           phone.Service
	PhoneAPIKeyRepo        phoneapikey.Repository
	PhoneAPIKeyService     phoneapikey.Service
	MessageRepo            message.Repository
	MessageService         message.Service
	HeartbeatRepo          heartbeat.Repository
	HeartbeatService       heartbeat.Service
	WebhookRepo            webhook.Repository
	WebhookService         webhook.Service
	BillingService         billing.Service
	TurnstileService       turnstile.Service
	Auth0Service           auth0.Service
	EventsService          events.Service
	TelemetryService       telemetry.Service
	FiberApp               *fiber.App
}

// New creates a new instance of the Container
func New(cfg *config.Config) (*Container, error) {
	c := &Container{
		Config: cfg,
	}

	logger, err := c.logger()
	if err != nil {
		return nil, errors.Wrap(err, "cannot initialize logger")
	}
	c.Logger = &logger

	db, err := c.database()
	if err != nil {
		return nil, errors.Wrap(err, "cannot initialize database")
	}
	c.DB = db

	rdb, err := c.redis()
	if err != nil {
		return nil, errors.Wrap(err, "cannot initialize redis")
	}
	c.RedisClient = rdb

	fbApp, err := c.firebaseApp()
	if err != nil {
		return nil, errors.Wrap(err, "cannot initialize firebase app")
	}

	fbAuth, err := fbApp.Auth(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "cannot initialize firebase auth")
	}
	c.FirebaseAuthClient = fbAuth

	fbMsg, err := fbApp.Messaging(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "cannot initialize firebase messaging")
	}
	c.FirebaseMessaging = fbMsg

	return c, nil
}

func (c *Container) database() (*sql.DB, error) {
	return database.New(c.Config.DatabaseURL)
}

func (c *Container) redis() (*redis.Client, error) {
	opt, err := redis.ParseURL(c.Config.RedisURI)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse redis url")
	}
	return redis.NewClient(opt), nil
}

func (c *Container) firebaseApp() (*firebase.App, error) {
	var opts []option.ClientOption
	if c.Config.FirebaseCredentials != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(c.Config.FirebaseCredentials)))
	}
	return firebase.NewApp(context.Background(), nil, opts...)
}

func (c *Container) logger() (zerolog.Logger, error) {
	if c.Config.AppEnv == "production" && os.Getenv("AXIOM_TOKEN") != "" {
		return c.axiomLogger()
	}

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	return zerolog.New(output).With().Timestamp().Logger(), nil
}

func (c *Container) axiomLogger() (zerolog.Logger, error) {
	if os.Getenv("AXIOM_TOKEN") == "" {
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		return zerolog.New(output).With().Timestamp().Logger(), nil
	}

	client, err := axiom.NewClient()
	if err != nil {
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		return zerolog.New(output).With().Timestamp().Logger(), nil
	}

	writer, err := adapter.New(adapter.SetClient(client), adapter.SetDataset(c.Config.AxiomDataset))
	if err != nil {
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		return zerolog.New(output).With().Timestamp().Logger(), nil
	}

	return zerolog.New(writer).With().Timestamp().Logger(), nil
}
