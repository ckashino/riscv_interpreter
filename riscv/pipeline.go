package riscv

import (
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type CPU struct {
	PC            uint32
	Registers     [32]int32
	Memory        []byte
	MemorySize    uint32
	instructions  []string
	Done          bool
	Labels        map[string]uint32
	MemoryHistory []string
	entryPoint    string
}

var abiToRegister = map[string]int{
	"zero": 0, "x0": 0,
	"ra": 1, "x1": 1,
	"sp": 2, "x2": 2,
	"gp": 3, "x3": 3,
	"tp": 4, "x4": 4,
	"t0": 5, "x5": 5,
	"t1": 6, "x6": 6,
	"t2": 7, "x7": 7,
	"fp": 8, "s0": 8, "x8": 8,
	"s1": 9, "x9": 9,
	"a0": 10, "x10": 10,
	"a1": 11, "x11": 11,
	"a2": 12, "x12": 12,
	"a3": 13, "x13": 13,
	"a4": 14, "x14": 14,
	"a5": 15, "x15": 15,
	"a6": 16, "x16": 16,
	"a7": 17, "x17": 17,
	"s2": 18, "x18": 18,
	"s3": 19, "x19": 19,
	"s4": 20, "x20": 20,
	"s5": 21, "x21": 21,
	"s6": 22, "x22": 22,
	"s7": 23, "x23": 23,
	"s8": 24, "x24": 24,
	"s9": 25, "x25": 25,
	"s10": 26, "x26": 26,
	"s11": 27, "x27": 27,
	"t3": 28, "x28": 28,
	"t4": 29, "x29": 29,
	"t5": 30, "x30": 30,
	"t6": 31, "x31": 31,
}

func getRegisterNumber(abiName string) int8 {
	var reg int
	var ok bool

	if reg, ok = abiToRegister[abiName]; !ok {
		panic(fmt.Sprintf("invalid register: %s", abiName))
	}

	return int8(reg)
}

func parseImm(imm_str string) int32 {
	if imm, err := strconv.Atoi(imm_str); err == nil {
		return int32(imm)
	} else {
		panic(fmt.Sprintf("immediate parse error: %s", imm_str))
	}
}

func NewCPU(memorySize uint32) CPU {

	cpu := CPU{
		Memory:     make([]byte, memorySize),
		Labels:     make(map[string]uint32),
		MemorySize: memorySize,
		PC:         16,
	}

	cpu.Registers[abiToRegister["sp"]] = int32(memorySize)

	return cpu
}

func (cpu *CPU) checkMemoryAccess(address uint32) error {
	if address+4 > uint32(len(cpu.Memory)) {
		return errors.New("invalid memory access")
	}

	return nil
}

func (cpu *CPU) LoadInstructions(instrs []string) {
	labelRe := regexp.MustCompile(`(.+):`)
	globalRe := regexp.MustCompile(`.global\s(\w+)`)

	cpu.entryPoint = ""

	for i, instr := range instrs {
		labelMatch := labelRe.FindStringSubmatch(instr)
		if len(labelMatch) == 2 {
			label := labelMatch[1]
			cpu.Labels[label] = uint32((i * 4) + 16 + 4)
			continue
		}

		globalMatch := globalRe.FindStringSubmatch(instr)
		if len(globalMatch) == 2 {
			cpu.entryPoint = globalMatch[1]
		}
	}

	cpu.instructions = instrs

	instr_num := int((cpu.PC - 16) / 4)

	if instr_num > (len(cpu.instructions) - 1) {
		cpu.Done = true
	} else {
		cpu.Done = false
	}
}

func (cpu *CPU) RunProgram() {
	for !cpu.Done {
		cpu.RunNextInstruction()
	}

	cpu.PC = 16
	cpu.Done = false
}

func (cpu *CPU) RunNextInstruction() {
	if cpu.PC < 16 {
		cpu.Done = true
		return
	}

	instr_num := int((cpu.PC - 16) / 4)

	if instr_num > (len(cpu.instructions) - 1) {
		cpu.Done = true
		return
	}

	instr := DecodeInstr(&cpu.instructions[instr_num])

	switch v := instr.(type) {
	case *NoOp:
		print(v.reason)
	}
	instr.Operate(cpu)
}

func (cpu *CPU) GetCurrInstr() string {
	instr_num := int((cpu.PC - 16) / 4)

	if instr_num < (len(cpu.instructions)) && instr_num >= 0 {
		return cpu.instructions[instr_num]
	} else {
		return ""
	}
}

var instrToThreePtOp = map[string]func(int32, int32) int32{
	"add": func(a, b int32) int32 { return a + b },
	"sub": func(a, b int32) int32 { return a - b },
	"mul": func(a, b int32) int32 { return a * b },
	"div": func(a, b int32) int32 { return a / b },
	"rem": func(a, b int32) int32 { return a % b },
	"and": func(a, b int32) int32 { return a & b },
	"or":  func(a, b int32) int32 { return a | b },
	"xor": func(a, b int32) int32 { return a ^ b },
	"sll": func(a, b int32) int32 { return a << b },
	"srl": func(a, b int32) int32 { return int32(uint32(a) >> b) },
	"sra": func(a, b int32) int32 { return a >> b },
}

func parseThreePt(tokens []string) Instr {
	var ok bool

	op, ok := instrToThreePtOp[tokens[0]]

	if !ok {
		panic(fmt.Sprintf("invalid operation: %s", tokens[0]))
	}

	instr := InstrThreePt{
		rd:  getRegisterNumber(tokens[1]),
		rs1: getRegisterNumber(tokens[2]),
		rs2: getRegisterNumber(tokens[3]),
		op:  op,
	}

	return &instr
}

var instrToThreePtImmOp = map[string]func(int32, int32) int32{
	"addi": func(a, b int32) int32 { return a + b },
	"andi": func(a, b int32) int32 { return a & b },
	"ori":  func(a, b int32) int32 { return a | b },
	"xori": func(a, b int32) int32 { return a ^ b },
	"slli": func(a, b int32) int32 { return a << b },
	"srli": func(a, b int32) int32 { return int32(uint32(a) >> b) },
	"srai": func(a, b int32) int32 { return a >> b },
}

func parseThreePtImm(tokens []string) Instr {
	var ok bool

	op, ok := instrToThreePtImmOp[tokens[0]]

	if !ok {
		return &NoOp{reason: "Invalid Operation"}
	}

	instr := InstrThreePtImm{
		rd:  getRegisterNumber(tokens[1]),
		rs1: getRegisterNumber(tokens[2]),
		imm: parseImm(tokens[3]),
		op:  op,
	}

	return &instr
}

var instrToLoadImmOp = map[string]func(*CPU, int32) int32{
	"li":    func(cpu *CPU, imm int32) int32 { return imm },
	"lui":   func(cpu *CPU, imm int32) int32 { return imm << 12 },
	"auipc": func(cpu *CPU, imm int32) int32 { return int32(cpu.PC) + imm<<12 },
}

func parseLoadImm(tokens []string) Instr {

	var ok bool

	op, ok := instrToLoadImmOp[tokens[0]]

	if !ok {
		return &NoOp{reason: "Invalid Operation"}
	}

	instr := LoadImmInstr{
		rd:  getRegisterNumber(tokens[1]),
		imm: parseImm(tokens[2]),
		op:  op,
	}

	return &instr
}

func (cpu *CPU) loadWord(address uint32) int32 {
	err := cpu.checkMemoryAccess(address)
	if err != nil {
		return 0
	}

	value := int32(binary.LittleEndian.Uint32(cpu.Memory[address:]))
	cpu.MemoryHistory = append([]string{fmt.Sprintf("Loaded word (%d) from  %d", value, address)}, cpu.MemoryHistory...)
	return value
}

func (cpu *CPU) loadHalf(address uint32) uint16 {
	err := cpu.checkMemoryAccess(address)
	if err != nil {
		return 0
	}
	value := binary.LittleEndian.Uint16(cpu.Memory[address:])
	cpu.MemoryHistory = append([]string{fmt.Sprintf("Loaded half (%d) from  %d", value, address)}, cpu.MemoryHistory...)
	return value
}

func (cpu *CPU) loadByte(address uint32) uint8 {
	err := cpu.checkMemoryAccess(address)
	if err != nil {
		return 0
	}
	value := uint8(cpu.Memory[address])
	cpu.MemoryHistory = append([]string{fmt.Sprintf("Loaded byte (%d) from  %d", value, address)}, cpu.MemoryHistory...)
	return value
}

var instrToLoadOp = map[string]func(*CPU, int32, int32) int32{
	"lw": func(cpu *CPU, rs1_val int32, imm int32) int32 { return cpu.loadWord(uint32(rs1_val + imm)) },
	"lh": func(cpu *CPU, rs1_val int32, imm int32) int32 {
		return int32(cpu.loadHalf(uint32(rs1_val + imm)))
	},
	"lhu": func(cpu *CPU, rs1_val int32, imm int32) int32 {
		return int32(uint32(cpu.loadHalf(uint32(rs1_val + imm))))
	},
	"lb": func(cpu *CPU, rs1_val int32, imm int32) int32 {
		return int32(cpu.loadByte(uint32(rs1_val + imm)))
	},
	"lbu": func(cpu *CPU, rs1_val int32, imm int32) int32 {
		return int32(uint32(cpu.loadByte(uint32(rs1_val + imm))))
	},
}

func parseLoad(tokens []string) Instr {
	var ok bool

	op, ok := instrToLoadOp[tokens[0]]

	if !ok {
		return &NoOp{reason: "Invalid Operation"}
	}

	instr := LoadInstr{
		rd:  getRegisterNumber(tokens[1]),
		rs1: getRegisterNumber(tokens[3]),
		imm: parseImm(tokens[2]),
		op:  op,
	}

	return &instr
}

func (cpu *CPU) storeWord(address uint32, value int32) {
	err := cpu.checkMemoryAccess(address)
	if err != nil {
		return
	}

	cpu.MemoryHistory = append([]string{fmt.Sprintf("Stored word (%d) to address %d", value, address)}, cpu.MemoryHistory...)
	binary.LittleEndian.PutUint32(cpu.Memory[address:], uint32(value))
}

func (cpu *CPU) storeHalf(address uint32, value int32) {
	err := cpu.checkMemoryAccess(address)
	if err != nil {
		return
	}
	cpu.MemoryHistory = append([]string{fmt.Sprintf("Stored half-word (%d) to address %d", value, address)}, cpu.MemoryHistory...)
	binary.LittleEndian.PutUint16(cpu.Memory[address:], uint16(value))
}

func (cpu *CPU) storeByte(address uint32, value int32) {
	err := cpu.checkMemoryAccess(address)
	if err != nil {
		return
	}

	cpu.MemoryHistory = append([]string{fmt.Sprintf("Stored byte (%d) to address %d", value, address)}, cpu.MemoryHistory...)
	cpu.Memory[address] = uint8(value)
}

var instrToStoreOp = map[string]func(*CPU, int32, int32, int32){
	"sw": func(cpu *CPU, rs1_val int32, rs2_val int32, imm int32) { cpu.storeWord(uint32(imm+rs1_val), rs2_val) },
	"sh": func(cpu *CPU, rs1_val int32, rs2_val int32, imm int32) { cpu.storeHalf(uint32(imm+rs1_val), rs2_val) },
	"sb": func(cpu *CPU, rs1_val int32, rs2_val int32, imm int32) { cpu.storeByte(uint32(imm+rs1_val), rs2_val) },
}

func parseStore(tokens []string) Instr {
	var ok bool

	op, ok := instrToStoreOp[tokens[0]]

	if !ok {
		return &NoOp{reason: "Invalid Operation"}
	}

	instr := StoreInstr{
		rs1: getRegisterNumber(tokens[3]),
		rs2: getRegisterNumber(tokens[1]),
		imm: parseImm(tokens[2]),
		op:  op,
	}

	return &instr
}

func immOrLabel(cpu *CPU, destination string) int {
	var imm int
	var targetAddr uint32
	var err error
	var ok bool

	if imm, err = strconv.Atoi(destination); err != nil {
		if targetAddr, ok = cpu.Labels[destination]; !ok {
			panic("Invalid Jump")
		}
		imm = int(targetAddr) - int(cpu.PC)
	}
	return imm
}

func trueOrNext(cpu *CPU, valid bool, destination string) {
	if valid {
		cpu.PC += uint32(immOrLabel(cpu, destination))
	} else {
		cpu.PC += 4
	}
}

var instrToBranchThreeOp = map[string]func(*CPU, int32, int32, string){
	"beq": func(cpu *CPU, rs1, rs2 int32, destination string) { trueOrNext(cpu, rs1 == rs2, destination) },
	"bne": func(cpu *CPU, rs1, rs2 int32, destination string) { trueOrNext(cpu, rs1 != rs2, destination) },
	"blt": func(cpu *CPU, rs1, rs2 int32, destination string) { trueOrNext(cpu, rs1 < rs2, destination) },
	"bltu": func(cpu *CPU, rs1, rs2 int32, destination string) {
		trueOrNext(cpu, uint32(rs1) < uint32(rs2), destination)
	},
	"bgt": func(cpu *CPU, rs1, rs2 int32, destination string) { trueOrNext(cpu, rs1 > rs2, destination) },
	"bgtu": func(cpu *CPU, rs1, rs2 int32, destination string) {
		trueOrNext(cpu, uint32(rs1) > uint32(rs2), destination)
	},
	"ble": func(cpu *CPU, rs1, rs2 int32, destination string) { trueOrNext(cpu, rs1 <= rs2, destination) },
	"bleu": func(cpu *CPU, rs1, rs2 int32, destination string) {
		trueOrNext(cpu, uint32(rs1) <= uint32(rs2), destination)
	},
	"bge": func(cpu *CPU, rs1, rs2 int32, destination string) { trueOrNext(cpu, rs1 >= rs2, destination) },
	"bgeu": func(cpu *CPU, rs1, rs2 int32, destination string) {
		trueOrNext(cpu, uint32(rs1) >= uint32(rs2), destination)
	},
}

func parseBranchThree(tokens []string) Instr {
	var ok bool

	op, ok := instrToBranchThreeOp[tokens[0]]

	if !ok {
		return &NoOp{reason: "Invalid Operation"}
	}

	instr := BranchThreeInstr{
		rs1:         getRegisterNumber(tokens[1]),
		rs2:         getRegisterNumber(tokens[2]),
		destination: tokens[3],
		op:          op,
	}

	return &instr
}

var instrToBranchTwoOp = map[string]func(*CPU, int32, string){
	"beqz": func(cpu *CPU, rs1 int32, destination string) { trueOrNext(cpu, rs1 == 0, destination) },
	"bnez": func(cpu *CPU, rs1 int32, destination string) { trueOrNext(cpu, rs1 != 0, destination) },
	"bltz": func(cpu *CPU, rs1 int32, destination string) { trueOrNext(cpu, rs1 < 0, destination) },
	"bgtz": func(cpu *CPU, rs1 int32, destination string) { trueOrNext(cpu, rs1 > 0, destination) },
	"bgez": func(cpu *CPU, rs1 int32, destination string) { trueOrNext(cpu, rs1 >= 0, destination) },
}

func parseBranchTwo(tokens []string) Instr {
	op, ok := instrToBranchTwoOp[tokens[0]]

	if !ok {
		return &NoOp{reason: "Invalid Operation"}
	}

	instr := BranchTwoInstr{
		rs1:         getRegisterNumber(tokens[1]),
		destination: tokens[2],
		op:          op,
	}

	return &instr
}

func parseJal(tokens []string) Instr {

	instr := JumpAndLinkInstr{
		rd:          getRegisterNumber(tokens[1]),
		destination: tokens[2],
	}

	return &instr
}

func parseJalr(tokens []string) Instr {

	instr := JumpAndLinkRInstr{
		rd:  getRegisterNumber(tokens[1]),
		rs1: getRegisterNumber(tokens[2]),
		imm: parseImm(tokens[3])}

	return &instr
}

var instrToSetOp = map[string]func(int32, int32) bool{
	"slt":  func(rs1, rs2 int32) bool { return rs1 < rs2 },
	"sltu": func(rs1, rs2 int32) bool { return uint32(rs1) < uint32(rs2) },
}

func parseSet(tokens []string) Instr {

	op, ok := instrToSetOp[tokens[0]]

	if !ok {
		return &NoOp{reason: "Invalid Operation"}
	}

	instr := SetInstr{
		rd:  getRegisterNumber(tokens[1]),
		rs1: getRegisterNumber(tokens[2]),
		rs2: getRegisterNumber(tokens[3]),
		op:  op,
	}

	return &instr
}

var instrToSetImmOp = map[string]func(int32, int32) bool{
	"slti":  func(rs1, imm int32) bool { return rs1 < imm },
	"sltiu": func(rs1, imm int32) bool { return uint32(rs1) < uint32(imm) },
}

func parseSetImm(tokens []string) Instr {

	op, ok := instrToSetImmOp[tokens[0]]

	if !ok {
		return &NoOp{reason: "Invalid Operation"}
	}

	instr := SetImmInstr{
		rd:  getRegisterNumber(tokens[1]),
		rs1: getRegisterNumber(tokens[2]),
		imm: parseImm(tokens[3]),
		op:  op,
	}

	return &instr
}

func DecodeInstr(instr_str_raw *string) Instr {
	// simple decoding by matching the instr token with the lists defined in instructions.go

	instr_str := strings.TrimSpace(*instr_str_raw)

	// uses regex also means we dont need to check the amount of tokens, since in order to match,
	// they NEED to have the right amount

	firstTokenRe := regexp.MustCompile(`^(\w+)`)
	threePtRe := regexp.MustCompile(`(\w+)\s+(\w+)\s*,\s*(\w+)\s*,\s*(\-?\.?\w+)`)
	twoPtImmRe := regexp.MustCompile(`(\w+)\s+(\w+)\s*,\s*(\w+)`)
	loadStoreRe := regexp.MustCompile(`(\w+)\s+(\w+)\s*,\s*(-?[0-9]+)\(([a-z0-9]+)\)`)
	jumpRe := regexp.MustCompile(`(\w)\s+(.?\w+)`)

	instrTypeToken := firstTokenRe.FindString(instr_str)

	if slices.Contains(threePtInstrTypes, instrTypeToken) {
		tokens := threePtRe.FindStringSubmatch(instr_str)
		if len(tokens) == 0 {
			return &NoOp{}
		}
		return parseThreePt(tokens[1:])
	}

	if slices.Contains(threePtImmInstrTypes, instrTypeToken) {
		tokens := threePtRe.FindStringSubmatch(instr_str)
		if len(tokens) == 0 {
			return &NoOp{}
		}
		return parseThreePtImm(tokens[1:])
	}

	if slices.Contains(loadImmInstrTypes, instrTypeToken) {
		tokens := twoPtImmRe.FindStringSubmatch(instr_str)
		if len(tokens) == 0 {
			return &NoOp{}
		}
		return parseLoadImm(tokens[1:])
	}

	if slices.Contains(loadInstrTypes, instrTypeToken) {
		tokens := loadStoreRe.FindStringSubmatch(instr_str)
		if len(tokens) == 0 {
			return &NoOp{}
		}
		return parseLoad(tokens[1:])
	}

	if slices.Contains(storeInstrTypes, instrTypeToken) {
		tokens := loadStoreRe.FindStringSubmatch(instr_str)
		if len(tokens) == 0 {
			return &NoOp{}
		}
		return parseStore(tokens[1:])
	}

	if slices.Contains(branchThreeInstrTypes, instrTypeToken) {
		tokens := threePtRe.FindStringSubmatch(instr_str)
		if len(tokens) == 0 {
			return &NoOp{}
		}
		return parseBranchThree(tokens[1:])
	}

	if slices.Contains(branchTwoInstrTypes, instrTypeToken) {
		tokens := twoPtImmRe.FindStringSubmatch(instr_str)
		if len(tokens) == 0 {
			return &NoOp{}
		}
		return parseBranchTwo(tokens[1:])
	}

	if slices.Contains(setInstrTypes, instrTypeToken) {
		tokens := threePtRe.FindStringSubmatch(instr_str)
		if len(tokens) == 0 {
			return &NoOp{}
		}

		return parseSet(tokens[1:])
	}

	if slices.Contains(setImmInstrTypes, instrTypeToken) {
		tokens := threePtRe.FindStringSubmatch(instr_str)
		if len(tokens) == 0 {
			return &NoOp{}
		}

		return parseSetImm(tokens[1:])
	}

	if instrTypeToken == "j" {
		tokens := jumpRe.FindStringSubmatch(instr_str)
		if len(tokens) == 0 {
			return &NoOp{}
		}

		return &JumpInstr{destination: tokens[2]}
	}

	if instrTypeToken == "call" {
		tokens := jumpRe.FindStringSubmatch(instr_str)
		if len(tokens) == 0 {
			return &NoOp{}
		}

		return &JumpAndLinkInstr{rd: int8(abiToRegister["ra"]), destination: tokens[2]}
	}

	if instrTypeToken == "jr" {
		tokens := jumpRe.FindStringSubmatch(instr_str)
		if len(tokens) == 0 {
			return &NoOp{}
		}

		return &JumpAndLinkRInstr{
			rd:  0,
			rs1: getRegisterNumber(tokens[2]),
			imm: 0,
		}
	}

	if instrTypeToken == "jal" {
		tokens := twoPtImmRe.FindStringSubmatch(instr_str)
		if len(tokens) == 0 {
			return &NoOp{}
		}

		return parseJal(tokens)
	}

	if instrTypeToken == "jalr" {
		tokens := threePtRe.FindStringSubmatch(instr_str)
		if len(tokens) == 0 {
			return &NoOp{}
		}

		parseJalr(tokens)
	}

	if instrTypeToken == "mv" {
		tokens := twoPtImmRe.FindStringSubmatch(instr_str)
		if len(tokens) == 0 {
			return &NoOp{}
		}

		instr := InstrThreePtImm{}
		instr.rd = getRegisterNumber(tokens[2])
		instr.rs1 = getRegisterNumber(tokens[3])
		instr.imm = 0

		instr.op = func(i1, i2 int32) int32 { return i1 }

		return &instr
	}

	return &NoOp{}
}
