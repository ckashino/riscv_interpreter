// instructions are seperated based on how they are written in assembly and not
// their machine code format (i.e. R, I, ...).

package riscv

type Instr interface {
	Operate(cpu *CPU)
}

type NoOp struct {
	reason string
}

func (instr *NoOp) Operate(cpu *CPU) {
	cpu.PC += 4
}

var threePtInstrTypes = []string{
	"add",
	"sub",
	"mul",
	"div",
	"rem",
	"and",
	"or",
	"xor",
	"sll",
	"srl",
	"sra",
}

type InstrThreePt struct {
	rd  int8
	rs1 int8
	rs2 int8
	op  func(int32, int32) int32
}

func (instr *InstrThreePt) Operate(cpu *CPU) {
	if instr.rd != 0 {
		cpu.Registers[instr.rd] = instr.op(
			cpu.Registers[instr.rs1],
			cpu.Registers[instr.rs2],
		)
	}
	cpu.PC += 4
}

var threePtImmInstrTypes = []string{
	"addi",
	"andi",
	"ori",
	"xori",
	"slli",
	"srli",
	"srai",
}

type InstrThreePtImm struct {
	rd  int8
	rs1 int8
	imm int32
	op  func(int32, int32) int32
}

func (instr *InstrThreePtImm) Operate(cpu *CPU) {
	if instr.rd != 0 {
		cpu.Registers[instr.rd] = instr.op(
			cpu.Registers[instr.rs1],
			instr.imm,
		)
	}
	cpu.PC += 4
}

var loadImmInstrTypes = []string{
	"li",
	"lui",
	"auipc",
}

type LoadImmInstr struct {
	rd  int8
	imm int32
	op  func(*CPU, int32) int32
}

func (instr *LoadImmInstr) Operate(cpu *CPU) {
	if instr.rd != 0 {
		cpu.Registers[instr.rd] = instr.op(cpu, instr.imm)
	}
	cpu.PC += 4
}

var loadInstrTypes = []string{
	"lw",
	"lh",
	"lhu",
	"lb",
	"lbu",
}

type LoadInstr struct {
	rd  int8
	rs1 int8
	imm int32
	op  func(*CPU, int32, int32) int32
}

func (instr *LoadInstr) Operate(cpu *CPU) {
	if instr.rd != 0 {
		cpu.Registers[instr.rd] = instr.op(cpu, cpu.Registers[instr.rs1], instr.imm)
	}
	cpu.PC += 4
}

var storeInstrTypes = []string{
	"sw",
	"sh",
	"sb",
}

type StoreInstr struct {
	rs1 int8
	rs2 int8
	imm int32
	op  func(*CPU, int32, int32, int32)
}

func (instr *StoreInstr) Operate(cpu *CPU) {
	instr.op(cpu, cpu.Registers[instr.rs1], cpu.Registers[instr.rs2], instr.imm)
	cpu.PC += 4
}

var branchThreeInstrTypes = []string{
	"beq",
	"bne",
	"blt",
	"bltu",
	"bgt",
	"bgtu",
	"ble",
	"bleu",
	"bge",
	"bgeu",
}

type BranchThreeInstr struct {
	rs1         int8
	rs2         int8
	destination string
	op          func(*CPU, int32, int32, string)
}

func (instr *BranchThreeInstr) Operate(cpu *CPU) {
	instr.op(cpu, cpu.Registers[instr.rs1], cpu.Registers[instr.rs2], instr.destination)
}

var branchTwoInstrTypes = []string{
	"beqz",
	"bnez",
	"bltz",
	"bgtz",
	"blez",
	"bgez",
}

type BranchTwoInstr struct {
	rs1         int8
	destination string
	op          func(*CPU, int32, string)
}

func (instr *BranchTwoInstr) Operate(cpu *CPU) {
	instr.op(cpu, cpu.Registers[instr.rs1], instr.destination)
}

type JumpInstr struct {
	destination string
}

func (instr *JumpInstr) Operate(cpu *CPU) {
	cpu.PC += uint32(immOrLabel(cpu, instr.destination))
}

type JumpAndLinkInstr struct {
	destination string
	rd          int8
}

func (instr *JumpAndLinkInstr) Operate(cpu *CPU) {
	if instr.rd != 0 {
		cpu.Registers[instr.rd] = int32(cpu.PC) + 4
	}
	cpu.PC += uint32(immOrLabel(cpu, instr.destination))
}

type JumpAndLinkRInstr struct {
	imm int32
	rd  int8
	rs1 int8
}

func (instr *JumpAndLinkRInstr) Operate(cpu *CPU) {
	if instr.rd != 0 {
		cpu.Registers[instr.rd] = int32(cpu.PC) + 4
	}
	cpu.PC = uint32(int(instr.imm) + int(cpu.Registers[instr.rs1]))
}

var setInstrTypes = []string{
	"slt",
	"sltu",
}

type SetInstr struct {
	rd, rs1, rs2 int8
	op           func(int32, int32) bool
}

func (instr *SetInstr) Operate(cpu *CPU) {
	if instr.rd != 0 {
		if instr.op(cpu.Registers[instr.rs1], cpu.Registers[instr.rs2]) {
			cpu.Registers[instr.rd] = 1
		} else {
			cpu.Registers[instr.rd] = 0
		}
	}
	cpu.PC += 4
}

var setImmInstrTypes = []string{
	"slti",
	"sltiu",
}

type SetImmInstr struct {
	rd, rs1 int8
	imm     int32
	op      func(int32, int32) bool
}

func (instr *SetImmInstr) Operate(cpu *CPU) {
	if instr.rd != 0 {
		if instr.op(cpu.Registers[instr.rs1], instr.imm) {
			cpu.Registers[instr.rd] = 1
		} else {
			cpu.Registers[instr.rd] = 0
		}
	}
	cpu.PC += 4
}
