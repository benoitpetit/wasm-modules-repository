# Enhanced Math WASM Module

A comprehensive high-performance mathematical calculation module written in Go and compiled to WebAssembly. This module provides a wide range of mathematical functions including basic arithmetic, advanced mathematics, trigonometry, logarithms, number theory, statistics, and utility functions.

## 🚀 Features

- **Basic Arithmetic**: Addition, subtraction, multiplication, division, power, factorial
- **Advanced Mathematics**: Square root, absolute value, min/max operations
- **Trigonometry**: Sine, cosine, tangent functions (radians)
- **Logarithms**: Natural logarithm and base-10 logarithm
- **Number Theory**: GCD, LCM, prime checking, Fibonacci sequence
- **Statistics**: Mean, median, standard deviation
- **Utilities**: Rounding, ceiling, floor functions
- **Performance**: Compiled to WebAssembly for optimal speed
- **Error Handling**: Comprehensive input validation and error messages
- **GoWM Integration**: Optimized for the GoWM library

## 📦 Installation

### Using GoWM (Recommended)

```javascript
import { loadFromGitHub } from 'gowm';

const math = await loadFromGitHub('benoitpetit/wasm-modules-repository', {
  path: 'math-wasm',
  filename: 'main.wasm',
  name: 'math-wasm',
  branch: 'master'
});

// Configure for production
math.call('setSilentMode', true);
```

### Manual Installation

1. Download `main.wasm` and `wasm_exec.js`
2. Include both files in your web project
3. Load the WASM module using standard WebAssembly APIs

## 🔧 API Reference

### Basic Arithmetic

#### `add(a, b)` → `number`
Add two numbers.
```javascript
math.call('add', 5, 3); // Returns: 8
```

#### `subtract(a, b)` → `number`
Subtract two numbers.
```javascript
math.call('subtract', 10, 3); // Returns: 7
```

#### `multiply(a, b)` → `number`
Multiply two numbers.
```javascript
math.call('multiply', 4, 7); // Returns: 28
```

#### `divide(a, b)` → `number`
Divide two numbers with zero-division protection.
```javascript
math.call('divide', 15, 3); // Returns: 5
math.call('divide', 10, 0); // Returns: "Error: division by zero"
```

#### `power(base, exponent)` → `number`
Raise a number to a power (supports negative and fractional exponents).
```javascript
math.call('power', 2, 3);     // Returns: 8
math.call('power', 16, 0.5);  // Returns: 4
math.call('power', 2, -2);    // Returns: 0.25
```

#### `factorial(n)` → `number`
Calculate factorial of non-negative integer (0 ≤ n ≤ 170).
```javascript
math.call('factorial', 5);   // Returns: 120
math.call('factorial', 0);   // Returns: 1
math.call('factorial', -1);  // Returns: "Error: factorial not defined for negative numbers"
```

### Advanced Mathematics

#### `sqrt(x)` → `number`
Calculate square root of a non-negative number.
```javascript
math.call('sqrt', 16);  // Returns: 4
math.call('sqrt', -4);  // Returns: "Error: square root of negative number"
```

#### `abs(x)` → `number`
Calculate absolute value.
```javascript
math.call('abs', -5);   // Returns: 5
math.call('abs', 3.14); // Returns: 3.14
```

#### `min(...numbers)` → `number`
Find minimum value among multiple numbers.
```javascript
math.call('min', 5, 2, 8, 1);  // Returns: 1
```

#### `max(...numbers)` → `number`
Find maximum value among multiple numbers.
```javascript
math.call('max', 5, 2, 8, 1);  // Returns: 8
```

### Trigonometry

All trigonometric functions work with radians.

#### `sin(x)` → `number`
Calculate sine of angle in radians.
```javascript
math.call('sin', Math.PI / 2);  // Returns: 1
math.call('sin', 0);            // Returns: 0
```

#### `cos(x)` → `number`
Calculate cosine of angle in radians.
```javascript
math.call('cos', 0);            // Returns: 1
math.call('cos', Math.PI);      // Returns: -1
```

