/*
Copyright (c) Facebook, Inc. and its affiliates.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Package blazesym provides a CGO wrapper around the Blazesym library.
package blazesym

/*
#include <blazesym_c/blazesym_c.h>
// The generated struct in cgo does not contain syms for blazesym result
// see:
//type _Ctype_struct_blaze_result struct {
//	cnt _Ctype_size_t
//}
// Adding a C function to return syms from blaze_result
struct blaze_sym get_result(blaze_result* res) {
	return res->syms[0];
}
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

const UnknownSymbol string = "[Unknown]"

// Symbolizer represents a Blazesym symbolizer.
type Symbolizer struct {
	s *C.blaze_symbolizer
}

// BlazeErr represents an error code from the blazesym library.
type BlazeErr int

const (
	BLAZE_ERR_OK                BlazeErr = 0
	BLAZE_ERR_NOT_FOUND         BlazeErr = -2
	BLAZE_ERR_PERMISSION_DENIED BlazeErr = -1
	BLAZE_ERR_ALREADY_EXISTS    BlazeErr = -17
	BLAZE_ERR_WOULD_BLOCK       BlazeErr = -11
	BLAZE_ERR_INVALID_DATA      BlazeErr = -22
	BLAZE_ERR_TIMED_OUT         BlazeErr = -110
	BLAZE_ERR_UNSUPPORTED       BlazeErr = -95
	BLAZE_ERR_OUT_OF_MEMORY     BlazeErr = -12
	BLAZE_ERR_INVALID_INPUT     BlazeErr = -256
	BLAZE_ERR_WRITE_ZERO        BlazeErr = -257
	BLAZE_ERR_UNEXPECTED_EOF    BlazeErr = -258
	BLAZE_ERR_INVALID_DWARF     BlazeErr = -259
	BLAZE_ERR_OTHER             BlazeErr = -260
)

func (e BlazeErr) Error() error {
	return errors.New(C.GoString(C.blaze_err_str(C.enum_blaze_err(e))))
}

// NewSymbolizer returns a new Blazesym symbolizer.
func NewSymbolizer() (*Symbolizer, error) {
	s := C.blaze_symbolizer_new()
	if s == nil {
		return nil, fmt.Errorf("failed to create symbolizer")
	}
	return &Symbolizer{s: s}, nil
}

// Symbolize symbolizes an address using the Blazesym symbolizer.
func (s *Symbolizer) Symbolize(pid uint32, addr uint64) (string, error) {
	caddr := C.ulong(addr)
	symProcess := &C.struct_blaze_symbolize_src_process{}
	symProcess.type_size = C.ulong(unsafe.Sizeof(symProcess))
	symProcess.pid = C.uint32_t(pid)
	symbolized := C.blaze_symbolize_process_abs_addrs(s.s, symProcess, &caddr, 1)
	lastErr := BlazeErr(C.blaze_err_last())
	if lastErr != BLAZE_ERR_OK {
		return UnknownSymbol, lastErr.Error()
	}
	if symbolized.cnt == 0 {
		return UnknownSymbol, nil
	}
	symbolizedResult := C.get_result(symbolized)
	name := C.GoString(symbolizedResult.name)
	if len(name) == 0 {
		C.blaze_result_free(symbolized)
		return fmt.Sprintf("%d : %d <No symbol>", symbolizedResult.addr, symbolizedResult.offset), nil
	}
	//offset := symbolizedResult.offset
	//dir := C.GoString(symbolizedResult.code_info.dir)
	//file := C.GoString(symbolizedResult.code_info.file)
	//line := symbolizedResult.code_info.line
	//column := symbolizedResult.code_info.column
	C.blaze_result_free(symbolized)
	return name, nil
}

// Close closes the Blazesym symbolizer.
func (s *Symbolizer) Close() {
	C.blaze_symbolizer_free(s.s)
}
