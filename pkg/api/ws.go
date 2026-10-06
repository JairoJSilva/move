package api

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Permite conexões do painel frontend em qualquer host/porta
	},
}

// WSHub gerencia clientes conectados via WebSocket e realiza broadcast de métricas.
type WSHub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]bool
}

// NewWSHub cria uma nova instância do Hub.
func NewWSHub() *WSHub {
	return &WSHub{
		clients: make(map[*websocket.Conn]bool),
	}
}

// HandleWS faz o upgrade de HTTP para WebSocket e registra o novo cliente.
func (h *WSHub) HandleWS(w http.ResponseWriter, r *http.Request, onConnect func() *WSMessage) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Erro ao atualizar conexão WebSocket: %v", err)
		return
	}

	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	// Envia estado inicial imediatamente se disponível
	if onConnect != nil {
		if initialMsg := onConnect(); initialMsg != nil {
			_ = conn.WriteJSON(initialMsg)
		}
	}

	// Mantém conexão viva lendo mensagens (ou aguardando encerramento)
	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.clients, conn)
			h.mu.Unlock()
			conn.Close()
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

// Broadcast envia uma mensagem JSON para todos os clientes conectados.
func (h *WSHub) Broadcast(msg WSMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for conn := range h.clients {
		err := conn.WriteJSON(msg)
		if err != nil {
			conn.Close()
			delete(h.clients, conn)
		}
	}
}
