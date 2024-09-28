package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"goexpert-clean-architecture/configs"
	"goexpert-clean-architecture/internal/event"
	event_handler "goexpert-clean-architecture/internal/event/handler"
	"goexpert-clean-architecture/internal/infra/database"
	"goexpert-clean-architecture/internal/infra/graph"
	"goexpert-clean-architecture/internal/infra/grpc/pb"
	"goexpert-clean-architecture/internal/infra/grpc/service"
	"goexpert-clean-architecture/internal/infra/rest"
	usecase "goexpert-clean-architecture/internal/usecase/order"
	"goexpert-clean-architecture/pkg/events"

	_ "github.com/lib/pq"
)

func main() {
	cfg, err := configs.CarregaConfiguracao()
	if err != nil {
		panic(err)
	}

	db, err := conecataBaseDados()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	aplicaMigrationDaseDados(db)

	rabbitMQChannel := conectaAoRabbitMQ()
	defer rabbitMQChannel.Close()

	eventDispatcher := constroiEventDispatcher(rabbitMQChannel)

	createOrderUseCase := constroiCriaOrdemUseCase(db, eventDispatcher)
	listOrderUseCase := constroiListagemOrdemUseCase(db)

	iniciaServidorRest(cfg.WebServerPort, createOrderUseCase, listOrderUseCase)
	iniciaServidorGRPC(cfg.GRPCServerPort, *createOrderUseCase, *listOrderUseCase)
	iniciaServidorGraphQL(cfg.GraphQLServerPort, *createOrderUseCase, *listOrderUseCase)
}

func aplicaMigrationDaseDados(db *sql.DB) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	fSrc, err := (&file.File{}).Open("./migrations")
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithInstance("file", fSrc, "postgres", driver)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}
}

func conectaAoRabbitMQ() *amqp.Channel {
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		panic(err)
	}
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	return ch
}

func constroiEventDispatcher(ch *amqp.Channel) *events.EventDispatcher {
	eventDispatcher := events.NewEventDispatcher()
	eventDispatcher.Register("OrderCreated", &event_handler.OrderCreatedHandler{
		RabbitMQChannel: ch,
	})
	return eventDispatcher
}

func constroiCriaOrdemUseCase(db *sql.DB, eventDispatcher *events.EventDispatcher) *usecase.CreateOrderUseCase {
	orderRepository := database.NewOrderRepository(db)
	orderCreated := event.NewOrderCreated()
	return usecase.NewCreateOrderUseCase(orderRepository, orderCreated, eventDispatcher)
}

func constroiListagemOrdemUseCase(db *sql.DB) *usecase.ListOrderUseCase {
	orderRepository := database.NewOrderRepository(db)
	return usecase.NewListOrderUseCase(orderRepository)
}

func iniciaServidorRest(port string, createOrderUseCase *usecase.CreateOrderUseCase, listOrderUseCase *usecase.ListOrderUseCase) {
	ws := rest.NewServer(":8000")
	orderPath := "/order"
	webOrderHandler := rest.NewWebOrderHandler(createOrderUseCase, listOrderUseCase)
	ws.AddHandler(rest.NewRoute(orderPath, "POST", webOrderHandler.Create))
	ws.AddHandler(rest.NewRoute(orderPath, "GET", webOrderHandler.GetOrders))
	fmt.Println("Starting REST server on port", port)
	go ws.Start()
}

func iniciaServidorGRPC(port string, createOrderUseCase usecase.CreateOrderUseCase, listOrderUseCase usecase.ListOrderUseCase) {
	grpcServer := grpc.NewServer()
	orderService := service.NewOrderService(createOrderUseCase, listOrderUseCase)
	pb.RegisterOrderServiceServer(grpcServer, orderService)
	reflection.Register(grpcServer)

	fmt.Println("Starting gRPC server on port", port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		panic(err)
	}
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()
}

func iniciaServidorGraphQL(port string, createOrderUseCase usecase.CreateOrderUseCase, listOrderUseCase usecase.ListOrderUseCase) {
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{
		Resolvers: &graph.Resolver{
			CreateOrderUseCase: createOrderUseCase,
			ListOrderUseCase:   listOrderUseCase,
		},
	}))
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	fmt.Println("Starting GraphQL server on port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start GraphQL server: %v", err)
	}
}

func conecataBaseDados() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	return db, nil
}
