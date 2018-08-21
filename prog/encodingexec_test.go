// Copyright 2016 syzkaller project authors. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package prog

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/google/syzkaller/sys"
)

func TestSerializeForExecRandom(t *testing.T) {
	rs, iters := initTest(t)
	for i := 0; i < iters; i++ {
		p := Generate(rs, 10, nil)
		p.SerializeForExec()
	}
}

func TestSerializeForExec(t *testing.T) {
	// A brief recap of exec format.
	// Exec format is an sequence of uint64's which encodes a sequence of calls.
	// The sequence is terminated by a speciall call ExecInstrEOF.
	// Each call is (call ID, number of arguments, arguments...).
	// Each argument is (type, size, value).
	// There are 3 types of arguments:
	//  - ExecArgConst: value is const value
	//  - ExecArgResult: value is index of a call whose result we want to reference
	//  - ExecArgData: value is a binary blob (represented as ]size/8[ uint64's)
	// There are 2 other special call:
	//  - ExecInstrCopyin: copies its second argument into address specified by first argument
	//  - ExecInstrCopyout: reads value at address specified by first argument (result can be referenced by ExecArgResult)
	const (
		instrEOF     = uint64(ExecInstrEOF)
		instrCopyin  = uint64(ExecInstrCopyin)
		instrCopyout = uint64(ExecInstrCopyout)
		argConst     = uint64(ExecArgConst)
		argResult    = uint64(ExecArgResult)
		argData      = uint64(ExecArgData)
	)
	callID := func(name string) uint64 {
		c := sys.CallMap[name]
		if c == nil {
			t.Fatalf("unknown syscall %v", name)
		}
		return uint64(c.ID)
	}
	tests := []struct {
		prog       string
		serialized []uint64
	}{
		{
			"syz_test()",
			[]uint64{
				callID("syz_test"), 0,
				instrEOF,
			},
		},
		{
			"syz_test$int(0x1, 0x2, 0x3, 0x4, 0x5)",
			[]uint64{
				callID("syz_test$int"), 5, argConst, 8, 1, argConst, 1, 2, argConst, 2, 3, argConst, 4, 4, argConst, 8, 5,
				instrEOF,
			},
		},
		{
			"syz_test$align0(&(0x7f0000000000)={0x1, 0x2, 0x3, 0x4, 0x5})",
			[]uint64{
				instrCopyin, dataOffset + 0, argConst, 2, 1,
				instrCopyin, dataOffset + 2, argConst, 4, 2,
				instrCopyin, dataOffset + 6, argConst, 1, 3,
				instrCopyin, dataOffset + 7, argConst, 2, 4,
				instrCopyin, dataOffset + 9, argConst, 8, 5,
				callID("syz_test$align0"), 1, argConst, ptrSize, dataOffset,
				instrEOF,
			},
		},
	}

	for i, test := range tests {
		p, err := Deserialize([]byte(test.prog))
		if err != nil {
			t.Fatalf("failed to deserialize prog %v: %v", i, err)
		}
		t.Run(fmt.Sprintf("%v:%v", i, p.String()), func(t *testing.T) {
			data := p.SerializeForExec()
			w := new(bytes.Buffer)
			binary.Write(w, binary.LittleEndian, test.serialized)
			if !bytes.Equal(data, w.Bytes()) {
				got := make([]uint64, len(data)/8)
				binary.Read(bytes.NewReader(data), binary.LittleEndian, &got)
				t.Logf("want: %v", test.serialized)
				t.Logf("got:  %v", got)
				t.Fatalf("mismatch")
			}

		})
	}
}
