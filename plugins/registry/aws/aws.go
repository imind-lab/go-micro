package aws

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	sd "github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
	"go-micro.dev/v4/cmd"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/registry"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type awsRegistry struct {
	client      *sd.Client
	namespaceId string
	options     registry.Options

	sync.RWMutex

	services  map[string]map[string]string
	instances map[string]map[string]struct{}
	watchers  map[string]*Watcher
}

func init() {
	fmt.Println("Registries aws")
	cmd.DefaultRegistries["aws"] = NewRegistry
}

func NewRegistry(opts ...registry.Option) registry.Registry {
	e := &awsRegistry{
		options:   registry.Options{ Timeout: 5*time.Second},
		services:  make(map[string]map[string]string),
		instances: make(map[string]map[string]struct{}),
		watchers:  make(map[string]*Watcher),
	}
	fmt.Println("new")
	configure(e, opts...)
	return e
}

func configure(e *awsRegistry, _ ...registry.Option) error {
	fmt.Println("configure1")
	region, profile := os.Getenv("AWS_REGION"), os.Getenv("AWS_PROFILE")
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region), config.WithSharedConfigProfile(profile))
	if err != nil {
		log.Fatalf("failed to load SDK configuration, %v", err)
		return err
	}

	client := sd.NewFromConfig(cfg)
	e.client = client

	namespace := os.Getenv("AWS_REGISTRY_NAMESPACE")
	e.namespaceId = namespace
	fmt.Println("configure2")
	return nil
}

func (a *awsRegistry) Init(opts ...registry.Option) error {
	fmt.Println("Init")
	return configure(a, opts...)
}

func (a *awsRegistry) Options() registry.Options {
	fmt.Println("Options")
	return a.options
}

func (a *awsRegistry) registerNode(s *registry.Service, node *registry.Node, opts ...registry.RegisterOption) error {
	fmt.Println("registerNode", s.Version, s.Name, node.Id, node.Address)
	if len(s.Nodes) == 0 {
		return errors.New("Require at least one node")
	}

	// check existing lease cache
	a.RLock()
	services, ok := a.services[s.Name]
	a.RUnlock()

	// missing lease, check if the key exists
	ctx, cancel := context.WithTimeout(context.Background(), a.options.Timeout)
	defer cancel()

	if !ok {
		services = make(map[string]string)

		// look for the existing key
		name := s.Name + "_" + s.Version

		rsp, err := a.client.ListServices(ctx, &sd.ListServicesInput{Filters: []types.ServiceFilter{
			{
				Name:   types.ServiceFilterNameNamespaceId,
				Values: []string{a.namespaceId},
			},
		}})
		if err != nil {
			return err
		}

		svcId := ""
		for _, svc := range rsp.Services {
			if *svc.Name == name {
				svcId = *svc.Id
			}
		}

		if len(svcId) == 0 {
			rsp, err := a.client.CreateService(ctx, &sd.CreateServiceInput{Name: &name, NamespaceId: &a.namespaceId})
			if err != nil {
				return err
			}
			svcId = *rsp.Service.Id
		}

		services[s.Version] = svcId

		host, port, err := net.SplitHostPort(node.Address)
		if err != nil {
			return err
		}

		attr := map[string]string{
			"AWS_INSTANCE_IPV4": host,
			"AWS_INSTANCE_PORT": port,
		}

		a.client.RegisterInstance(ctx, &sd.RegisterInstanceInput{
			Attributes: attr,
			InstanceId: &node.Address,
			ServiceId:  &svcId,
		})

		a.instances[svcId] = map[string]struct{}{node.Address: {}}

		a.services[s.Name] = services
		return nil
	}

	service, ok := services[s.Version]
	if !ok {

		// look for the existing key
		name := s.Name + "_" + s.Version
		rsp, err := a.client.ListServices(ctx, &sd.ListServicesInput{Filters: []types.ServiceFilter{
			{
				Name:   types.ServiceFilterNameNamespaceId,
				Values: []string{a.namespaceId},
			},
		}})
		if err != nil {
			return err
		}

		svcId := ""
		for _, svc := range rsp.Services {
			if *svc.Name == name {
				svcId = *svc.Id
			}
		}

		if len(svcId) == 0 {
			rsp, err := a.client.CreateService(ctx, &sd.CreateServiceInput{Name: &name, NamespaceId: &a.namespaceId})
			if err != nil {
				return err
			}
			svcId = *rsp.Service.Id
		}

		services[s.Version] = svcId

		host, port, err := net.SplitHostPort(node.Address)
		if err != nil {
			return err
		}

		attr := map[string]string{
			"AWS_INSTANCE_IPV4": host,
			"AWS_INSTANCE_PORT": port,
		}

		a.client.RegisterInstance(ctx, &sd.RegisterInstanceInput{
			Attributes: attr,
			InstanceId: &node.Address,
			ServiceId:  &svcId,
		})

		a.instances[svcId] = map[string]struct{}{node.Address: {}}

		a.services[s.Name][s.Version] = service
		return nil
	}

	instances, ok := a.instances[service]
	if !ok {
		host, port, err := net.SplitHostPort(node.Address)
		if err != nil {
			return err
		}

		attr := map[string]string{
			"AWS_INSTANCE_IPV4": host,
			"AWS_INSTANCE_PORT": port,
		}

		a.client.RegisterInstance(ctx, &sd.RegisterInstanceInput{
			Attributes: attr,
			InstanceId: &node.Address,
			ServiceId:  &service,
		})

		a.instances[service] = map[string]struct{}{node.Address: {}}

		a.services[s.Name][s.Version] = service
	}

	_, ok = instances[node.Address]
	if !ok {
		host, port, err := net.SplitHostPort(node.Address)
		if err != nil {
			return err
		}

		attr := map[string]string{
			"AWS_INSTANCE_IPV4": host,
			"AWS_INSTANCE_PORT": port,
		}

		a.client.RegisterInstance(ctx, &sd.RegisterInstanceInput{
			Attributes: attr,
			InstanceId: &node.Address,
			ServiceId:  &service,
		})

		a.instances[service][node.Address] = struct{}{}
	}
	return nil
}

