//go:build js && wasm

package main

import (
	"fmt"
	"math"
	"sort"
	"syscall/js"
)

var silentMode = false

// setSilentMode enables/disables silent mode for console logs
func setSilentMode(this js.Value, args []js.Value) interface{} {
	if len(args) == 1 {
		silentMode = args[0].Bool()
	}
	return js.ValueOf(silentMode)
}

// Basic arithmetic operations
func add(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf("Error: two arguments required for add")
	}

	a := args[0].Float()
	b := args[1].Float()
	result := a + b

	if !silentMode {
		fmt.Printf("Go WASM: %f + %f = %f\n", a, b, result)
	}
	return js.ValueOf(result)
}

func subtract(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf("Error: two arguments required for subtract")
	}

	a := args[0].Float()
	b := args[1].Float()
	result := a - b

	if !silentMode {
		fmt.Printf("Go WASM: %f - %f = %f\n", a, b, result)
	}
	return js.ValueOf(result)
}

func multiply(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf("Error: two arguments required for multiply")
	}

	a := args[0].Float()
	b := args[1].Float()
	result := a * b

	if !silentMode {
		fmt.Printf("Go WASM: %f * %f = %f\n", a, b, result)
	}
	return js.ValueOf(result)
}

func divide(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf("Error: two arguments required for divide")
	}

	a := args[0].Float()
	b := args[1].Float()

	if b == 0 {
		return js.ValueOf("Error: division by zero")
	}

	result := a / b
	if !silentMode {
		fmt.Printf("Go WASM: %f / %f = %f\n", a, b, result)
	}
	return js.ValueOf(result)
}

func power(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf("Error: two arguments required for power")
	}

	base := args[0].Float()
	exp := args[1].Float()

	result := math.Pow(base, exp)

	if math.IsNaN(result) || math.IsInf(result, 0) {
		return js.ValueOf("Error: invalid result (NaN or Infinity)")
	}

	if !silentMode {
		fmt.Printf("Go WASM: %f^%f = %f\n", base, exp, result)
	}
	return js.ValueOf(result)
}

func factorial(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for factorial")
	}

	n := int(args[0].Float())

	if n < 0 {
		return js.ValueOf("Error: factorial not defined for negative numbers")
	}

	if n > 170 {
		return js.ValueOf("Error: factorial overflow for numbers greater than 170")
	}

	if n == 0 || n == 1 {
		return js.ValueOf(1)
	}

	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}

	if !silentMode {
		fmt.Printf("Go WASM: %d! = %d\n", n, result)
	}
	return js.ValueOf(result)
}

// Advanced mathematical functions
func sqrt(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for sqrt")
	}

	x := args[0].Float()

	if x < 0 {
		return js.ValueOf("Error: square root of negative number")
	}

	result := math.Sqrt(x)

	if !silentMode {
		fmt.Printf("Go WASM: sqrt(%f) = %f\n", x, result)
	}
	return js.ValueOf(result)
}

func abs(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for abs")
	}

	x := args[0].Float()
	result := math.Abs(x)

	if !silentMode {
		fmt.Printf("Go WASM: abs(%f) = %f\n", x, result)
	}
	return js.ValueOf(result)
}

func min(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf("Error: at least two arguments required for min")
	}

	result := args[0].Float()
	for i := 1; i < len(args); i++ {
		val := args[i].Float()
		if val < result {
			result = val
		}
	}

	if !silentMode {
		fmt.Printf("Go WASM: min of %d numbers = %f\n", len(args), result)
	}
	return js.ValueOf(result)
}

func max(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf("Error: at least two arguments required for max")
	}

	result := args[0].Float()
	for i := 1; i < len(args); i++ {
		val := args[i].Float()
		if val > result {
			result = val
		}
	}

	if !silentMode {
		fmt.Printf("Go WASM: max of %d numbers = %f\n", len(args), result)
	}
	return js.ValueOf(result)
}

// Trigonometric functions
func sin(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for sin")
	}

	x := args[0].Float()
	result := math.Sin(x)

	if !silentMode {
		fmt.Printf("Go WASM: sin(%f) = %f\n", x, result)
	}
	return js.ValueOf(result)
}

func cos(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for cos")
	}

	x := args[0].Float()
	result := math.Cos(x)

	if !silentMode {
		fmt.Printf("Go WASM: cos(%f) = %f\n", x, result)
	}
	return js.ValueOf(result)
}

