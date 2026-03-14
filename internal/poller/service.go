package poller

import (
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
		log.Println("Polling all servers...")

		servers, err := s.serverRepo.GetEnabledServers()
		if err != nil {
			log.Printf("error getting enabled servers: %v", err)
			continue
		}

		for _, server := range servers {
			log.Printf("Polling server: %s", server.Name)
			for _, file := range server.MonitoredFiles {
				log.Printf("  - Refreshing file: %s", file.DestPath)
			}
		}
		log.Println("Polling finished.")
	}
}
