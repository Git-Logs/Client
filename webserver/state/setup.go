package state

import (
	"context"
	"webserver/config"

	"github.com/bwmarrin/discordgo"
	"github.com/dgraph-io/badger/v4"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

var (
	Json      = jsoniter.ConfigCompatibleWithStandardLibrary
	Discord   *discordgo.Session
	Pool      *pgxpool.Pool
	Context   = context.Background()
	Validator = validator.New()
	Badger    *badger.DB
	Logger    *zap.Logger
	Config    *config.Config
)

const (
	configFile = "api-config.yaml"
)
