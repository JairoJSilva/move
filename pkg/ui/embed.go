package ui

import (
	"embed"
	"io/fs"
	"net/http"

	"migrations-engine/web"
)

// DistFS expõe o sistema de arquivos embutido compilado no pacote web.
var DistFS embed.FS = web.DistFS

// GetFS retorna a sub-árvore com raiz na pasta "dist", pronta para ser servida.
func GetFS() (fs.FS, error) {
	return fs.Sub(web.DistFS, "dist")
}

// HasEmbeddedUI verifica se o frontend SPA está presente nos arquivos embarcados.
func HasEmbeddedUI() bool {
	sub, err := fs.Sub(web.DistFS, "dist")
	if err != nil {
		return false
	}
	f, err := sub.Open("index.html")
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}

// HTTPHandler retorna um http.Handler para os arquivos embarcados.
func HTTPHandler() (http.Handler, error) {
	sub, err := GetFS()
	if err != nil {
		return nil, err
	}
	return http.FileServer(http.FS(sub)), nil
}
