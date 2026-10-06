package platform

// PriorityMode define o modo de prioridade do processo.
type PriorityMode int

const (
	// PriorityNormal modo de prioridade normal do sistema operacional.
	PriorityNormal PriorityMode = iota
	// PriorityLow modo de baixa prioridade (Best Effort classe 7 no Linux, Below Normal no Windows).
	PriorityLow
	// PriorityIdle modo ocioso (IOPRIO_CLASS_IDLE no Linux, PROCESS_MODE_BACKGROUND_BEGIN no Windows).
	PriorityIdle
)
