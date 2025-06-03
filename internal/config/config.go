package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type DBRoleConfig struct {
	Writer DBConnection  `json:"writer"`
	Reader *DBConnection `json:"reader"`
}

type DBConnection struct {
	User            string  `json:"user"`
	Password        string  `json:"password"`
	Host            string  `json:"host"`
	Port            string  `json:"port"`
	DBName          string  `json:"db_name"`
	Options         *string `json:"options"`
	MaxOpenConns    *int    `json:"max_open_connections"` // max number of open connections
	MaxIdleConns    *int    `json:"max_idle_connections"` // max number of idle connections
	ConnMaxLifetime *int    `json:"max_life_time"`        // seconds
}

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
	Level  string `json:"level"`
	Format string `json:"format"`
}

type DynamoConfig struct {
	Region    string  `json:"region"`
	AccessKey string  `json:"access_key"`
	SecretKey string  `json:"secret_key"`
	Endpoint  *string `json:"endpoint"`
}

type Config struct {
	Http     *Http                    `json:"http"`
	MySQL    map[string]*DBRoleConfig `json:"mysql"`
	Postgres map[string]*DBRoleConfig `json:"postgres"`
	Dynamo   map[string]*DynamoConfig `json:"dynamo"`
	Log      *Log                     `json:"log"`
}

var cfg *Config

func GetConfig() *Config {
	if cfg != nil {
		return cfg
	}

	file, err := os.Open("config.json")
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
		Level:  "INFO",
		Format: "JSON",
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