#### `tan(x)` → `number`
Calculate tangent of angle in radians.
```javascript
math.call('tan', Math.PI / 4);  // Returns: 1
math.call('tan', Math.PI / 2);  // Returns: "Error: tangent is undefined for this value"
```

### Logarithms

#### `log(x)` → `number`
Calculate natural logarithm (ln).
```javascript
math.call('log', Math.E);  // Returns: 1
math.call('log', 1);       // Returns: 0
math.call('log', 0);       // Returns: "Error: logarithm of non-positive number"
```

#### `log10(x)` → `number`
Calculate base-10 logarithm.
```javascript
math.call('log10', 100);   // Returns: 2
math.call('log10', 1000);  // Returns: 3
```

### Number Theory

#### `gcd(a, b)` → `number`
Calculate Greatest Common Divisor.
```javascript
math.call('gcd', 48, 18);  // Returns: 6
math.call('gcd', 17, 13);  // Returns: 1
```

#### `lcm(a, b)` → `number`
Calculate Least Common Multiple.
```javascript
math.call('lcm', 4, 6);    // Returns: 12
math.call('lcm', 8, 12);   // Returns: 24
```

#### `isPrime(n)` → `boolean`
Check if number is prime.
```javascript
math.call('isPrime', 17);  // Returns: true
math.call('isPrime', 15);  // Returns: false
math.call('isPrime', 2);   // Returns: true
```

#### `fibonacci(n)` → `number`
Calculate nth Fibonacci number (0 ≤ n ≤ 92).
```javascript
math.call('fibonacci', 10);  // Returns: 55
math.call('fibonacci', 0);   // Returns: 0
math.call('fibonacci', 1);   // Returns: 1
```

### Statistics

#### `mean(...numbers)` → `number`
Calculate arithmetic mean (average).
```javascript
math.call('mean', 1, 2, 3, 4, 5);  // Returns: 3
```

#### `median(...numbers)` → `number`
Calculate median value.
```javascript
math.call('median', 1, 2, 3, 4, 5);     // Returns: 3
math.call('median', 1, 2, 3, 4, 5, 6);  // Returns: 3.5
```

#### `standardDeviation(...numbers)` → `number`
Calculate population standard deviation.
```javascript
math.call('standardDeviation', 2, 4, 4, 4, 5, 5, 7, 9);  // Returns: 2
```

### Utilities

#### `round(x, precision?)` → `number`
Round number to specified decimal places.
```javascript
math.call('round', 3.14159);     // Returns: 3
math.call('round', 3.14159, 2);  // Returns: 3.14
```

#### `ceil(x)` → `number`
Round up to nearest integer.
```javascript
math.call('ceil', 3.14);  // Returns: 4
math.call('ceil', -2.1);  // Returns: -2
```

#### `floor(x)` → `number`
Round down to nearest integer.
```javascript
math.call('floor', 3.14);  // Returns: 3
math.call('floor', -2.1);  // Returns: -3
```

### System Functions

#### `setSilentMode(enabled)` → `boolean`
Enable/disable console logging.
```javascript
math.call('setSilentMode', true);   // Returns: true (enables silent mode)
math.call('setSilentMode', false);  // Returns: false (disables silent mode)
```

#### `getAvailableFunctions()` → `Array<string>`
Get list of all available functions.
```javascript
const functions = math.call('getAvailableFunctions');
console.log(functions); // ['add', 'subtract', 'multiply', ...]
```

## 🛡️ Error Handling

The module uses string-based error handling. When an operation fails, it returns an error string starting with "Error:".

```javascript
function safeMath(operation, ...args) {
  const result = math.call(operation, ...args);
  
  if (typeof result === 'string' && result.startsWith('Error:')) {
    throw new Error(`Math operation '${operation}' failed: ${result}`);
  }
  
  return result;
}

// Usage
try {
  const result = safeMath('divide', 10, 2);  // Returns: 5
  console.log(result);
} catch (error) {
  console.error(error.message);
}
```

