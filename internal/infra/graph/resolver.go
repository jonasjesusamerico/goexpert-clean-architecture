package graph

import (
	"goexpert-clean-architecture/internal/usecase/order"
)

type Resolver struct {
	CreateOrderUseCase order.CreateOrderUseCase
	ListOrderUseCase   order.ListOrderUseCase
}
