package tools

/*
#cgo LDFLAGS: -lpcap
#include <pcap.h>
#include <stdlib.h>

// 手动定义 PCAP_DLT_EN10MB 常量
#define PCAP_DLT_EN10MB 1
*/
import "C"
import (
	"unsafe"
)

func BpfFilterValid(bpfFilter string) bool {

	// 转换为 C 字符串
	cFilter := C.CString(bpfFilter)
	defer C.free(unsafe.Pointer(cFilter))

	// 创建一个虚拟的 pcap 句柄
	pcapHandle := C.pcap_open_dead(C.PCAP_DLT_EN10MB, 65535)
	if pcapHandle == nil {
		return false
	}
	defer C.pcap_close(pcapHandle)

	// 编译 BPF 过滤器
	var bpfProgram C.struct_bpf_program
	if C.pcap_compile(pcapHandle, &bpfProgram, cFilter, 1, 0) != 0 {
		return false
	} else {
		C.pcap_freecode(&bpfProgram)
		return true
	}
}
