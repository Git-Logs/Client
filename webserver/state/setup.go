package state

import (
	"os"
	"webserver/config"
	"webserver/mapofmu"

	"github.com/bwmarrin/discordgo"
	"github.com/infinitybotlist/eureka/genconfig"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

func Setup() {
	MapMutex = mapofmu.New[string]()

	// Initialize logger
	w := zapcore.AddSync(os.Stdout)

	var level = zap.InfoLevel
	if os.Getenv("DEBUG") == "true" {
		level = zap.DebugLevel
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		w,
		level,
	)

	Logger = zap.New(core)

	Logger.Info("Generating config")
	genconfig.SampleFileName = configFile + ".sample"

	genconfig.GenConfig(config.Config{})

	Logger.Info("Loading config", zap.String("file", configFile))
	cfg, err := os.ReadFile(configFile)

	if err != nil {
		Logger.Fatal("Could not read config file", zap.Error(err), zap.String("file", configFile))
	}

	err = yaml.Unmarshal(cfg, &Config)

	if err != nil {
		panic(err)
	}

	Logger.Info("Validating config")
	err = Validator.Struct(Config)

	if err != nil {
		Logger.Fatal("Config validation failed", zap.Error(err))
	}

	Logger.Info("Connecting to service [postgres]")
	Pool, err = pgxpool.New(Context, Config.PostgresURL)

	if err != nil {
		Logger.Fatal("Could not connect to postgres", zap.Error(err))
	}

	Logger.Info("Connecting to service [discord]")
	Discord, err = discordgo.New("Bot " + Config.Token)

	Discord.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMembers

	if err != nil {
		Logger.Fatal("Could not connect to discord", zap.Error(err))
	}

	err = Discord.Open()

	if err != nil {
		Logger.Fatal("Could not open discord connection", zap.Error(err))
	}

	Logger.Info("Connecting to service [badger]")
}
