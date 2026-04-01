package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

const TypeDockerContainers = "metrics:docker"

// DockerContainer represents a running Docker container.
type DockerContainer struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	State   string `json:"state"`
	Status  string `json:"status"`
	Created int64  `json:"created"`
}

// DockerPayload is the payload for Docker container messages.
type DockerPayload struct {
	Available  bool              `json:"available"`
	Containers []DockerContainer `json:"containers"`
	Total      int               `json:"total"`
	Timestamp  int64             `json:"timestamp"`
}

const dockerSocket = "/var/run/docker.sock"

// DockerCollector creates a collector that gathers Docker container info.
// Returns empty payload if Docker is not available.
func DockerCollector(interval time.Duration) *Collector {
	return New(interval, func() (string, any, error) {
		if _, err := os.Stat(dockerSocket); err != nil {
			return TypeDockerContainers, DockerPayload{
				Available: false,
				Timestamp: time.Now().UnixMilli(),
			}, nil
		}

		containers, err := queryDocker()
		if err != nil {
			return TypeDockerContainers, DockerPayload{
				Available: true,
				Timestamp: time.Now().UnixMilli(),
			}, nil
		}

		return TypeDockerContainers, DockerPayload{
			Available:  true,
			Containers: containers,
			Total:      len(containers),
			Timestamp:  time.Now().UnixMilli(),
		}, nil
	})
}

// dockerAPIContainer matches the Docker Engine API /containers/json response.
type dockerAPIContainer struct {
	ID      string   `json:"Id"`
	Names   []string `json:"Names"`
	Image   string   `json:"Image"`
	State   string   `json:"State"`
	Status  string   `json:"Status"`
	Created int64    `json:"Created"`
}

func queryDocker() ([]DockerContainer, error) {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", dockerSocket)
			},
		},
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("http://localhost/containers/json?all=true")
	if err != nil {
		return nil, fmt.Errorf("docker api: %w", err)
	}
	defer resp.Body.Close()

	var apiContainers []dockerAPIContainer
	if err := json.NewDecoder(resp.Body).Decode(&apiContainers); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	containers := make([]DockerContainer, len(apiContainers))
	for i, c := range apiContainers {
		name := c.ID[:12]
		if len(c.Names) > 0 {
			name = c.Names[0]
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
		}

		containers[i] = DockerContainer{
			ID:      c.ID[:12],
			Name:    name,
			Image:   c.Image,
			State:   c.State,
			Status:  c.Status,
			Created: c.Created,
		}
	}

	return containers, nil
}
