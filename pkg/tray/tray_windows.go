//go:build windows

package tray

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

const (
	NIM_ADD    = 0x00000000
	NIM_MODIFY = 0x00000001
	NIM_DELETE = 0x00000002

	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004
	NIF_INFO    = 0x00000010

	NIIF_INFO = 0x00000001

	WM_DESTROY       = 0x0002
	WM_APP           = 0x8000
	WM_TRAYICON      = WM_APP + 1
	WM_LBUTTONDBLCLK = 0x0203
	WM_RBUTTONUP     = 0x0205
	WM_CONTEXTMENU   = 0x007B

	TPM_RETURNCMD   = 0x0100
	TPM_LEFTALIGN   = 0x0000
	TPM_BOTTOMALIGN = 0x0020

	MF_STRING    = 0x0000
	MF_SEPARATOR = 0x0800

	IDM_OPEN   = 1001
	IDM_STATUS = 1002
	IDM_TOGGLE = 1003
	IDM_QUIT   = 1004

	IDI_APPLICATION = uintptr(32512)
)

var (
	shell32                = syscall.NewLazyDLL("shell32.dll")
	procShellNotifyIconW   = shell32.NewProc("Shell_NotifyIconW")
	procShellExecuteW      = shell32.NewProc("ShellExecuteW")

	user32                 = syscall.NewLazyDLL("user32.dll")
	procRegisterClassExW   = user32.NewProc("RegisterClassExW")
	procCreateWindowExW    = user32.NewProc("CreateWindowExW")
	procDestroyWindow      = user32.NewProc("DestroyWindow")
	procDefWindowProcW     = user32.NewProc("DefWindowProcW")
	procPostQuitMessage    = user32.NewProc("PostQuitMessage")
	procGetMessageW        = user32.NewProc("GetMessageW")
	procTranslateMessage   = user32.NewProc("TranslateMessage")
	procDispatchMessageW   = user32.NewProc("DispatchMessageW")
	procCreatePopupMenu    = user32.NewProc("CreatePopupMenu")
	procAppendMenuW        = user32.NewProc("AppendMenuW")
	procTrackPopupMenu     = user32.NewProc("TrackPopupMenu")
	procDestroyMenu        = user32.NewProc("DestroyMenu")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procGetCursorPos       = user32.NewProc("GetCursorPos")
	procLoadIconW          = user32.NewProc("LoadIconW")
	procPostMessageW       = user32.NewProc("PostMessageW")

	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procGetModuleHandleW   = kernel32.NewProc("GetModuleHandleW")
)

type POINT struct {
	X int32
	Y int32
}

type MSG struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type WNDCLASSEXW struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type NOTIFYICONDATAW struct {
	cbSize            uint32
	hWnd              uintptr
	uID               uint32
	uFlags            uint32
	uCallbackMessage  uint32
	hIcon             uintptr
	szTip             [128]uint16
	dwState           uint32
	dwStateMask       uint32
	szInfo            [256]uint16
	uTimeoutOrVersion uint32
	szInfoTitle       [64]uint16
	dwInfoFlags       uint32
	guidItem          [16]byte
	hBalloonIcon      uintptr
}

type windowsTray struct {
	cfg   Config
	hWnd  uintptr
	nid   NOTIFYICONDATAW
	mu    sync.Mutex
	stopC chan struct{}
}

// Instância global para vincular o callback nativo wndProc à instância do controlador
var (
	globalTray   *windowsTray
	globalTrayMu sync.Mutex
)

// NewTray inicializa a bandeja para Windows sem dependência de CGO.
func NewTray(cfg Config) TrayController {
	if cfg.Title == "" {
		cfg.Title = "MoveOps"
	}
	t := &windowsTray{
		cfg:   cfg,
		stopC: make(chan struct{}),
	}
	return t
}

func (t *windowsTray) Run() error {
	globalTrayMu.Lock()
	globalTray = t
	globalTrayMu.Unlock()

	hInstance, _, _ := procGetModuleHandleW.Call(0)
	className, _ := syscall.UTF16PtrFromString("MoveOpsTrayWindowClass")

	var wc WNDCLASSEXW
	wc.cbSize = uint32(unsafe.Sizeof(wc))
	wc.hInstance = hInstance
	wc.lpszClassName = className
	wc.lpfnWndProc = syscall.NewCallback(wndProc)

	hIcon, _, _ := procLoadIconW.Call(0, IDI_APPLICATION)
	wc.hIcon = hIcon
	wc.hIconSm = hIcon

	res, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if res == 0 && err != syscall.Errno(1410) { // 1410 = ERROR_CLASS_ALREADY_EXISTS
		return fmt.Errorf("falha ao registrar classe de janela do Tray: %w", err)
	}

	windowName, _ := syscall.UTF16PtrFromString(t.cfg.Title)
	hWnd, _, err := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		0,
		0, 0, 0, 0,
		0, // HWND_MESSAGE
		0,
		hInstance,
		0,
	)
	if hWnd == 0 {
		return fmt.Errorf("falha ao criar janela oculta para o Tray: %w", err)
	}
	t.hWnd = hWnd

	// Configura dados da área de notificação
	t.nid.cbSize = uint32(unsafe.Sizeof(t.nid))
	t.nid.hWnd = hWnd
	t.nid.uID = 1
	t.nid.uFlags = NIF_MESSAGE | NIF_ICON | NIF_TIP
	t.nid.uCallbackMessage = WM_TRAYICON
	t.nid.hIcon = hIcon

	tipUTF16, _ := syscall.UTF16FromString(fmt.Sprintf("%s (Porta %d)", t.cfg.Title, t.cfg.Port))
	for i := range t.nid.szTip {
		t.nid.szTip[i] = 0
	}
	copy(t.nid.szTip[:], tipUTF16)
	t.nid.szTip[len(t.nid.szTip)-1] = 0

	r, _, _ := procShellNotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&t.nid)))
	if r == 0 {
		return fmt.Errorf("falha ao registrar ícone na bandeja do Windows (Shell_NotifyIconW)")
	}

	// Exibe balão de boas-vindas
	_ = t.Notify("MoveOps Ativo", fmt.Sprintf("Servidor pronto em %s", t.cfg.URL))

	// Loop de mensagens Win32
	var msg MSG
	for {
		select {
		case <-t.stopC:
			procPostMessageW.Call(t.hWnd, WM_DESTROY, 0, 0)
		default:
		}

		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}

	// Limpa ícone ao sair
	procShellNotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&t.nid)))
	return nil
}

