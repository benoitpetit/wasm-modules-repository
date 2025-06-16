//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"
)

var silentMode = false

// Fonction pour activer/désactiver le mode silencieux
func setSilentMode(this js.Value, args []js.Value) interface{} {
	if len(args) == 1 {
		silentMode = args[0].Bool()
	}
	return js.ValueOf(silentMode)
}

// Fonction d'addition
func add(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf("Erreur: deux arguments requis pour add")
	}
	
	a := args[0].Float()
	b := args[1].Float()
	result := a + b
	
	// Mode silencieux pour éviter trop de logs pendant les tests de performance
	if !silentMode {
		fmt.Printf("Go WASM: %f + %f = %f\n", a, b, result)
	}
	return js.ValueOf(result)
}

// Fonction de soustraction
func subtract(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf("Erreur: deux arguments requis pour subtract")
	}
	
	a := args[0].Float()
	b := args[1].Float()
	result := a - b
	
	if !silentMode {
		fmt.Printf("Go WASM: %f - %f = %f\n", a, b, result)
	}
	return js.ValueOf(result)
}

// Fonction de multiplication
func multiply(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf("Erreur: deux arguments requis pour multiply")
	}
	
	a := args[0].Float()
	b := args[1].Float()
	result := a * b
	
	if !silentMode {
		fmt.Printf("Go WASM: %f * %f = %f\n", a, b, result)
	}
	return js.ValueOf(result)
}

// Fonction de division
func divide(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf("Erreur: deux arguments requis pour divide")
	}
	
	a := args[0].Float()
	b := args[1].Float()
	
	if b == 0 {
		return js.ValueOf("Erreur: division par zéro")
	}
	
	result := a / b
	if !silentMode {
		fmt.Printf("Go WASM: %f / %f = %f\n", a, b, result)
	}
	return js.ValueOf(result)
}

// Fonction pour calculer la puissance
func power(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf("Erreur: deux arguments requis pour power")
	}
	
	base := args[0].Float()
	exp := args[1].Float()
	
	// Calcul simple de puissance (entiers positifs seulement)
	if exp < 0 {
		return js.ValueOf("Erreur: exposant négatif non supporté")
	}
	
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	
	if !silentMode {
		fmt.Printf("Go WASM: %f^%f = %f\n", base, exp, result)
	}
	return js.ValueOf(result)
}

// Fonction pour obtenir les fonctions disponibles
func getAvailableFunctions(this js.Value, args []js.Value) interface{} {
	functions := []interface{}{
		"add", "subtract", "multiply", "divide", "power", "factorial", "getAvailableFunctions", "setSilentMode",
	}
	return js.ValueOf(functions)
}

// Fonction factorielle
func factorial(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Erreur: un argument requis pour factorial")
	}
	
	n := int(args[0].Float())
	
	if n < 0 {
		return js.ValueOf("Erreur: factorielle non définie pour les nombres négatifs")
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

func main() {
	fmt.Println("Go WASM Math Module initializing...")
	
	// Enregistrer toutes les fonctions
	js.Global().Set("add", js.FuncOf(add))
	js.Global().Set("subtract", js.FuncOf(subtract))
	js.Global().Set("multiply", js.FuncOf(multiply))
	js.Global().Set("divide", js.FuncOf(divide))
	js.Global().Set("power", js.FuncOf(power))
	js.Global().Set("factorial", js.FuncOf(factorial))
	js.Global().Set("getAvailableFunctions", js.FuncOf(getAvailableFunctions))
	js.Global().Set("setSilentMode", js.FuncOf(setSilentMode))
	
	// Signal de prêt pour GoWM
	js.Global().Set("__gowm_ready", js.ValueOf(true))
	
	fmt.Println("Go WASM Math Module ready! Available functions: add, subtract, multiply, divide, power, factorial, setSilentMode")
	
	// Maintenir le programme en vie
	select {}
}