func tan(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for tan")
	}

	x := args[0].Float()
	result := math.Tan(x)

	if math.IsNaN(result) || math.IsInf(result, 0) {
		return js.ValueOf("Error: tangent is undefined for this value")
	}

	if !silentMode {
		fmt.Printf("Go WASM: tan(%f) = %f\n", x, result)
	}
	return js.ValueOf(result)
}

// Logarithmic functions
func log(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for log")
	}

	x := args[0].Float()

	if x <= 0 {
		return js.ValueOf("Error: logarithm of non-positive number")
	}

	result := math.Log(x)

	if !silentMode {
		fmt.Printf("Go WASM: ln(%f) = %f\n", x, result)
	}
	return js.ValueOf(result)
}

func log10(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for log10")
	}

	x := args[0].Float()

	if x <= 0 {
		return js.ValueOf("Error: logarithm of non-positive number")
	}

	result := math.Log10(x)

	if !silentMode {
		fmt.Printf("Go WASM: log10(%f) = %f\n", x, result)
	}
	return js.ValueOf(result)
}

// Number theory functions
func gcd(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf("Error: two arguments required for gcd")
	}

	a := int(math.Abs(args[0].Float()))
	b := int(math.Abs(args[1].Float()))

	for b != 0 {
		a, b = b, a%b
	}

	if !silentMode {
		fmt.Printf("Go WASM: gcd(%d, %d) = %d\n", int(args[0].Float()), int(args[1].Float()), a)
	}
	return js.ValueOf(a)
}

func lcm(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf("Error: two arguments required for lcm")
	}

	a := int(math.Abs(args[0].Float()))
	b := int(math.Abs(args[1].Float()))

	if a == 0 || b == 0 {
		return js.ValueOf(0)
	}

	// Calculate GCD first
	gcdVal := a
	bVal := b
	for bVal != 0 {
		gcdVal, bVal = bVal, gcdVal%bVal
	}

	result := (a * b) / gcdVal

	if !silentMode {
		fmt.Printf("Go WASM: lcm(%d, %d) = %d\n", int(args[0].Float()), int(args[1].Float()), result)
	}
	return js.ValueOf(result)
}

func isPrime(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for isPrime")
	}

	n := int(args[0].Float())

	if n < 2 {
		return js.ValueOf(false)
	}

	if n == 2 {
		return js.ValueOf(true)
	}

	if n%2 == 0 {
		return js.ValueOf(false)
	}

	for i := 3; i*i <= n; i += 2 {
		if n%i == 0 {
			return js.ValueOf(false)
		}
	}

	if !silentMode {
		fmt.Printf("Go WASM: isPrime(%d) = true\n", n)
	}
	return js.ValueOf(true)
}

func fibonacci(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for fibonacci")
	}

	n := int(args[0].Float())

	if n < 0 {
		return js.ValueOf("Error: fibonacci not defined for negative numbers")
	}

	if n > 92 {
		return js.ValueOf("Error: fibonacci overflow for numbers greater than 92")
	}

	if n <= 1 {
		return js.ValueOf(n)
	}

	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}

	if !silentMode {
		fmt.Printf("Go WASM: fibonacci(%d) = %d\n", n, b)
	}
	return js.ValueOf(b)
}

// Statistical functions
func mean(this js.Value, args []js.Value) interface{} {
	if len(args) == 0 {
		return js.ValueOf("Error: at least one argument required for mean")
	}

	sum := 0.0
	for i := 0; i < len(args); i++ {
		sum += args[i].Float()
	}

	result := sum / float64(len(args))

	if !silentMode {
		fmt.Printf("Go WASM: mean of %d numbers = %f\n", len(args), result)
	}
	return js.ValueOf(result)
}

func median(this js.Value, args []js.Value) interface{} {
	if len(args) == 0 {
		return js.ValueOf("Error: at least one argument required for median")
	}

	numbers := make([]float64, len(args))
	for i := 0; i < len(args); i++ {
		numbers[i] = args[i].Float()
	}

	sort.Float64s(numbers)

	n := len(numbers)
	var result float64

	if n%2 == 0 {
		result = (numbers[n/2-1] + numbers[n/2]) / 2
	} else {
		result = numbers[n/2]
	}

	if !silentMode {
		fmt.Printf("Go WASM: median of %d numbers = %f\n", len(args), result)
	}
	return js.ValueOf(result)
}

