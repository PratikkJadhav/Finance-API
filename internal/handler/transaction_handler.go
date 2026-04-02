// internal/handler/transaction_handler.go
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/PratikkJadhav/Finance-API/internal/dto"
	"github.com/PratikkJadhav/Finance-API/internal/middleware"
	"github.com/PratikkJadhav/Finance-API/internal/service"
	"github.com/go-chi/chi/v5"
)

type TransactionHandler struct {
	txnService *service.TransactionService
}

func NewTransactionHandler(txnService *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{txnService: txnService}
}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input dto.CreateTransactionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input.UserID = middleware.GetUserID(r)

	txn, err := h.txnService.Create(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, txn)
}

func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := dto.TransactionFilter{
		Type:     r.URL.Query().Get("type"),
		Category: r.URL.Query().Get("category"),
		From:     r.URL.Query().Get("from"),
		To:       r.URL.Query().Get("to"),
		Page:     parseIntQuery(r, "page", 1),
		Limit:    parseIntQuery(r, "limit", 20),
	}

	userID := middleware.GetUserID(r)
	role := middleware.GetRole(r)

	// viewer and analysts only see their own transactions
	// admins can see all
	if role != "admin" {
		filter.UserID = userID
	}

	txns, total, err := h.txnService.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch transactions")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"transactions": txns,
		"total":        total,
		"page":         filter.Page,
		"limit":        filter.Limit,
	})
}

func (h *TransactionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	txn, err := h.txnService.GetByID(r.Context(), id)
	if err != nil {
		handleRepoError(w, err) // was: writeError(w, http.StatusNotFound, "transaction not found")
		return
	}
	writeJSON(w, http.StatusOK, txn)
}

func (h *TransactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var input dto.UpdateTransactionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	txn, err := h.txnService.Update(r.Context(), id, input)
	if err != nil {
		handleRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, txn)
}

func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.txnService.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "transaction deleted successfully",
	})
}

func (h *TransactionHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	role := middleware.GetRole(r)

	summary, err := h.txnService.GetSummary(r.Context(), userID, role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch summary")
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func (h *TransactionHandler) GetTrends(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	role := middleware.GetRole(r)

	trends, err := h.txnService.GetTrends(r.Context(), userID, role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch trends")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"trends": trends,
	})
}

func (h *TransactionHandler) GetCategoryTotals(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	role := middleware.GetRole(r)

	totals, err := h.txnService.GetCategoryTotals(r.Context(), userID, role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch category totals")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"categories": totals,
	})
}

func (h *TransactionHandler) GetRecent(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	role := middleware.GetRole(r)

	txns, err := h.txnService.GetRecent(r.Context(), userID, role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch recent transactions")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"recent": txns,
		"count":  len(txns),
	})
}
