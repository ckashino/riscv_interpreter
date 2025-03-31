# Demo
![demo](https://github.com/user-attachments/assets/04633dac-b94e-435e-86ce-3736dff82afc)

*Demo of stepping though a program that computes the fibonacci of 6. Source in riscv/test.asm*

# RISC-V Command Line Interpreter
A simple RISC-V interpreter that can handle the instructions within the base integer instruction set (RV32I) as well as many pseudo commands. 

Does not support .text, .data, .global, etc. As such, the entry point will always be the first instruction.
