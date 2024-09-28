package rest

import (
	"encoding/json"
	"goexpert-clean-architecture/internal/usecase/order"
	"net/http"
)

type WebOrderHandler struct {
	CreateOrderUseCase *order.CreateOrderUseCase
	ListOrderUseCase   *order.ListOrderUseCase
}

func NewWebOrderHandler(
	createOrderUseCase *order.CreateOrderUseCase,
	listOrderUseCase *order.ListOrderUseCase,
) *WebOrderHandler {
	return &WebOrderHandler{
		CreateOrderUseCase: createOrderUseCase,
		ListOrderUseCase:   listOrderUseCase,
	}
}

func (h *WebOrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response, err := h.ListOrderUseCase.GetOrders()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h *WebOrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var dto order.OrderInputDTO

	err := json.NewDecoder(r.Body).Decode(&dto)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	output, err := h.CreateOrderUseCase.Execute(dto)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(output)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
