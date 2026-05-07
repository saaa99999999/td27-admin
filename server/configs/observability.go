package configs

type Observability struct {
	Prometheus Prometheus `mapstructure:"prometheus" json:"prometheus" yaml:"prometheus"`
	Otel       Otel       `mapstructure:"otel" json:"otel" yaml:"otel"`
}

type Prometheus struct {
	Enabled     bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	MetricsPath string `mapstructure:"metrics-path" json:"metrics-path" yaml:"metrics-path"`
}

type Otel struct {
	Enabled      bool    `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Endpoint     string  `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint"`
	ServiceName  string  `mapstructure:"service-name" json:"service-name" yaml:"service-name"`
	SamplingRate float64 `mapstructure:"sampling-rate" json:"sampling-rate" yaml:"sampling-rate"`
}
