// Copyright 2016 syzkaller project authors. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package prog

import (
	"bytes"
	"strings"
	"testing"
)

func TestAssignSizeRandom(t *testing.T) {
	target, rs, iters := initTest(t)
	for i := 0; i < iters; i++ {
		p := target.Generate(rs, 10, nil)
		data0 := p.Serialize()
		for _, call := range p.Calls {
			target.assignSizesCall(call)
		}
		if data1 := p.Serialize(); !bytes.Equal(data0, data1) {
			t.Fatalf("different lens assigned, initial:\n%s\nnew:\n%s\n", data0, data1)
		}
		p.Mutate(rs, 10, nil, nil)
		p.Serialize()
		for _, call := range p.Calls {
			target.assignSizesCall(call)
		}
	}
}

func TestAssignSize(t *testing.T) {
	target := initTargetTest(t, "test", "64")
	// nolint: lll
	tests := []struct {
		unsizedProg string
		sizedProg   string
	}{
		{
			"test$length0(&(0x7f0000000000)={0xff, 0x0})",
			"test$length0(&(0x7f0000000000)={0xff, 0x2})",
		},
		{
			"test$length1(&(0x7f0000001000)={0xff, 0x0})",
			"test$length1(&(0x7f0000001000)={0xff, 0x4})",
		},
		{
			"test$length2(&(0x7f0000001000)={0xff, 0x0})",
			"test$length2(&(0x7f0000001000)={0xff, 0x8})",
		},
		{
			"test$length3(&(0x7f0000005000)={0xff, 0x0, 0x0})",
			"test$length3(&(0x7f0000005000)={0xff, 0x4, 0x2})",
		},
		{
			"test$length4(&(0x7f0000003000)={0x0, 0x0})",
			"test$length4(&(0x7f0000003000)={0x2, 0x2})",
		},
		{
			"test$length5(&(0x7f0000002000)={0xff, 0x0})",
			"test$length5(&(0x7f0000002000)={0xff, 0x4})",
		},
		{
			"test$length6(&(0x7f0000002000)={[0xff, 0xff, 0xff, 0xff], 0x0})",
			"test$length6(&(0x7f0000002000)={[0xff, 0xff, 0xff, 0xff], 0x4})",
		},
		{
			"test$length7(&(0x7f0000003000)={[0xff, 0xff, 0xff, 0xff], 0x0})",
			"test$length7(&(0x7f0000003000)={[0xff, 0xff, 0xff, 0xff], 0x8})",
		},
		{
			"test$length8(&(0x7f000001f000)={0x00, {0xff, 0x0, 0x00, [0xff, 0xff, 0xff]}, [{0xff, 0x0, 0x00, [0xff, 0xff, 0xff]}], 0x00, 0x0, [0xff, 0xff]})",
			"test$length8(&(0x7f000001f000)={0x32, {0xff, 0x1, 0x10, [0xff, 0xff, 0xff]}, [{0xff, 0x1, 0x10, [0xff, 0xff, 0xff]}], 0x10, 0x1, [0xff, 0xff]})",
		},
		{
			"test$length9(&(0x7f000001f000)={&(0x7f0000000000/0x5000)=nil, 0x0000})",
			"test$length9(&(0x7f000001f000)={&(0x7f0000000000/0x5000)=nil, 0x5000})",
		},
		{
			"test$length10(&(0x7f0000000000/0x5000)=nil, 0x0000, 0x0000, 0x0000, 0x0000)",
			"test$length10(&(0x7f0000000000/0x5000)=nil, 0x5000, 0x5000, 0x2800, 0x1400)",
		},
		{
			"test$length11(&(0x7f0000000000)={0xff, 0xff, [0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff]}, 0x00)",
			"test$length11(&(0x7f0000000000)={0xff, 0xff, [0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff]}, 0x30)",
		},
		{
			"test$length12(&(0x7f0000000000)={0xff, 0xff, [0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff]}, 0x00)",
			"test$length12(&(0x7f0000000000)={0xff, 0xff, [0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff]}, 0x30)",
		},
		{
			"test$length13(&(0x7f0000000000)={0xff, 0xff, [0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff]}, &(0x7f0000001000)=0x00)",
			"test$length13(&(0x7f0000000000)={0xff, 0xff, [0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff]}, &(0x7f0000001000)=0x30)",
		},
		{
			"test$length14(&(0x7f0000000000)={0xff, 0xff, [0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff]}, &(0x7f0000001000)=0x00)",
			"test$length14(&(0x7f0000000000)={0xff, 0xff, [0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff]}, &(0x7f0000001000)=0x30)",
		},
		{
			"test$length15(0xff, 0x0)",
			"test$length15(0xff, 0x2)",
		},
		{
			"test$length16(&(0x7f0000000000)={[0x42, 0x42], 0xff, 0xff, 0xff, 0xff, 0xff})",
			"test$length16(&(0x7f0000000000)={[0x42, 0x42], 0x2, 0x10, 0x8, 0x4, 0x2})",
		},
		{
			"test$length17(&(0x7f0000000000)={0x42, 0xff, 0xff, 0xff, 0xff})",
			"test$length17(&(0x7f0000000000)={0x42, 0x8, 0x4, 0x2, 0x1})",
		},
		{
			"test$length18(&(0x7f0000000000)={0x42, 0xff, 0xff, 0xff, 0xff})",
			"test$length18(&(0x7f0000000000)={0x42, 0x8, 0x4, 0x2, 0x1})",
		},
		{
			"test$length19(&(0x7f0000000000)={{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}, 0xff, 0xff, 0xff})",
			"test$length19(&(0x7f0000000000)={{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x14}, 0x14, 0x14, 0x5})",
		},
		{
			"test$length20(&(0x7f0000000000)={{{0xff, 0xff, 0xff, 0xff}, 0xff, 0xff, 0xff}, 0xff, 0xff})",
			"test$length20(&(0x7f0000000000)={{{0x4, 0x4, 0x7, 0x9}, 0x7, 0x7, 0x9}, 0x9, 0x9})",
		},
		{
			"test$length21(&(0x7f0000000000)=0x0, 0x0)",
			"test$length21(&(0x7f0000000000), 0x40)",
		},
		{
			"test$length22(&(0x7f0000000000)='12345', 0x0)",
			"test$length22(&(0x7f0000000000)='12345', 0x28)",
		},
		{
			"test$length23(&(0x7f0000000000)={0x1, {0x2, 0x0}})",
			"test$length23(&(0x7f0000000000)={0x1, {0x2, 0x6}})",
		},
		{
			"test$length24(&(0x7f0000000000)={{0x0, {0x0}}, {0x0, {0x0}}})",
			"test$length24(&(0x7f0000000000)={{0x0, {0x8}}, {0x0, {0x10}}})",
		},
		{
			"test$length26(&(0x7f0000000000), 0x0)",
			"test$length26(&(0x7f0000000000), 0x8)",
		},
		{
			"test$length27(&(0x7f0000000000), 0x0)",
			"test$length27(&(0x7f0000000000), 0x2a)",
		},
		{
			"test$length28(&(0x7f0000000000), 0x0)",
			"test$length28(&(0x7f0000000000), 0x2a)",
		},
		{
			"test$length29(&(0x7f0000000000)={'./a\\x00', './b/c\\x00', 0x0, 0x0, 0x0})",
			"test$length29(&(0x7f0000000000)={'./a\\x00', './b/c\\x00', 0xa, 0x14, 0x21})",
		},
	}

	for i, test := range tests {
		p, err := target.Deserialize([]byte(test.unsizedProg), Strict)
		if err != nil {
			t.Fatalf("failed to deserialize prog %v: %v", i, err)
		}
		for _, call := range p.Calls {
			target.assignSizesCall(call)
		}
		p1 := strings.TrimSpace(string(p.Serialize()))
		if p1 != test.sizedProg {
			t.Fatalf("failed to assign sizes in prog %v\ngot  %v\nwant %v", i, p1, test.sizedProg)
		}
	}
}
