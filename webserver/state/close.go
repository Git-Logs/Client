package state

func Close() {
	Logger.Info("Closing service [postgres]")
	Pool.Close()

	Logger.Info("Closing service [discord]")
	Discord.Close()
}
