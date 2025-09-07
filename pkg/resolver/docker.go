package resolver

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type DockerInspector struct {
	cli *client.Client
}

func (d *DockerInspector) GetContainerMapping(ctx context.Context, filter Filter) (HostnameIPMapping, error) {
	dockerClient, dockerClientErr := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if dockerClientErr != nil {
		return nil, dockerClientErr
	}

	kvPairs := []filters.KeyValuePair{}
	for _, label := range filter.Labels {
		kvPairs = append(kvPairs, filters.Arg("label", label))
	}
	if filter.Name != "" {
		kvPairs = append(kvPairs, filters.Arg("name", filter.Name))
	}

	summary, summaryErr := dockerClient.ContainerList(ctx, container.ListOptions{
		Filters: filters.NewArgs(
			kvPairs...,
		),
	})
	if summaryErr != nil {
		return nil, summaryErr
	}

	containerMap := make(HostnameIPMapping)
	for _, container := range summary {
		inspectResp, inspectRespErr := dockerClient.ContainerInspect(ctx, container.ID)
		if inspectRespErr != nil {
			return nil, inspectRespErr
		}
		if inspectResp.NetworkSettings != nil {
			for _, network := range inspectResp.NetworkSettings.Networks {
				if network != nil {
					// create a set of strings
					for _, alias := range network.Aliases {
						containerMap[alias] = network.IPAddress
					}
					for _, hostname := range network.DNSNames {
						containerMap[hostname] = network.IPAddress
					}
				}
			}
		}
	}

	return containerMap, nil
}

func NewDockerInspector() (*DockerInspector, error) {
	dockerClient, dockerClientErr := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if dockerClientErr != nil {
		return nil, dockerClientErr
	}

	return &DockerInspector{
		cli: dockerClient,
	}, nil
}
