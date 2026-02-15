//go:build windows

package process

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	version  = windows.NewLazySystemDLL("version.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procEnumWindows               = user32.NewProc("EnumWindows")
	procIsWindowVisible           = user32.NewProc("IsWindowVisible")
	procGetWindowTextLengthW      = user32.NewProc("GetWindowTextLengthW")
	procGetWindowThreadProcessId  = user32.NewProc("GetWindowThreadProcessId")
	procQueryFullProcessImageName = kernel32.NewProc("QueryFullProcessImageNameW")
	procGetCurrentProcessId       = kernel32.NewProc("GetCurrentProcessId")
	procProcessIdToSessionId      = kernel32.NewProc("ProcessIdToSessionId")
	procWTSGetActiveConsoleSessionId = kernel32.NewProc("WTSGetActiveConsoleSessionId")
	procGetFileVersionInfoSize    = version.NewProc("GetFileVersionInfoSizeW")
	procGetFileVersionInfo        = version.NewProc("GetFileVersionInfoW")
	procVerQueryValue             = version.NewProc("VerQueryValueW")

	enumCbOnce uintptr
	enumCbInit sync.Once
)

type WindowsAdapter struct{}

func NewWindowsAdapter() *WindowsAdapter {
	return &WindowsAdapter{}
}

func (a *WindowsAdapter) ListProcesses() ([]ProcessInfo, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	if err := windows.Process32First(snapshot, &entry); err != nil {
		return nil, err
	}

	var procs []ProcessInfo
	for {
		procs = append(procs, ProcessInfo{
			PID:  entry.ProcessID,
			Name: windows.UTF16ToString(entry.ExeFile[:]),
		})
		if err := windows.Process32Next(snapshot, &entry); err != nil {
			break
		}
	}

	return procs, nil
}

func (a *WindowsAdapter) ListApplications() ([]ProcessInfo, error) {
	if currentSessionID() == 0 {
		return listApplicationsForActiveSession()
	}

	pids := visibleWindowPIDs()

	var result []ProcessInfo
	for pid := range pids {
		info, err := processInfoFromPID(pid)
		if err != nil {
			continue
		}
		result = append(result, info)
	}
	return result, nil
}

func currentSessionID() uint32 {
	pid, _, _ := procGetCurrentProcessId.Call()
	var sessionID uint32
	procProcessIdToSessionId.Call(pid, uintptr(unsafe.Pointer(&sessionID)))
	return sessionID
}

func listApplicationsForActiveSession() ([]ProcessInfo, error) {
	activeSession, _, _ := procWTSGetActiveConsoleSessionId.Call()

	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	if err := windows.Process32First(snapshot, &entry); err != nil {
		return nil, err
	}

	var result []ProcessInfo
	for {
		pid := entry.ProcessID
		var sessionID uint32
		ret, _, _ := procProcessIdToSessionId.Call(uintptr(pid), uintptr(unsafe.Pointer(&sessionID)))
		if ret != 0 && uintptr(sessionID) == activeSession {
			if info, err := processInfoFromPID(pid); err == nil {
				result = append(result, info)
			}
		}
		if err := windows.Process32Next(snapshot, &entry); err != nil {
			break
		}
	}
	return result, nil
}

func (a *WindowsAdapter) KillProcess(pid uint32) error {
	handle, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, pid)
	if err != nil {
		return nil
	}
	defer windows.CloseHandle(handle)
	return windows.TerminateProcess(handle, 1)
}

type enumWindowsState struct {
	pids map[uint32]struct{}
}

func enumWindowsProc(hwnd uintptr, lParam uintptr) uintptr {
	if visible, _, _ := procIsWindowVisible.Call(hwnd); visible == 0 {
		return 1
	}
	if titleLen, _, _ := procGetWindowTextLengthW.Call(hwnd); titleLen == 0 {
		return 1
	}
	state := (*enumWindowsState)(unsafe.Pointer(lParam))
	var pid uint32
	procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid != 0 {
		state.pids[pid] = struct{}{}
	}
	return 1
}

func visibleWindowPIDs() map[uint32]struct{} {
	enumCbInit.Do(func() {
		enumCbOnce = windows.NewCallback(enumWindowsProc)
	})

	state := &enumWindowsState{pids: make(map[uint32]struct{})}
	procEnumWindows.Call(enumCbOnce, uintptr(unsafe.Pointer(state)))
	runtime.KeepAlive(state)
	return state.pids
}

func processInfoFromPID(pid uint32) (ProcessInfo, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return ProcessInfo{}, err
	}
	defer windows.CloseHandle(handle)

	var buf [windows.MAX_PATH]uint16
	size := uint32(len(buf))
	ret, _, err := procQueryFullProcessImageName.Call(
		uintptr(handle),
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 {
		return ProcessInfo{}, fmt.Errorf("QueryFullProcessImageName: %w", err)
	}

	path := windows.UTF16ToString(buf[:size])
	return ProcessInfo{
		PID:         pid,
		Name:        filepath.Base(path),
		Path:        path,
		Description: fileDescription(path),
	}, nil
}

func fileDescription(path string) string {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return ""
	}

	size, _, _ := procGetFileVersionInfoSize.Call(uintptr(unsafe.Pointer(pathPtr)), 0)
	if size == 0 {
		return ""
	}

	buf := make([]byte, size)
	ret, _, _ := procGetFileVersionInfo.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		0,
		size,
		uintptr(unsafe.Pointer(&buf[0])),
	)
	if ret == 0 {
		return ""
	}

	var transPtr uintptr
	var transLen uint32
	transBlock, _ := windows.UTF16PtrFromString(`\VarFileInfo\Translation`)
	ret, _, _ = procVerQueryValue.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(transBlock)),
		uintptr(unsafe.Pointer(&transPtr)),
		uintptr(unsafe.Pointer(&transLen)),
	)
	if ret == 0 || transLen < 4 {
		return ""
	}

	lang := *(*uint16)(unsafe.Pointer(transPtr))
	cp := *(*uint16)(unsafe.Pointer(transPtr + 2))
	subBlock, _ := windows.UTF16PtrFromString(
		fmt.Sprintf(`\StringFileInfo\%04X%04X\FileDescription`, lang, cp),
	)

	var descPtr uintptr
	var descLen uint32
	ret, _, _ = procVerQueryValue.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(subBlock)),
		uintptr(unsafe.Pointer(&descPtr)),
		uintptr(unsafe.Pointer(&descLen)),
	)
	if ret == 0 || descLen == 0 {
		return ""
	}

	descSlice := unsafe.Slice((*uint16)(unsafe.Pointer(descPtr)), descLen)
	return windows.UTF16ToString(descSlice)
}
