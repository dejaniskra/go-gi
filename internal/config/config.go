package config

import (
	"encoding/json"
	"fmt"
	"os"
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
type Config struct {
	Http *Http `json:"http"`
}

func LoadConfig(path string) (*Config, error) {
	fmt.Println("Loading config from: ", path)

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config JSON: %w", err)
	}

	validateConfig(&cfg)

	return &cfg, nil
}

func validateConfig(cfg *Config) {
	if cfg.Http != nil {
		if cfg.Http.Port == nil {
			fmt.Print("No http_port provided, defaulting to 1738")
			defaultPort := 1738
			cfg.Http.Port = &defaultPort
		} else {
			if *cfg.Http.Port <= 0 {
				panic("http.port must be a positive integer")
			}
		}

		if cfg.Http.Protocols == nil {
			fmt.Println("No http_protocols provided, defaulting to HTTP1")

			cfg.Http.Protocols = &Protocols{
				HTTP1: true,
				HTTP2: false,
			}
		} else {
			if cfg.Http.Protocols.HTTP1 == false && cfg.Http.Protocols.HTTP2 == false {
				fmt.Println("All http_protocols set to false, defaulting to HTTP1")
				cfg.Http.Protocols.HTTP1 = true
			} else {
				if cfg.Http.Protocols.HTTP2 == true {
					cfg.Http.Protocols.HTTP1 = false
				} else {
					cfg.Http.Protocols.HTTP1 = true
					cfg.Http.Protocols.HTTP2 = false
				}
			}
		}

		if cfg.Http.Timeouts == nil {
			fmt.Println("No http_timeouts provided, setting default values 30 seconds")
			defaultTimeout := 30
			cfg.Http.Timeouts = &Timeouts{
				ReadRequest:       &defaultTimeout,
				ReadRequestHeader: &defaultTimeout,
				ResponseWrite:     &defaultTimeout,
				Idle:              &defaultTimeout,
			}
		} else {
			if cfg.Http.Timeouts.ReadRequest == nil {
				fmt.Println("No http.timeouts.read_request provided, defaulting to 30 seconds")
				defaultReadRequest := 30
				cfg.Http.Timeouts.ReadRequest = &defaultReadRequest
			} else {
				if *cfg.Http.Timeouts.ReadRequest <= 0 {
					panic("http.timeouts.read_request must be a positive integer")
				}
			}

			if cfg.Http.Timeouts.ReadRequestHeader == nil {
				fmt.Println("No http.timeouts.read_request provided, defaulting to 30 seconds")
				defaultReadRequestHeader := 30
				cfg.Http.Timeouts.ReadRequestHeader = &defaultReadRequestHeader
			} else {
				if *cfg.Http.Timeouts.ReadRequestHeader <= 0 {
					panic("http.timeouts.read_request_header must be a positive integer")
				}
			}

			if cfg.Http.Timeouts.ResponseWrite == nil {
				fmt.Println("No http.timeouts.read_request provided, defaulting to 30 seconds")
				defaultResponseWrite := 30
				cfg.Http.Timeouts.ResponseWrite = &defaultResponseWrite
			} else {
				if *cfg.Http.Timeouts.ResponseWrite <= 0 {
					panic("http.timeouts.response_write must be a positive integer")
				}
			}

			if cfg.Http.Timeouts.Idle == nil {
				fmt.Println("No http.timeouts.idle provided, defaulting to 30 seconds")
				defaultIdle := 30
				cfg.Http.Timeouts.Idle = &defaultIdle
			} else {
				if *cfg.Http.Timeouts.Idle <= 0 {
					panic("http.timeouts.idle must be a positive integer")
				}
			}
		}

		if cfg.Http.MaxHeaderBytes == nil {
			fmt.Println("No http.max_header_bytes provided, defaulting to 1 MB")
			defaultMaxHeaderBytes := 1 << 20
			cfg.Http.MaxHeaderBytes = &defaultMaxHeaderBytes
		} else {
			if *cfg.Http.MaxHeaderBytes <= 0 {
				panic("http.max_header_bytes must be a positive integer")
			}
		}
	}
}
