package poller

import (
	"fmt"
	"log"
	"time"

	"mymodule/internal/config"
	"mymodule/internal/store"
)

type Service struct {
	cfg        *config.Config
	serverRepo *store.ServerRepo
}

func NewService(cfg *config.Config, serverRepo *store.ServerRepo) *Service {
	return &Service{
		cfg:        cfg,
		serverRepo: serverRepo,
	}
}

func (s *Service) Start() {
	ticker := time.NewTicker(time.Duration(s.cfg.Polling.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Polling servers...")

		servers, err := s.serverRepo.GetEnabledServers()
		if err != nil {
			log.Printf("error getting enabled servers: %v", err)
			continue
		}

		for _, server := range servers {
			fmt.Printf("Server: %s (%s:%d)\n", server.Name, server.Host, server.Port)
			for _, file := range server.MonitoredFiles {
				fmt.Printf("  - Monitored File: %s\n", file.DestPath)
			}
		}
	}
}
