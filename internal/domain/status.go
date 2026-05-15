package domain

import (
	"errors"
	"fmt"
)

var ErrInvalidTransition = errors.New("invalid status transition")

type Status string

const (
	RECEIVED               Status = "Recebida"
	IN_ANALYSIS            Status = "Em diagnóstico"
	ANALYSIS_FINISHED      Status = "Diagnóstico finalizado"
	AWAITING_APPROVAL      Status = "Aguardando aprovação"
	AWAITING_STOCK_CONSULT Status = "Aguardando consulta de estoque"
	AWAITING_STOCK_ORDER   Status = "Aguardando reposição de estoque"
	IN_PROGRESS            Status = "Em execução"
	FINISHED               Status = "Finalizado"
	AWAITING_PAYMENT       Status = "Aguardando pagamento"
	PAYMENT_APPROVED       Status = "Pagamento aprovado"
	PAYMENT_FAILED         Status = "Pagamento falhou"
	REJECTED               Status = "Rejeitado"
	DELIVERED              Status = "Entregue"
)

// Statuses is a convenience accessor for all valid status values.
var Statuses = struct {
	RECEIVED               Status
	IN_ANALYSIS            Status
	ANALYSIS_FINISHED      Status
	AWAITING_APPROVAL      Status
	AWAITING_STOCK_CONSULT Status
	AWAITING_STOCK_ORDER   Status
	IN_PROGRESS            Status
	FINISHED               Status
	AWAITING_PAYMENT       Status
	PAYMENT_APPROVED       Status
	PAYMENT_FAILED         Status
	REJECTED               Status
	DELIVERED              Status
}{
	RECEIVED:               RECEIVED,
	IN_ANALYSIS:            IN_ANALYSIS,
	ANALYSIS_FINISHED:      ANALYSIS_FINISHED,
	AWAITING_APPROVAL:      AWAITING_APPROVAL,
	AWAITING_STOCK_CONSULT: AWAITING_STOCK_CONSULT,
	AWAITING_STOCK_ORDER:   AWAITING_STOCK_ORDER,
	IN_PROGRESS:            IN_PROGRESS,
	FINISHED:               FINISHED,
	AWAITING_PAYMENT:       AWAITING_PAYMENT,
	PAYMENT_APPROVED:       PAYMENT_APPROVED,
	PAYMENT_FAILED:         PAYMENT_FAILED,
	REJECTED:               REJECTED,
	DELIVERED:              DELIVERED,
}

// canTransition maps each status to the set of statuses it may transition into.
var canTransition = map[Status][]Status{
	RECEIVED:               {IN_ANALYSIS},
	IN_ANALYSIS:            {ANALYSIS_FINISHED},
	ANALYSIS_FINISHED:      {AWAITING_APPROVAL},
	AWAITING_APPROVAL:      {AWAITING_STOCK_CONSULT, REJECTED},
	AWAITING_STOCK_CONSULT: {IN_PROGRESS, AWAITING_STOCK_ORDER},
	AWAITING_STOCK_ORDER:   {IN_PROGRESS},
	IN_PROGRESS:            {FINISHED},
	FINISHED:               {AWAITING_PAYMENT},
	AWAITING_PAYMENT:       {PAYMENT_APPROVED, PAYMENT_FAILED},
	PAYMENT_APPROVED:       {DELIVERED},
}

// CanTransition reports whether transitioning from → to is valid.
func CanTransition(from, to Status) bool {
	for _, s := range canTransition[from] {
		if s == to {
			return true
		}
	}
	return false
}

// InvalidTransitionError wraps ErrInvalidTransition with context.
func InvalidTransitionError(from, to Status) error {
	return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, from, to)
}
