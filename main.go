package main

import (
	"context"
	_ "embed"
	"errors"
	"log"
	"os"

	"github.com/jerbob92/wazero-function-pointer-test/imports"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/experimental"
	"github.com/tetratelabs/wazero/experimental/logging"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed test.wasm
var invokeWasm []byte

func main() {
	// Choose the context to use for function calls.
	// Set context to one that has an experimental listener
	ctx := context.WithValue(context.Background(), experimental.FunctionListenerFactoryKey{}, logging.NewLoggingListenerFactory(os.Stdout))

	// Uncomment to disable tracing
	//ctx = context.Background()
	// Create a new WebAssembly Runtime.
	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfigInterpreter())
	defer r.Close(ctx) // This closes everything this Runtime created.

	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		log.Panicln(err)
	}

	// Add missing emscripten and syscalls
	if _, err := imports.Instantiate(ctx, r); err != nil {
		log.Panicln(err)
	}

	compiled, err := r.CompileModule(ctx, invokeWasm)
	if err != nil {
		log.Panicln(err)
	}

	mod, err := r.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().WithStartFunctions("_initialize").WithFS(os.DirFS("")).WithStdout(os.Stdout).WithStderr(os.Stderr))
	if err != nil {
		log.Panicln(err)
	}

	// Open just file reader to demo this.
	testFile, err := os.Open("pdf-test.pdf")
	if err != nil {
		log.Panicln(err)
	}

	defer testFile.Close()

	// Get filesize
	statFile, err := testFile.Stat()
	if err != nil {
		log.Panicln(err)
	}

	// 1 is just a dummy value here, we use it to keep track of file readers across the application.
	fileIndex := uint32(1)
	imports.OpenFiles[fileIndex] = testFile

	malloc := mod.ExportedFunction("malloc")

	// Memory allocation for a FPDF_FILEACCESS.
	// Size of m_FileLen, m_GetBlock, m_Param
	ret, err := malloc.Call(ctx, 12)
	if err != nil {
		log.Panicln(err)
	}

	FPDF_FILEACCESSPointer := ret[0]

	fileSize := uint32(statFile.Size())

	// Write m_FileLen to FPDF_FILEACCESS struct.
	if !mod.Memory().WriteUint32Le(ctx, uint32(FPDF_FILEACCESSPointer), fileSize) {
		log.Panicln(errors.New("could not write file len data to memory"))
	}

	// Write m_GetBlock function pointer to FPDF_FILEACCESS struct.
	// @todo: how to get a pointer to the function?
	// @todo: I basically want to get the index of `FPDF_LoadCustomDocument_m_GetBlock` registered in imports/imports.go in the function table here.
	// @todo: See imports/functionpointer.go for the implementation of the function pointer.
	//if !mod.Memory().WriteUint32Le(ctx, uint32(FPDF_FILEACCESSPointer+4), functionPointer) {
	//	log.Panicln(errors.New("could not write file len data to memory"))
	//}

	// Memory allocation for a m_Param.
	ret, err = malloc.Call(ctx, 4)
	if err != nil {
		log.Panicln(err)
	}

	m_ParamPointer := ret[0]

	// Write value to m_Param to keep track of which file reader to use.
	if !mod.Memory().WriteUint32Le(ctx, uint32(m_ParamPointer), fileIndex) {
		log.Panicln(errors.New("could not write file len data to memory"))
	}

	// Write pointer to m_Param to FPDF_FILEACCESS struct.
	if !mod.Memory().WriteUint32Le(ctx, uint32(FPDF_FILEACCESSPointer+8), uint32(m_ParamPointer)) {
		log.Panicln(errors.New("could not write file len data to memory"))
	}

	// Call FPDF_LoadCustomDocument with a pointer to the FPDF_FILEACCESS.
	// This will loop through all bytes of te file reader and call them one by one.
	ret, err = mod.ExportedFunction("FPDF_LoadCustomDocument").Call(ctx, FPDF_FILEACCESSPointer)
	if err != nil {
		log.Panicln(err)
	}

	log.Println(ret)
}
