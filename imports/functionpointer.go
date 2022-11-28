package imports

import (
	"context"
	"github.com/tetratelabs/wazero/api"
	"os"
)

var OpenFiles = map[uint32]*os.File{}

type FPDF_LoadCustomDocument_m_GetBlock struct {
}

func (FPDF_LoadCustomDocument_m_GetBlock) Call(ctx context.Context, mod api.Module, params []uint64) []uint64 {
	paramPointer := params[0]
	position := params[1]
	pBufPointer := params[2]
	size := params[3]

	mem := mod.Memory()

	param, ok := mem.ReadUint32Le(ctx, uint32(paramPointer))
	if !ok {
		return []uint64{0}
	}

	// Check if we have the file referenced in param.
	openFile, ok := OpenFiles[param]
	if !ok {
		return []uint64{0}
	}

	// Seek to the right position.
	_, err := openFile.Seek(int64(position), 0)
	if err != nil {
		return []uint64{0}
	}

	// Read the requested data into a buffer.
	readBuffer := make([]byte, size)
	n, err := openFile.Read(readBuffer)
	if n == 0 || err != nil {
		return []uint64{0}
	}

	ok = mem.Write(ctx, uint32(pBufPointer), readBuffer)
	if err != nil {
		return []uint64{0}
	}

	return []uint64{uint64(n)}
}