func standardDeviation(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf("Error: at least two arguments required for standard deviation")
	}

	// Calculate mean
	sum := 0.0
	for i := 0; i < len(args); i++ {
		sum += args[i].Float()
	}
	meanVal := sum / float64(len(args))

	// Calculate variance
	variance := 0.0
	for i := 0; i < len(args); i++ {
		diff := args[i].Float() - meanVal
		variance += diff * diff
	}
	variance /= float64(len(args))

	result := math.Sqrt(variance)

	if !silentMode {
		fmt.Printf("Go WASM: stddev of %d numbers = %f\n", len(args), result)
	}
	return js.ValueOf(result)
}

// Utility functions
func round(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 || len(args) > 2 {
		return js.ValueOf("Error: one or two arguments required for round")
	}

	x := args[0].Float()
	precision := 0

	if len(args) == 2 {
		precision = int(args[1].Float())
	}

	multiplier := math.Pow(10, float64(precision))
	result := math.Round(x*multiplier) / multiplier

	if !silentMode {
		fmt.Printf("Go WASM: round(%f, %d) = %f\n", x, precision, result)
	}
	return js.ValueOf(result)
}

func ceil(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for ceil")
	}

	x := args[0].Float()
	result := math.Ceil(x)

	if !silentMode {
		fmt.Printf("Go WASM: ceil(%f) = %f\n", x, result)
	}
	return js.ValueOf(result)
}

func floor(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for floor")
	}

	x := args[0].Float()
	result := math.Floor(x)

	if !silentMode {
		fmt.Printf("Go WASM: floor(%f) = %f\n", x, result)
	}
	return js.ValueOf(result)
}

func getAvailableFunctions(this js.Value, args []js.Value) interface{} {
	functions := []interface{}{
		// Basic arithmetic
		"add", "subtract", "multiply", "divide", "power", "factorial",
		// Advanced math
		"sqrt", "abs", "min", "max",
		// Trigonometric
		"sin", "cos", "tan",
		// Logarithmic
		"log", "log10",
		// Number theory
		"gcd", "lcm", "isPrime", "fibonacci",
		// Statistical
		"mean", "median", "standardDeviation",
		// Utility
		"round", "ceil", "floor",
		// System
		"getAvailableFunctions", "setSilentMode",
	}
	return js.ValueOf(functions)
}

func main() {
	fmt.Println("Go WASM Enhanced Math Module initializing...")

	// Register basic arithmetic functions
	js.Global().Set("add", js.FuncOf(add))
	js.Global().Set("subtract", js.FuncOf(subtract))
	js.Global().Set("multiply", js.FuncOf(multiply))
	js.Global().Set("divide", js.FuncOf(divide))
	js.Global().Set("power", js.FuncOf(power))
	js.Global().Set("factorial", js.FuncOf(factorial))

	// Register advanced mathematical functions
	js.Global().Set("sqrt", js.FuncOf(sqrt))
	js.Global().Set("abs", js.FuncOf(abs))
	js.Global().Set("min", js.FuncOf(min))
	js.Global().Set("max", js.FuncOf(max))

	// Register trigonometric functions
	js.Global().Set("sin", js.FuncOf(sin))
	js.Global().Set("cos", js.FuncOf(cos))
	js.Global().Set("tan", js.FuncOf(tan))

	// Register logarithmic functions
	js.Global().Set("log", js.FuncOf(log))
	js.Global().Set("log10", js.FuncOf(log10))

	// Register number theory functions
	js.Global().Set("gcd", js.FuncOf(gcd))
	js.Global().Set("lcm", js.FuncOf(lcm))
	js.Global().Set("isPrime", js.FuncOf(isPrime))
	js.Global().Set("fibonacci", js.FuncOf(fibonacci))

	// Register statistical functions
	js.Global().Set("mean", js.FuncOf(mean))
	js.Global().Set("median", js.FuncOf(median))
	js.Global().Set("standardDeviation", js.FuncOf(standardDeviation))

	// Register utility functions
	js.Global().Set("round", js.FuncOf(round))
	js.Global().Set("ceil", js.FuncOf(ceil))
	js.Global().Set("floor", js.FuncOf(floor))

	// Register system functions
	js.Global().Set("getAvailableFunctions", js.FuncOf(getAvailableFunctions))
	js.Global().Set("setSilentMode", js.FuncOf(setSilentMode))

	// Signal readiness for GoWM
	js.Global().Set("__gowm_ready", js.ValueOf(true))

	fmt.Println("Go WASM Enhanced Math Module ready!")
	fmt.Println("Available functions: Basic arithmetic, Advanced math, Trigonometry, Logarithms, Number theory, Statistics, Utilities")

	// Keep the program alive
	select {}
}
