package config

import "time"

type Scheduler struct {
	Enabled   bool          `mapstructure:"SCHEDULER_ENABLED"`
	Interval  time.Duration `mapstructure:"SCHEDULER_INTERVAL"`
	BatchSize int           `mapstructure:"SCHEDULER_BATCH_SIZE"`
}

type Outbox struct {
	Interval  time.Duration `mapstructure:"OUTBOX_RELAY_INTERVAL"`
	BatchSize int           `mapstructure:"OUTBOX_RELAY_BATCH_SIZE"`
}