func (t *windowsTray) Notify(title, message string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.hWnd == 0 {
		return nil
	}

	t.nid.uFlags |= NIF_INFO
	t.nid.dwInfoFlags = NIIF_INFO

	titleUTF16, _ := syscall.UTF16FromString(title)
	msgUTF16, _ := syscall.UTF16FromString(message)

	// Zera buffers para garantir ausência de vazamento de memória ou sobreposição
	for i := range t.nid.szInfoTitle {
		t.nid.szInfoTitle[i] = 0
	}
	for i := range t.nid.szInfo {
		t.nid.szInfo[i] = 0
	}

	copy(t.nid.szInfoTitle[:], titleUTF16)
	copy(t.nid.szInfo[:], msgUTF16)
	t.nid.szInfoTitle[len(t.nid.szInfoTitle)-1] = 0
	t.nid.szInfo[len(t.nid.szInfo)-1] = 0

	procShellNotifyIconW.Call(NIM_MODIFY, uintptr(unsafe.Pointer(&t.nid)))
	return nil
}

func (t *windowsTray) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	select {
	case <-t.stopC:
	default:
		close(t.stopC)
	}

	if t.hWnd != 0 {
		procPostQuitMessage.Call(0)
	}
}

func wndProc(hWnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	globalTrayMu.Lock()
	t := globalTray
	globalTrayMu.Unlock()

	if t == nil {
		r, _, _ := procDefWindowProcW.Call(hWnd, uintptr(msg), wParam, lParam)
		return r
	}

	switch msg {
	case WM_TRAYICON:
		switch lParam {
		case WM_LBUTTONDBLCLK:
			// Duplo-clique: Abre painel no navegador
			if t.cfg.OnOpen != nil {
				t.cfg.OnOpen()
			}
			return 0

		case WM_RBUTTONUP, WM_CONTEXTMENU:
			// Botão direito: exibe menu popup
			var pt POINT
			procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
			procSetForegroundWindow.Call(hWnd)

			hMenu, _, _ := procCreatePopupMenu.Call()

			strOpen, _ := syscall.UTF16PtrFromString("🌐 Abrir Painel no Navegador")
			strStatus, _ := syscall.UTF16PtrFromString("📊 Status do MoveOps")
			strToggle, _ := syscall.UTF16PtrFromString("⏸️ Pausar / Retomar Migração")
			strQuit, _ := syscall.UTF16PtrFromString("❌ Encerrar MoveOps")

			procAppendMenuW.Call(hMenu, MF_STRING, IDM_OPEN, uintptr(unsafe.Pointer(strOpen)))
			procAppendMenuW.Call(hMenu, MF_STRING, IDM_STATUS, uintptr(unsafe.Pointer(strStatus)))
			procAppendMenuW.Call(hMenu, MF_STRING, IDM_TOGGLE, uintptr(unsafe.Pointer(strToggle)))
			procAppendMenuW.Call(hMenu, MF_SEPARATOR, 0, 0)
			procAppendMenuW.Call(hMenu, MF_STRING, IDM_QUIT, uintptr(unsafe.Pointer(strQuit)))

			cmd, _, _ := procTrackPopupMenu.Call(
				hMenu,
				TPM_RETURNCMD|TPM_BOTTOMALIGN|TPM_LEFTALIGN,
				uintptr(pt.X),
				uintptr(pt.Y),
				0,
				hWnd,
				0,
			)
			procDestroyMenu.Call(hMenu)

			switch cmd {
			case IDM_OPEN:
				if t.cfg.OnOpen != nil {
					t.cfg.OnOpen()
				}
			case IDM_STATUS:
				status := "Pronto / Ocioso"
				if t.cfg.OnStatus != nil {
					status = t.cfg.OnStatus()
				}
				_ = t.Notify("Status do MoveOps", status)
			case IDM_TOGGLE:
				if t.cfg.OnToggle != nil {
					_, msg := t.cfg.OnToggle()
					_ = t.Notify("Controle de Migração", msg)
				}
			case IDM_QUIT:
				if t.cfg.OnQuit != nil {
					t.cfg.OnQuit()
				}
				t.Stop()
			}
			return 0
		}

	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	}

	r, _, _ := procDefWindowProcW.Call(hWnd, uintptr(msg), wParam, lParam)
	return r
}

// openBrowserPlatform invoca ShellExecuteW diretamente da shell32.dll sem passar por cmd.exe.
// Previne Command Injection e elimina a janela intermitente de prompt.
func openBrowserPlatform(rawURL string) error {
	verb, _ := syscall.UTF16PtrFromString("open")
	urlPtr, _ := syscall.UTF16PtrFromString(rawURL)
	ret, _, err := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(urlPtr)),
		0,
		0,
		1, // SW_SHOWNORMAL
	)
	if ret <= 32 {
		return fmt.Errorf("ShellExecuteW falhou ao abrir URL no Windows (código %d): %w", ret, err)
	}
	return nil
}
