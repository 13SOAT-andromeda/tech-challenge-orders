package metrics

import (
	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/domain"
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
	case domain.APPROVED:
		return "aprovado"
	case domain.IN_PROGRESS:
		return "em_execucao"
	case domain.FINISHED:
		return "finalizado"
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
	case domain.APPROVED, domain.IN_PROGRESS:
		return "execucao"
	case domain.FINISHED, domain.DELIVERED:
		return "finalizacao"
	default:
		return "unknown"
	}
}