func (a *awsRegistry) Deregister(s *registry.Service, opts ...registry.DeregisterOption) error {
	if len(s.Nodes) == 0 {
		return errors.New("Require at least one node")
	}

	a.RLock()
	services, ok := a.services[s.Name]
	a.RUnlock()

	if !ok {
		return errors.New("Service does not exist")
	}

	service, ok := services[s.Version]
	if !ok {
		return errors.New("Service does not exist")
	}

	a.RLock()
	_, ok = a.instances[service]
	a.RUnlock()
	if !ok {
		return errors.New("Service does not exist")
	}

	for _, node := range s.Nodes {
		a.Lock()
		delete(a.instances[service], node.Address)
		a.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), a.options.Timeout)
		defer cancel()

		if logger.V(logger.TraceLevel, logger.DefaultLogger) {
			logger.Tracef("Deregistering %s id %s", s.Name, node.Id)
		}
		_, err := a.client.DeregisterInstance(ctx, &sd.DeregisterInstanceInput{
			InstanceId: &node.Address,
			ServiceId:  &service,
		})
		if err != nil {
			return err
		}
	}

	a.RLock()
	remainIns := len(a.instances[service])
	a.RUnlock()

	if remainIns == 0 {
		a.Lock()
		delete(a.instances, service)
		delete(a.services[s.Name], s.Version)
		a.Unlock()

		a.RLock()
		remainSvc := len(a.services[s.Name])
		a.RUnlock()

		if remainSvc == 0 {
			ctx, cancel := context.WithTimeout(context.Background(), a.options.Timeout)
			defer cancel()

			if logger.V(logger.TraceLevel, logger.DefaultLogger) {
				logger.Tracef("DeleteService %s", s.Name)
			}
			count := 0
			for {
				_, err := a.client.DeleteService(ctx, &sd.DeleteServiceInput{Id: &service})
				if err != nil {
					if !strings.Contains(err.Error(), "ResourceInUse") {
						return err
					}
					if count > 5{
						break
					}
				} else {
					break
				}
				count++
			}

		}
	}

	return nil
}

