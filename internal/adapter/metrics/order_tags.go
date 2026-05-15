package metrics

import (
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
)

func orderStatusTag(s domain.Status) string {
	switch s {
	case domain.RECEIVED:
		return "recebida"
	case domain.IN_ANALYSIS:
		return "em_diagnostico"
	case domain.ANALYSIS_FINISHED:
		return "diagnostico_finalizado"
	case domain.AWAITING_APPROVAL:
		return "aguardando_aprovacao"
	case domain.AWAITING_STOCK_CONSULT:
		return "aguardando_consulta_estoque"
	case domain.AWAITING_STOCK_ORDER:
		return "aguardando_reposicao_estoque"
	case domain.IN_PROGRESS:
		return "em_execucao"
	case domain.FINISHED:
		return "finalizado"
	case domain.AWAITING_PAYMENT:
		return "aguardando_pagamento"
	case domain.PAYMENT_APPROVED:
		return "pagamento_aprovado"
	case domain.PAYMENT_FAILED:
		return "pagamento_falhou"
	case domain.REJECTED:
		return "rejeitado"
	case domain.DELIVERED:
		return "entregue"
	default:
		return "unknown"
	}
}

func orderPhaseForPreviousStatus(s domain.Status) string {
	switch s {
	case domain.RECEIVED, domain.IN_ANALYSIS, domain.ANALYSIS_FINISHED, domain.AWAITING_APPROVAL:
		return "diagnostico"
	case domain.AWAITING_STOCK_CONSULT, domain.AWAITING_STOCK_ORDER, domain.IN_PROGRESS:
		return "execucao"
	case domain.FINISHED, domain.AWAITING_PAYMENT, domain.PAYMENT_APPROVED, domain.PAYMENT_FAILED:
		return "pagamento"
	case domain.DELIVERED:
		return "finalizacao"
	default:
		return "unknown"
	}
}
