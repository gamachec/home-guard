//go:build windows

package notify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modWtsapi32               = windows.NewLazySystemDLL("wtsapi32.dll")
	procWTSQueryUserToken     = modWtsapi32.NewProc("WTSQueryUserToken")
	procWTSEnumerateSessionsW = modWtsapi32.NewProc("WTSEnumerateSessionsW")
	procWTSFreeMemory         = modWtsapi32.NewProc("WTSFreeMemory")
)

type wtsSessionInfo struct {
	SessionID      uint32
	WinStationName *uint16
	State          uint32
}

const wtsActive = 0

type SessionNotifier struct{}

func NewSessionNotifier() *SessionNotifier {
	return &SessionNotifier{}
}

func (n *SessionNotifier) Send(notif Notification) error {
	sessionID, err := activeConsoleSessionID()
	if err != nil {
		return err
	}

	var userToken windows.Token
	ret, _, err := procWTSQueryUserToken.Call(uintptr(sessionID), uintptr(unsafe.Pointer(&userToken)))
	if ret == 0 {
		return fmt.Errorf("WTSQueryUserToken: %w", err)
	}
	defer userToken.Close()

	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	payload, err := json.Marshal(notif)
	if err != nil {
		return err
	}
	encoded := base64.StdEncoding.EncodeToString(payload)

	cmdLine, err := windows.UTF16PtrFromString(fmt.Sprintf(`"%s" notify %s`, execPath, encoded))
	if err != nil {
		return err
	}

	var si windows.StartupInfo
	si.Cb = uint32(unsafe.Sizeof(si))
	var pi windows.ProcessInformation

	if err := windows.CreateProcessAsUser(
		userToken, nil, cmdLine,
		nil, nil, false,
		windows.CREATE_NO_WINDOW,
		nil, nil,
		&si, &pi,
	); err != nil {
		return fmt.Errorf("CreateProcessAsUser: %w", err)
	}

	windows.CloseHandle(pi.Process)
	windows.CloseHandle(pi.Thread)
	return nil
}

func activeConsoleSessionID() (uint32, error) {
	var pSessions *wtsSessionInfo
	var count uint32

	ret, _, err := procWTSEnumerateSessionsW.Call(
		0, 0, 1,
		uintptr(unsafe.Pointer(&pSessions)),
		uintptr(unsafe.Pointer(&count)),
	)
	if ret == 0 {
		return 0, fmt.Errorf("WTSEnumerateSessions: %w", err)
	}
	defer procWTSFreeMemory.Call(uintptr(unsafe.Pointer(pSessions)))

	for _, s := range unsafe.Slice(pSessions, count) {
		if s.State == wtsActive {
			return s.SessionID, nil
		}
	}
	return 0, fmt.Errorf("no active user session found")
}
