package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
	"goa.design/clue/debug"
	"goa.design/clue/log"
	"gopkg.in/yaml.v3"

	"github.com/PingCAP-QE/ee-apps/publisher"
	tiup "github.com/PingCAP-QE/ee-apps/publisher/gen/tiup"
	"github.com/PingCAP-QE/ee-apps/publisher/pkg/config"
)

func main() {
	// Define command line flags, add any other flag required to configure the
	// service.
	var (
		hostF      = flag.String("host", "localhost", "Server host (valid values: localhost)")
		domainF    = flag.String("domain", "", "Host domain name (overrides host domain specified in service design)")
		httpPortF  = flag.String("http-port", "", "HTTP port (overrides host HTTP port specified in service design)")
		configFile = flag.String("config", "config.yaml", "Path to config file")
		secureF    = flag.Bool("secure", false, "Use secure scheme (https or grpcs)")
		dbgF       = flag.Bool("debug", false, "Log request and response bodies")
	)
	flag.Parse()

	// Setup logger. Replace logger with your own log package of choice.
	format := log.FormatJSON
	if log.IsTerminal() {
		format = log.FormatTerminal
	}
	ctx := log.Context(context.Background(), log.WithFormat(format))
	if *dbgF {
		ctx = log.Context(ctx, log.WithDebug())
		log.Debugf(ctx, "debug logs enabled")
	}
	log.Print(ctx, log.KV{K: "http-port", V: *httpPortF})

	// Setup logger.
	logLevel := zerolog.InfoLevel
	if *dbgF {
		logLevel = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(logLevel)

	// Initialize the services.
	tiupSvc, err := initService(*configFile)
	if err != nil {
		log.Fatalf(ctx, err, "failed to initialize service")
	}

	// Wrap the services in endpoints that can be invoked from other services
	// potentially running in different processes.
	var (
		tiupEndpoints *tiup.Endpoints
	)
	{
		tiupEndpoints = tiup.NewEndpoints(tiupSvc)
		tiupEndpoints.Use(debug.LogPayloads())
		tiupEndpoints.Use(log.Endpoint)
	}

	// Create channel used by both the signal handler and server goroutines
	// to notify the main goroutine when to stop the server.
	errc := make(chan error)

	// Setup interrupt handler. This optional step configures the process so
	// that SIGINT and SIGTERM signals cause the services to stop gracefully.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)

	// Start the servers and send errors (if any) to the error channel.
	switch *hostF {
	case "localhost":
		{
			addr := "http://0.0.0.0:80"
			u, err := url.Parse(addr)
			if err != nil {
				log.Fatalf(ctx, err, "invalid URL %#v\n", addr)
			}
			if *secureF {
				u.Scheme = "https"
			}
			if *domainF != "" {
				u.Host = *domainF
			}
			if *httpPortF != "" {
				h, _, err := net.SplitHostPort(u.Host)
				if err != nil {
					log.Fatalf(ctx, err, "invalid URL %#v\n", u.Host)
				}
				u.Host = net.JoinHostPort(h, *httpPortF)
			} else if u.Port() == "" {
				u.Host = net.JoinHostPort(u.Host, "80")
			}
			handleHTTPServer(ctx, u, tiupEndpoints, &wg, errc, *dbgF)
		}

	default:
		log.Fatal(ctx, fmt.Errorf("invalid host argument: %q (valid hosts: localhost)", *hostF))
	}

	// Wait for signal.
	log.Printf(ctx, "exiting (%v)", <-errc)

	// Send cancellation signal to the goroutines.
	cancel()

	wg.Wait()
	log.Printf(ctx, "exited")
}

func initService(configFile string) (tiup.Service, error) {
	// Load and parse configuration
	var config config.Service
	{
		configData, err := os.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("Error reading config file: %v", err)
		}
		if err := yaml.Unmarshal(configData, &config); err != nil {
			return nil, fmt.Errorf("Error parsing config file: %v", err)
		}
	}

	logger := zerolog.New(os.Stderr).With().Timestamp().Str("service", tiup.ServiceName).Logger()

	// Configure Kafka kafkaWriter
	kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  config.Kafka.Brokers,
		Topic:    config.Kafka.Topic,
		Balancer: &kafka.LeastBytes{},
		Logger:   kafka.LoggerFunc(logger.Printf),
	})

	// Configure Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Addr,
		Password: config.Redis.Password,
		Username: config.Redis.Username,
		DB:       config.Redis.DB,
	})

	return publisher.NewTiup(&logger, kafkaWriter, redisClient, config.EventSource), nil
}
