package main

import (
	"fmt"
	"os"

	"tinygo.org/x/go-llvm"
)

func main() {
	// --- 1. 初始化LLVM ---
	// 初始化所有LLVM目标
	llvm.InitializeAllTargets()
	llvm.InitializeAllTargetMCs()
	llvm.InitializeAllTargetInfos()
	llvm.InitializeAllAsmParsers()
	llvm.InitializeAllAsmPrinters()
	
	// LLVM的很多操作都需要一个上下文(Context)对象，它管理着核心数据结构和内存。
	ctx := llvm.NewContext()
	// 使用defer确保在main函数退出时，所有与Context相关的资源都能被正确释放。
	defer ctx.Dispose()

	// 创建一个新的LLVM模块(Module)。模块是LLVM IR的顶级容器，
	// 相当于一个.c文件编译后的.o目标文件，它包含了函数、全局变量等。
	module := ctx.NewModule("main")
	// 同样，确保模块资源被释放。
	defer module.Dispose()

	// 创建一个指令构建器(Builder)。
	// Builder提供了一系列便捷的API，用于在基本块(Basic Block)中插入指令。
	builder := ctx.NewBuilder()
	defer builder.Dispose()

	// --- 2. 配置目标平台 ---
	// 这是至关重要的一步，它告诉LLVM我们的目标平台是什么。
	// 对于ESP32的Xtensa架构，目标三元组(Target Triple)通常是 "xtensa-esp32-none-elf"。
	// 如果不设置，LLVM会使用当前宿主机的默认配置。
	targetTriple := "xtensa"
	target, err := llvm.GetTargetFromTriple(targetTriple)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法获取目标 '%s': %s\n", targetTriple, err)
		os.Exit(1)
	}
	targetMachine := target.CreateTargetMachine(targetTriple, "esp32", "", llvm.CodeGenLevelDefault, llvm.RelocDefault, llvm.CodeModelDefault)
	module.SetTarget(targetTriple)
	// 修正: 使用 CreateTargetData() 而不是 CreateDataLayout()
	module.SetDataLayout(targetMachine.CreateTargetData().String())

	// --- 3. 声明外部C函数(printf) ---
	// 我们将直接调用C标准库中的`printf`函数。这避免了使用CGo和自定义的Go运行时。
	// 在LLVM IR中，我们必须先“声明”这个外部函数，告诉LLVM它的存在和函数签名，
	// 最终的链接器会负责找到`printf`的真正实现。
	//
	// C函数签名: int printf(const char* format, ...);
	// `...` 表示这是一个可变参数函数。

	// 定义 `printf` 的返回类型 (int -> i32)
	i32Type := ctx.Int32Type()
	// 定义 `printf` 的第一个参数类型 (const char* -> i8*)
	i8PtrType := llvm.PointerType(ctx.Int8Type(), 0)

	// 定义函数类型：返回i32，接受一个i8*参数，并且是可变参数(variadic)。
	// 最后一个布尔值参数设为`true`来表示函数是可变参数的。
	printfFuncType := llvm.FunctionType(i32Type, []llvm.Type{i8PtrType}, true)

	// 向模块中添加`printf`的函数声明。
	printfFunc := llvm.AddFunction(module, "printf", printfFuncType)

	// --- 4. 创建 main 函数 ---
	// 所有可执行程序都需要一个入口点。我们创建一个名为 "main" 的函数。
	// 函数类型：返回void，不接受任何参数。
	mainFuncType := llvm.FunctionType(ctx.VoidType(), []llvm.Type{}, false)
	mainFunc := llvm.AddFunction(module, "main", mainFuncType)

	// 为main函数创建一个基本块(Basic Block)，名为 "entry"。
	// 基本块是指令的容器，一个函数可以有多个基本块。
	entryBlock := ctx.AddBasicBlock(mainFunc, "entry")
	// 将构建器的插入点定位到 "entry" 块的末尾。
	builder.SetInsertPointAtEnd(entryBlock)

	// --- 5. 生成函数体指令 ---
	// 为printf创建格式化字符串 "%s\n"
	formatStr := builder.CreateGlobalStringPtr("%s\n", ".formatstr")

	// 创建要打印的字符串 "hello"
	helloStr := builder.CreateGlobalStringPtr("hello", ".str")

	// 生成对 "printf" 函数的调用指令。
	// 参数列表现在包含格式化字符串和要打印的字符串。
	builder.CreateCall(printfFuncType, printfFunc, []llvm.Value{formatStr, helloStr}, "")

	// 创建一个void返回指令。每个基本块都必须以一个“终结者指令”(Terminator Instruction)结尾。
	// RetVoid, Br, Switch等都是终结者指令。
	builder.CreateRetVoid()

	// --- 6. 验证并打印LLVM IR ---
	// 在生成IR后，最好运行验证器检查是否存在明显的错误。
	if ok := llvm.VerifyModule(module, llvm.ReturnStatusAction); ok != nil {
		fmt.Fprintf(os.Stderr, "错误: LLVM模块验证失败: %s\n", ok)
		os.Exit(1)
	}

	// 将最终生成的LLVM IR以人类可读的文本格式打印到标准错误流。
	// 您也可以使用 module.WriteBitcodeToFile("main.bc") 将其写入二进制的bitcode文件。
	fmt.Fprintln(os.Stderr, "--- 生成的LLVM IR ---")
	module.Dump()
}
