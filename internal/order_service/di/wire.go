// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/gocql/gocql"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/controller/grpc"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/controller/http"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/repository/cassandra"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/config"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/event"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/event/publisher"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/infrastructure/identifier"
	infratime "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/infrastructure/time"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase/command"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase/query"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// ProvideLogger provides a logger instance
func ProvideLogger() logger.Logger {
	return logger.NewZapLogger()
}

// ProvideConfig provides a config instance
func ProvideConfig() (*config.Config, error) {
	return config.LoadConfig("config/order_service.yaml")
}

// ProvideCassandraSession provides a Cassandra session
func ProvideCassandraSession(cfg *config.Config) (*gocql.Session, error) {
	cluster := gocql.NewCluster(cfg.Cassandra.Hosts...)
	cluster.Keyspace = cfg.Cassandra.Keyspace
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = cfg.Cassandra.Timeout
	cluster.ConnectTimeout = cfg.Cassandra.ConnectTimeout
	
	if cfg.Cassandra.Username != "" && cfg.Cassandra.Password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.Cassandra.Username,
			Password: cfg.Cassandra.Password,
		}
	}
	
	return cluster.CreateSession()
}

// ProvideIDGenerator provides an ID generator
func ProvideIDGenerator() valueobject.IDGenerator {
	return identifier.NewUUIDGenerator()
}

// ProvideTimeProvider provides a time provider
func ProvideTimeProvider() valueobject.TimeProvider {
	return infratime.NewSystemTimeProvider()
}

// ProvideOrderRepository provides an order repository
func ProvideOrderRepository(session *gocql.Session) repository.OrderRepository {
	return cassandra.NewCassandraOrderRepository(session)
}

// ProvideOrderReadRepository provides an order read repository
func ProvideOrderReadRepository(session *gocql.Session) repository.OrderReadRepository {
	return cassandra.NewCassandraOrderReadRepository(session)
}

// ProvideOrderEventRepository provides an order event repository
func ProvideOrderEventRepository(session *gocql.Session) repository.OrderEventRepository {
	return cassandra.NewCassandraOrderEventRepository(session)
}

// ProvidePaymentRepository provides a payment repository
func ProvidePaymentRepository(session *gocql.Session) repository.PaymentRepository {
	return cassandra.NewCassandraPaymentRepository(session)
}

// ProvideShippingRepository provides a shipping repository
func ProvideShippingRepository(session *gocql.Session) repository.ShippingRepository {
	return cassandra.NewCassandraShippingRepository(session)
}

// ProvideUnitOfWork provides a unit of work
func ProvideUnitOfWork(
	session *gocql.Session,
	orderRepo *cassandra.CassandraOrderRepository,
	orderReadRepo *cassandra.CassandraOrderReadRepository,
	orderEventRepo *cassandra.CassandraOrderEventRepository,
	paymentRepo *cassandra.CassandraPaymentRepository,
	shippingRepo *cassandra.CassandraShippingRepository,
) repository.UnitOfWork {
	return cassandra.NewCassandraUnitOfWork(
		session,
		orderRepo,
		orderReadRepo,
		orderEventRepo,
		paymentRepo,
		shippingRepo,
		20, // Default batch size
	)
}

// ProvideEventPublisher provides an event publisher
func ProvideEventPublisher(cfg *config.Config) event.Publisher {
	return publisher.NewKafkaOrderEventPublisher(cfg.Kafka.Brokers)
}

// ProvideCreateOrderUsecase provides a create order use case
func ProvideCreateOrderUsecase(
	unitOfWork repository.UnitOfWork,
	idGenerator valueobject.IDGenerator,
	timeProvider valueobject.TimeProvider,
	eventPublisher event.Publisher,
) command.CreateOrderUsecase {
	return command.NewCreateOrderUsecase(unitOfWork, idGenerator, timeProvider, eventPublisher)
}

// ProvideUpdateOrderUsecase provides an update order use case
func ProvideUpdateOrderUsecase(
	unitOfWork repository.UnitOfWork,
	idGenerator valueobject.IDGenerator,
	timeProvider valueobject.TimeProvider,
	eventPublisher event.Publisher,
) command.UpdateOrderUsecase {
	return command.NewUpdateOrderUsecase(unitOfWork, idGenerator, timeProvider, eventPublisher)
}

// ProvideCancelOrderUsecase provides a cancel order use case
func ProvideCancelOrderUsecase(
	unitOfWork repository.UnitOfWork,
	idGenerator valueobject.IDGenerator,
	timeProvider valueobject.TimeProvider,
	eventPublisher event.Publisher,
) command.CancelOrderUsecase {
	return command.NewCancelOrderUsecase(unitOfWork, idGenerator, timeProvider, eventPublisher)
}

// ProvideProcessPaymentUsecase provides a process payment use case
func ProvideProcessPaymentUsecase(
	unitOfWork repository.UnitOfWork,
	idGenerator valueobject.IDGenerator,
	timeProvider valueobject.TimeProvider,
	eventPublisher event.Publisher,
) command.ProcessPaymentUsecase {
	return command.NewProcessPaymentUsecase(unitOfWork, idGenerator, timeProvider, eventPublisher)
}

