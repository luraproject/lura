// SPDX-License-Identifier: Apache-2.0

/*

 */
package async

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/luraproject/lura/v2/backoff"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"

	"golang.org/x/sync/errgroup"
)

// Options contains the configuration to pass to the async agent factory
type Options struct {
	// Agent keeps the configuration for the async agent
	Agent *config.AsyncAgent
	// Endpoint encapsulates the configuration for the associated pipe
	Endpoint *config.EndpointConfig
	// Proxy is the pipe associated with the async agent
	Proxy proxy.Proxy
	// AgentPing is the channel for the agent to send ping messages
	AgentPing chan<- string
	// G is the error group responsible for managing the agents and the router itself
	G *errgroup.Group
	// ShouldContinue is a function signaling when to stop the connection retries
	ShouldContinue func(int) bool
	// BackoffF is a function encapsulating the backoff strategy
	BackoffF backoff.TimeToWaitBeforeRetry
	Logger   logging.Logger
}

// Factory is a function able to start an async agent
type Factory func(context.Context, Options) bool

// AgentStarter groups a set of factories to be used
type AgentStarter []Factory

// Start executes all the factories for each async agent configuration
func (a AgentStarter) Start(
	ctx context.Context,
	agents []*config.AsyncAgent,
	logger logging.Logger,
	agentPing chan<- string,
	pf proxy.Factory,
) func() error {
	if len(a) == 0 {
		return func() error { return ErrNoAgents }
	}

	g, ctx := errgroup.WithContext(ctx)

	for i, agent := range agents {
		i, agent := i, agent
		if agent.Name == "" {
			agent.Name = fmt.Sprintf("AsyncAgent-%02d", i)
		}

		logger.Debug(fmt.Sprintf("[SERVICE: AsyncAgent][%s] Starting the async agent", agent.Name))

		endpoint := &config.EndpointConfig{
			Endpoint:    agent.Name,
			Backend:     agent.Backend,
			ExtraConfig: agent.ExtraConfig,
		}
		p, err := pf.New(endpoint)
		if err != nil {
			logger.Error(fmt.Sprintf("[SERVICE: AsyncAgent][%s] building the proxy pipe:", agent.Name), err)
			continue
		}

		if agent.Connection.MaxRetries <= 0 {
			agent.Connection.MaxRetries = math.MaxInt64
		}

		opts := Options{
			Agent:          agent,
			Endpoint:       endpoint,
			Proxy:          p,
			AgentPing:      agentPing,
			G:              g,
			ShouldContinue: func(i int) bool { return i <= agent.Connection.MaxRetries },
			BackoffF:       backoff.GetByName(agent.Connection.BackoffStrategy),
			Logger:         logger,
		}

		for _, f := range a {
			if f(ctx, opts) {
				break
			}
		}

	}

	return g.Wait
}

var ErrNoAgents = errors.New("no agent factories defined")
