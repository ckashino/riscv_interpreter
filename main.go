package main

import (
	"fmt"
	"riscv_interpreter/riscv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var registerToABI = map[int]string{
	0:  "zero",
	1:  "ra",
	2:  "sp",
	3:  "gp",
	4:  "tp",
	5:  "t0",
	6:  "t1",
	7:  "t2",
	8:  "fp",
	9:  "s1",
	10: "a0",
	11: "a1",
	12: "a2",
	13: "a3",
	14: "a4",
	15: "a5",
	16: "a6",
	17: "a7",
	18: "s2",
	19: "s3",
	20: "s4",
	21: "s5",
	22: "s6",
	23: "s7",
	24: "s8",
	25: "s9",
	26: "s10",
	27: "s11",
	28: "t3",
	29: "t4",
	30: "t5",
	31: "t6",
}

func updateRegisterText(cpu *riscv.CPU, registerText *tview.TextView) {
	var builder strings.Builder
	for i, reg := range cpu.Registers {
		builder.WriteString(fmt.Sprintf("x%d (%s): %d\n", i, registerToABI[i], reg))
	}

	builder.WriteString(fmt.Sprintf("\nPC: %d", cpu.PC))

	registerText.SetText(builder.String())
}

func updateMemHist(cpu *riscv.CPU, memoryText *tview.TextView) {
	var builder strings.Builder
	for _, operation := range cpu.MemoryHistory {
		builder.WriteString(operation)
		builder.WriteString("\n")
	}

	memoryText.SetText(builder.String())
}

func exectute(cpu *riscv.CPU, instrs []string) {
	cpu.LoadInstructions(instrs)
	cpu.RunProgram()
}

func step(cpu *riscv.CPU, instrs []string) {
	cpu.LoadInstructions(instrs)
	if !(cpu.Done) {
		cpu.RunNextInstruction()
	}
}

func main() {
	cpu := riscv.NewCPU(1024 * 10)

	instructions := tview.NewTextArea()
	instructions.SetPlaceholder("Enter Instructions Here...")

	instructions.SetTitle("Instructions").
		SetBorder(true)

	registerInfo := tview.NewTextView()

	registerInfo.SetBorder(true).
		SetTitle("Registers")

	memoryInfo := tview.NewTextView()

	memoryInfo.SetBorder(true).
		SetTitle("Memory Summary")

	currInstr := tview.NewTextView()
	currInstr.SetBorder(true)

	grid := tview.NewGrid().
		SetRows(3, 0, 3).
		SetColumns(-1, -1, -1)

	title := tview.NewTextView().
		SetTextAlign(tview.AlignCenter)

	title.SetText("Risc-V Interpreter").SetBorder(true)

	controls := tview.NewTextView()
	controls.SetText("(N)ext step: C-n	(R)un/(R)estart: C-r").SetBorder(true)
	controls.SetTextAlign(tview.AlignCenter)

	grid.AddItem(title, 0, 0, 1, 3, 0, 0, false).
		AddItem(instructions, 1, 0, 1, 1, 0, 0, true).
		AddItem(registerInfo, 1, 1, 1, 1, 0, 0, false).
		AddItem(memoryInfo, 1, 2, 1, 1, 0, 0, false).
		AddItem(currInstr, 2, 0, 1, 1, 0, 0, false).
		AddItem(controls, 2, 1, 1, 2, 0, 0, false)

	app := tview.NewApplication()

	updateRegisterText(&cpu, registerInfo)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlR {
			tokens := strings.Split(instructions.GetText(), "\n")
			exectute(&cpu, tokens)
			updateRegisterText(&cpu, registerInfo)
			updateMemHist(&cpu, memoryInfo)
		}

		if event.Key() == tcell.KeyCtrlN {
			tokens := strings.Split(strings.TrimSpace(instructions.GetText()), "\n")
			step(&cpu, tokens)
			updateRegisterText(&cpu, registerInfo)
			updateMemHist(&cpu, memoryInfo)
		}

		currInstr.SetText(cpu.GetCurrInstr())

		return event
	})

	if err := app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		panic(err)
	}
}