## 📊 Performance

- **Basic Arithmetic**: < 0.1ms per operation
- **Trigonometric**: < 0.5ms per operation  
- **Statistical**: < 2ms for 1000 numbers
- **Number Theory**: < 1ms for typical inputs

## 🔒 Security Features

- Input validation for all numerical parameters
- Safe arithmetic operations with overflow protection
- Memory-safe Go implementation
- No external dependencies
- Protection against division by zero
- Bounds checking for factorial and fibonacci
- Range validation for mathematical functions

## 🌐 Browser Compatibility

- Chrome 57+
- Firefox 52+
- Safari 11+
- Edge 16+
- Node.js 14.0.0+

## 📝 Examples

### Scientific Calculator

```javascript
// Trigonometric calculations
const angle = Math.PI / 4;  // 45 degrees
console.log('sin(45°):', math.call('sin', angle));  // ~0.707
console.log('cos(45°):', math.call('cos', angle));  // ~0.707
console.log('tan(45°):', math.call('tan', angle));  // ~1

// Logarithmic calculations
console.log('ln(e):', math.call('log', Math.E));      // 1
console.log('log₁₀(1000):', math.call('log10', 1000)); // 3

// Number theory
console.log('GCD(48, 18):', math.call('gcd', 48, 18));     // 6
console.log('17 is prime:', math.call('isPrime', 17));     // true
console.log('F(10):', math.call('fibonacci', 10));         // 55
```

### Statistical Analysis

```javascript
const data = [2, 4, 6, 8, 10, 12, 14, 16, 18, 20];

console.log('Mean:', math.call('mean', ...data));                    // 11
console.log('Median:', math.call('median', ...data));                // 11
console.log('Std Dev:', math.call('standardDeviation', ...data));    // ~5.74
console.log('Min:', math.call('min', ...data));                      // 2
console.log('Max:', math.call('max', ...data));                      // 20
```

### React Calculator Component

```jsx
import React, { useState, useEffect } from 'react';
import { useWasmFromGitHub } from 'gowm/hooks/useWasm';

function Calculator() {
  const { wasm: math, loading, error } = useWasmFromGitHub(
    'benoitpetit/wasm-modules-repository',
    {
      path: 'math-wasm',
      filename: 'main.wasm',
      name: 'math-wasm'
    }
  );
  
  const [result, setResult] = useState('0');

  useEffect(() => {
    if (math) {
      math.call('setSilentMode', true);
    }
  }, [math]);

  const calculate = (operation, ...args) => {
    if (!math) return;
    
    const res = math.call(operation, ...args);
    if (typeof res === 'string' && res.startsWith('Error:')) {
      setResult(res);
    } else {
      setResult(res.toString());
    }
  };

  if (loading) return <div>Loading calculator...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div>
      <div>Result: {result}</div>
      <button onClick={() => calculate('add', 5, 3)}>5 + 3</button>
      <button onClick={() => calculate('sqrt', 16)}>√16</button>
      <button onClick={() => calculate('sin', Math.PI/2)}>sin(π/2)</button>
    </div>
  );
}
```

## 🔧 Building from Source

```bash
# Clone the repository
git clone <repository-url>
cd math-wasm

# Build the module
./build.sh

# The compiled WASM file will be available as main.wasm
```

## 📄 License

MIT License - see LICENSE file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## 📚 Version History

### v0.2.0 (2025-01-20)
- Added comprehensive trigonometric functions (sin, cos, tan)
- Implemented logarithmic functions (log, log10)
- Added number theory functions (gcd, lcm, isPrime, fibonacci)
- Implemented statistical functions (mean, median, standardDeviation)
- Added utility functions (round, ceil, floor, min, max)
- Enhanced error handling and input validation
- Improved power function using native math library
- Added overflow protection for factorial and fibonacci
- Updated documentation with comprehensive examples
- Optimized performance for all mathematical operations

### v0.1.9
- Initial release with basic arithmetic operations
- Added error handling for edge cases
- Optimized for GoWM integration 