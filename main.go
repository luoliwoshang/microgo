package main

import (
	"fmt"
	"os"

	"tinygo.org/x/go-llvm"
)

func init() {
	llvm.InitializeAllTargets()
	llvm.InitializeAllTargetMCs()
	llvm.InitializeAllTargetInfos()
	llvm.InitializeAllAsmParsers()
	llvm.InitializeAllAsmPrinters()
}

func main() {
	ctx := llvm.NewContext()
	defer ctx.Dispose()

	module := ctx.NewModule("main")
	defer module.Dispose()

	builder := ctx.NewBuilder()
	defer builder.Dispose()

	targetTriple := "xtensa"
	target, err := llvm.GetTargetFromTriple(targetTriple)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法获取目标 '%s': %s\n", targetTriple, err)
		os.Exit(1)
	}
	targetMachine := target.CreateTargetMachine(targetTriple, "esp32", "", llvm.CodeGenLevelDefault, llvm.RelocDefault, llvm.CodeModelDefault)
	module.SetTarget(targetTriple)
	module.SetDataLayout(targetMachine.CreateTargetData().String())

	// int printf(const char* format, ...);
	// return type
	i32Type := ctx.Int32Type()
	// first param (const char* -> i8*)
	i8PtrType := llvm.PointerType(ctx.Int8Type(), 0)

	// functype ：return i32，first i8*
	// last set `true` determines if the function is variadic.
	printfFuncType := llvm.FunctionType(i32Type, []llvm.Type{i8PtrType}, true)

	// `printf` declare。
	printfFunc := llvm.AddFunction(module, "printf", printfFuncType)

	// --- mai ---
	// entry with `main` function.
	// function type: void main()
	mainFuncType := llvm.FunctionType(ctx.VoidType(), []llvm.Type{}, false)
	mainFunc := llvm.AddFunction(module, "main", mainFuncType)

	entryBlock := ctx.AddBasicBlock(mainFunc, "entry")
	builder.SetInsertPointAtEnd(entryBlock)

	//  "%s\n"
	formatStr := builder.CreateGlobalStringPtr("%s\n", ".formatstr")

	//  "hello"
	helloStr := builder.CreateGlobalStringPtr("hello", ".str")

	// call "printf"
	builder.CreateCall(printfFuncType, printfFunc, []llvm.Value{formatStr, helloStr}, "")

	// Create a void return instruction. Each basic block must end with a "Terminator Instruction".
	// RetVoid, Br, Switch, etc. are all terminator instructions.
	builder.CreateRetVoid()

	if ok := llvm.VerifyModule(module, llvm.ReturnStatusAction); ok != nil {
		fmt.Fprintf(os.Stderr, "错误: LLVM模块验证失败: %s\n", ok)
		os.Exit(1)
	}

	os.WriteFile("main.ll", []byte(module.String()), 0644)
}