func (a *awsRegistry) Register(s *registry.Service, opts ...registry.RegisterOption) error {
	fmt.Println("Register", s)
	if len(s.Nodes) == 0 {
		return errors.New("Require at least one node")
	}

	var gerr error

	// register each node individually
	for _, node := range s.Nodes {
		err := a.registerNode(s, node, opts...)
		if err != nil {
			gerr = err
		}
	}

	return gerr
}

func (a *awsRegistry) GetService(name string, opts ...registry.GetOption) ([]*registry.Service, error) {
	fmt.Println("GetService", name)
	ctx, cancel := context.WithTimeout(context.Background(), a.options.Timeout)
	defer cancel()

	rsp, err := a.client.ListServices(ctx, &sd.ListServicesInput{Filters: []types.ServiceFilter{
		{
			Name:   types.ServiceFilterNameNamespaceId,
			Values: []string{a.namespaceId},
		},
	}})
	if err != nil {
		return nil, err
	}

	services := make([]*registry.Service, 0, len(rsp.Services))

	for _, svc := range rsp.Services {
		service := &registry.Service{}
		index := strings.LastIndex(*svc.Name, "_")
		nam := (*svc.Name)[:index]
		if name == nam {
			version := (*svc.Name)[index+1:]
			service.Name = nam
			service.Version = version

			serverId := *svc.Id
			a.Lock()
			a.services[nam][version] = serverId
			a.Unlock()

			ctx, cancel := context.WithTimeout(context.Background(), a.options.Timeout)
			defer cancel()

			rsp, err := a.client.ListInstances(ctx, &sd.ListInstancesInput{ServiceId: &serverId})
			if err != nil {
				return nil, err
			}

			nodes := make([]*registry.Node, 0, len(rsp.Instances))
			a.Lock()
			a.instances[serverId] = make(map[string]struct{}, len(rsp.Instances))
			a.Unlock()
			for _, instance := range rsp.Instances {
				a.Lock()
				a.instances[serverId][*instance.Id] = struct{}{}
				a.Unlock()

				node := &registry.Node{}
				node.Address = *instance.Id
				nodes = append(nodes, node)
			}

			service.Nodes = nodes

			services = append(services, service)
		}
	}

	return services, nil
}

func (a *awsRegistry) ListServices(_ ...registry.ListOption) ([]*registry.Service, error) {
	fmt.Println("ListServices")
	ctx, cancel := context.WithTimeout(context.Background(), a.options.Timeout)
	defer cancel()

	rsp, err := a.client.ListServices(ctx, &sd.ListServicesInput{Filters: []types.ServiceFilter{
		{
			Name:   types.ServiceFilterNameNamespaceId,
			Values: []string{a.namespaceId},
		},
	}})
	if err != nil {
		return nil, err
	}

	services := make([]*registry.Service, 0, len(rsp.Services))

	for _, svc := range rsp.Services {
		service := &registry.Service{}
		index := strings.LastIndex(*svc.Name, "_")
		name := (*svc.Name)[:index]
		version := (*svc.Name)[index+1:]
		service.Name = name
		service.Version = version

		serverId := *svc.Id
		a.Lock()
		a.services[name][version] = serverId
		a.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), a.options.Timeout)
		defer cancel()

		rsp, err := a.client.ListInstances(ctx, &sd.ListInstancesInput{ServiceId: &serverId})
		if err != nil {
			return nil, err
		}

		nodes := make([]*registry.Node, 0, len(rsp.Instances))
		a.Lock()
		a.instances[serverId] = make(map[string]struct{}, len(rsp.Instances))
		a.Unlock()
		for _, instance := range rsp.Instances {
			a.Lock()
			a.instances[serverId][*instance.Id] = struct{}{}
			a.Unlock()

			node := &registry.Node{}
			node.Address = *instance.Id
			nodes = append(nodes, node)
		}

		service.Nodes = nodes

		services = append(services, service)

	}

	return services, nil
}

func (a *awsRegistry) Watch(opts ...registry.WatchOption) (registry.Watcher, error) {
	return nil, nil
}

func (a *awsRegistry) String() string {
	return "aws"
}
