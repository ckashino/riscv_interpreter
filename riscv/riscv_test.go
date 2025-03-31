package riscv

import (
	"encoding/binary"
	"testing"
)

func TestStoreByte(t *testing.T) {
	cpu := NewCPU(16)
	cpu.LoadInstructions([]string{"li x1, 4", "sb x1, 0(x0)"})
	cpu.RunProgram()
	if cpu.Registers[1] != 4 {
		t.Error("Load Immediate Fail")
	}
	storedVal := cpu.Memory[0]

	if storedVal != 4 {
		t.Error("Store byte failure")
	}
}

func TestStoreHalf(t *testing.T) {
	cpu := NewCPU(16)
	cpu.LoadInstructions([]string{"li x1, 16", "sh x1, 0(x0)"})
	cpu.RunProgram()
	if cpu.Registers[1] != 16 {
		t.Error("Load Immediate Fail")
	}
	storedVal := binary.LittleEndian.Uint16(cpu.Memory[0:])

	if storedVal != 16 {
		t.Error("Store byte failure")
	}
}

func TestStore(t *testing.T) {
	cpu := NewCPU(16)
	cpu.LoadInstructions([]string{"li x1, 16", "sw x1, 4(x0)"})
	cpu.RunProgram()
	if cpu.Registers[1] != 16 {
		t.Error("Load Immediate Fail")
	}
	storedVal := binary.LittleEndian.Uint32(cpu.Memory[4:])

	if storedVal != 16 {
		t.Error("Store byte failure")
	}
}

func TestLoad(t *testing.T) {
	cpu := NewCPU(16)
	cpu.LoadInstructions([]string{"li x1, 16", "sw x1, 4(x0)", "lw x3, 4(x0)"})
	cpu.RunProgram()
	if cpu.Registers[1] != 16 {
		t.Error("Load Immediate Fail")
	}
	storedVal := binary.LittleEndian.Uint32(cpu.Memory[4:])

	if storedVal != 16 {
		t.Error("Store Byte Fail")
	}

	if cpu.Registers[3] != 16 {
		t.Error("Load Fail")
	}
}

func TestLabel(t *testing.T) {
	cpu := NewCPU(16)
	cpu.LoadInstructions([]string{"main:", "li x0, 100"})
	cpu.RunProgram()

	if len(cpu.Labels) != 1 {
		t.Error("Label Add Fail")
	}

	if cpu.Labels["main"] != 4+16 {
		t.Errorf("Label PC fail. actual %d", cpu.Labels["main"])
	}
}
