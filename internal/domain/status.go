package domain

type Status string

const (
	RECEIVED          Status = "Recebida"
	IN_ANALYSIS       Status = "Em diagnóstico"
	ANALYSIS_FINISHED Status = "Diagnóstico finalizado"
	AWAITING_APPROVAL Status = "Aguardando aprovação"
	APPROVED          Status = "Aprovado"
	IN_PROGRESS       Status = "Em execução"
	FINISHED          Status = "Finalizado"
	DELIVERED         Status = "Entregue"
)

var Statuses = struct {
	RECEIVED          Status
	IN_ANALYSIS       Status
	ANALYSIS_FINISHED Status
	AWAITING_APPROVAL Status
	APPROVED          Status
	IN_PROGRESS       Status
	FINISHED          Status
	DELIVERED         Status
}{
	RECEIVED:          RECEIVED,
	IN_ANALYSIS:       IN_ANALYSIS,
	ANALYSIS_FINISHED: ANALYSIS_FINISHED,
	AWAITING_APPROVAL: AWAITING_APPROVAL,
	APPROVED:          APPROVED,
	IN_PROGRESS:       IN_PROGRESS,
	FINISHED:          FINISHED,
	DELIVERED:         DELIVERED,
}
