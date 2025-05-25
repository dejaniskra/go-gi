package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dejaniskra/go-gi/logger"
)

type Protocols struct {
	HTTP1 bool `json:"http_1"`
	HTTP2 bool `json:"http_2"`
}
type Timeouts struct {
	ReadRequest       *int `json:"read_request"`
	ReadRequestHeader *int `json:"read_request_header"`
	ResponseWrite     *int `json:"response_write"`
	Idle              *int `json:"idle"`
}
type Http struct {
	Port           *int       `json:"port"`
	Protocols      *Protocols `json:"protocols"`
	Timeouts       *Timeouts  `json:"timeouts"`
	MaxHeaderBytes *int       `json:"max_header_bytes"`
}
type Log struct {
	Level  logger.Level  `json:"level"`
	Format logger.Format `json:"format"`
}

type Config struct {
	Http *Http `json:"http"`
	Log  *Log  `json:"log"`
}

var cfg *Config

func LoadConfig(path string) *Config {
	if cfg != nil {
		return cfg
	}

	fmt.Println("Loading config from: ", path)

	file, err := os.Open(path)
	if err != nil {
		panic(fmt.Sprintf("failed to open config file: %v", err))
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		panic(fmt.Sprintf("failed to decode config JSON: %v", err))
	}

	validateConfig(cfg)

	return cfg
}

func validateConfig(cfg *Config) {
	setDefaultLog(cfg)
	setDefaultHttp(cfg)
}

func setDefaultLog(cfg *Config) {
	if cfg.Log != nil {
		return
	}

	fmt.Println("No log configuration provided, defaulting to INFO level and JSON format")
	cfg.Log = &Log{
		Level:  logger.INFO,
		Format: logger.JSON,
	}
}

func setDefaultHttp(cfg *Config) {
	if cfg.Http == nil {
		fmt.Println("No http configuration provided, using defaults")
		cfg.Http = &Http{}
	}

	defaultInt(&cfg.Http.Port, 1738, "http.port")

	if cfg.Http.Protocols == nil {
		fmt.Println("No http.protocols provided, defaulting to HTTP1")
		cfg.Http.Protocols = &Protocols{HTTP1: true}
	} else {
		if !cfg.Http.Protocols.HTTP1 && !cfg.Http.Protocols.HTTP2 {
			fmt.Println("All http.protocols set to false, defaulting to HTTP1")
			cfg.Http.Protocols.HTTP1 = true
		} else {
			// Ensure only one protocol is true
			if cfg.Http.Protocols.HTTP2 {
				cfg.Http.Protocols.HTTP1 = false
			} else {
				cfg.Http.Protocols.HTTP1 = true
				cfg.Http.Protocols.HTTP2 = false
			}
		}
	}

	// Timeouts
	if cfg.Http.Timeouts == nil {
		cfg.Http.Timeouts = &Timeouts{}
		fmt.Println("No http.timeouts provided, defaulting to 30 seconds")
	}
	defaultInt(&cfg.Http.Timeouts.ReadRequest, 30, "http.timeouts.read_request")
	defaultInt(&cfg.Http.Timeouts.ReadRequestHeader, 30, "http.timeouts.read_request_header")
	defaultInt(&cfg.Http.Timeouts.ResponseWrite, 30, "http.timeouts.response_write")
	defaultInt(&cfg.Http.Timeouts.Idle, 30, "http.timeouts.idle")

	// Max header bytes
	defaultInt(&cfg.Http.MaxHeaderBytes, 1<<20, "http.max_header_bytes")

	if *cfg.Http.Port <= 0 {
		panic("http.port must be a positive integer")
	}
	if *cfg.Http.MaxHeaderBytes <= 0 {
		panic("http.max_header_bytes must be a positive integer")
	}
}

func defaultInt(target **int, value int, name string) {
	if *target == nil {
		fmt.Printf("No %s provided, defaulting to %d\n", name, value)
		*target = &value
	}
}