// ProvideUpdateShippingUsecase provides an update shipping use case
func ProvideUpdateShippingUsecase(
	unitOfWork repository.UnitOfWork,
	idGenerator valueobject.IDGenerator,
	timeProvider valueobject.TimeProvider,
	eventPublisher event.Publisher,
) command.UpdateShippingUsecase {
	return command.NewUpdateShippingUsecase(unitOfWork, idGenerator, timeProvider, eventPublisher)
}

// ProvideGetOrderUsecase provides a get order use case
func ProvideGetOrderUsecase(
	orderReadRepo repository.OrderReadRepository,
) query.GetOrderUsecase {
	return query.NewGetOrderUsecase(orderReadRepo)
}

// ProvideListOrdersUsecase provides a list orders use case
func ProvideListOrdersUsecase(
	orderReadRepo repository.OrderReadRepository,
) query.ListOrdersUsecase {
	return query.NewListOrdersUsecase(orderReadRepo)
}

// ProvideOrderHistoryUsecase provides an order history use case
func ProvideOrderHistoryUsecase(
	orderEventRepo repository.OrderEventRepository,
) query.OrderHistoryUsecase {
	return query.NewOrderHistoryUsecase(orderEventRepo)
}

// ProvideGRPCController provides a gRPC controller
func ProvideGRPCController(
	createOrderUsecase command.CreateOrderUsecase,
	updateOrderUsecase command.UpdateOrderUsecase,
	cancelOrderUsecase command.CancelOrderUsecase,
	processPaymentUsecase command.ProcessPaymentUsecase,
	updateShippingUsecase command.UpdateShippingUsecase,
	getOrderUsecase query.GetOrderUsecase,
	listOrdersUsecase query.ListOrdersUsecase,
	orderHistoryUsecase query.OrderHistoryUsecase,
	logger logger.Logger,
) *grpcctl.OrderController {
	return grpcctl.NewOrderController(
		createOrderUsecase,
		updateOrderUsecase,
		cancelOrderUsecase,
		processPaymentUsecase,
		updateShippingUsecase,
		getOrderUsecase,
		listOrdersUsecase,
		orderHistoryUsecase,
		logger,
	)
}

// ProvideHTTPController provides an HTTP controller
func ProvideHTTPController(
	createOrderUsecase command.CreateOrderUsecase,
	updateOrderUsecase command.UpdateOrderUsecase,
	cancelOrderUsecase command.CancelOrderUsecase,
	processPaymentUsecase command.ProcessPaymentUsecase,
	updateShippingUsecase command.UpdateShippingUsecase,
	getOrderUsecase query.GetOrderUsecase,
	listOrdersUsecase query.ListOrdersUsecase,
	orderHistoryUsecase query.OrderHistoryUsecase,
	shippingRepo repository.ShippingRepository,
	logger logger.Logger,
) *httpctl.OrderHandler {
	return httpctl.NewOrderHandler(
		createOrderUsecase,
		updateOrderUsecase,
		cancelOrderUsecase,
		processPaymentUsecase,
		updateShippingUsecase,
		getOrderUsecase,
		listOrdersUsecase,
		orderHistoryUsecase,
		shippingRepo,
		logger,
	)
}

// InitializeGRPCController initializes the gRPC controller with all its dependencies
func InitializeGRPCController() (*grpcctl.OrderController, error) {
	wire.Build(
		ProvideLogger,
		ProvideConfig,
		ProvideCassandraSession,
		ProvideIDGenerator,
		ProvideTimeProvider,
		ProvideOrderRepository,
		ProvideOrderReadRepository,
		ProvideOrderEventRepository,
		ProvidePaymentRepository,
		ProvideShippingRepository,
		ProvideUnitOfWork,
		ProvideEventPublisher,
		ProvideCreateOrderUsecase,
		ProvideUpdateOrderUsecase,
		ProvideCancelOrderUsecase,
		ProvideProcessPaymentUsecase,
		ProvideUpdateShippingUsecase,
		ProvideGetOrderUsecase,
		ProvideListOrdersUsecase,
		ProvideOrderHistoryUsecase,
		ProvideGRPCController,
	)
	return nil, nil
}

// InitializeHTTPController initializes the HTTP controller with all its dependencies
func InitializeHTTPController() (*httpctl.OrderHandler, error) {
	wire.Build(
		ProvideLogger,
		ProvideConfig,
		ProvideCassandraSession,
		ProvideIDGenerator,
		ProvideTimeProvider,
		ProvideOrderRepository,
		ProvideOrderReadRepository,
		ProvideOrderEventRepository,
		ProvidePaymentRepository,
		ProvideShippingRepository,
		ProvideUnitOfWork,
		ProvideEventPublisher,
		ProvideCreateOrderUsecase,
		ProvideUpdateOrderUsecase,
		ProvideCancelOrderUsecase,
		ProvideProcessPaymentUsecase,
		ProvideUpdateShippingUsecase,
		ProvideGetOrderUsecase,
		ProvideListOrdersUsecase,
		ProvideOrderHistoryUsecase,
		ProvideHTTPController,
	)
	return nil, nil
}
