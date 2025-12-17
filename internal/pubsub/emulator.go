package pubsub

import "os"

// EmulatorHostEnvVar is the environment variable used to specify the Pub/Sub emulator host.
const EmulatorHostEnvVar = "PUBSUB_EMULATOR_HOST"

// IsEmulatorEnabled returns true if the Pub/Sub emulator is configured via PUBSUB_EMULATOR_HOST.
func IsEmulatorEnabled() bool {
	return os.Getenv(EmulatorHostEnvVar) != ""
}

// GetEmulatorHost returns the emulator host address from the environment.
// Returns empty string if emulator is not enabled.
func GetEmulatorHost() string {
	return os.Getenv(EmulatorHostEnvVar)
}

