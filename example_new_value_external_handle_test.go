package v8go_test

import (
	"fmt"
	"runtime/cgo"

	v8 "github.com/tommie/v8go"
)

type Calculator struct {
	stack []int32
}

func (c *Calculator) Peek() int32 {
	l := len(c.stack)
	if l > 0 {
		return c.stack[l-1]
	} else {
		return 0
	}
}
func (c *Calculator) Push(v int32) { c.stack = append(c.stack, v) }
func (c *Calculator) Add() {
	l := len(c.stack)
	if l < 2 {
		panic("Not enough elements")
	}
	a := c.stack[l-1]
	b := c.stack[l-2]
	c.stack[l-2] = a + b
	c.stack = c.stack[0 : l-1]
}

// getInstance retrieves the go Calculator instance that is stored in an
// _internal field_.
func getInstance(info *v8.FunctionCallbackInfo) *Calculator {
	var internalField *v8.Value = info.This().GetInternalField(0)
	var handle cgo.Handle = internalField.ExternalHandle()
	calculator, ok := handle.Value().(*Calculator)
	if !ok {
		panic("Not a calculator")
	}
	return calculator
}

func CreateJSCalculator(iso *v8.Isolate) *v8.FunctionTemplate {
	constructor := v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		calculator := &Calculator{}
		handle := cgo.NewHandle(calculator)
		info.This().SetInternalField(0,
			v8.NewValueExternalHandle(iso, handle))
		return nil
	})

	// Set the internal field count on the **instance template**.
	instanceTemplate := constructor.InstanceTemplate()
	instanceTemplate.SetInternalFieldCount(1)

	// Create the methods on the **prototype template**.
	prototypeTemplate := constructor.PrototypeTemplate()
	prototypeTemplate.Set("push",
		v8.NewFunctionTemplate(iso,
			func(info *v8.FunctionCallbackInfo) *v8.Value {
				calculator := getInstance(info)
				calculator.Push(info.Args()[0].Int32())
				return nil
			}))
	prototypeTemplate.Set("add",
		v8.NewFunctionTemplate(iso,
			func(info *v8.FunctionCallbackInfo) *v8.Value {
				calculator := getInstance(info)
				calculator.Add()
				return nil
			}))
	prototypeTemplate.Set("peek",
		v8.NewFunctionTemplateWithError(iso,
			func(info *v8.FunctionCallbackInfo) (*v8.Value, error) {
				calculator := getInstance(info)
				return v8.NewValue(iso, calculator.Peek())
			}))
	return constructor
}

func ExampleNewValueExternalHandle() {
	iso := v8.NewIsolate()
	defer iso.Dispose()

	global := v8.NewObjectTemplate(iso)
	global.Set("Calculator", CreateJSCalculator(iso))
	ctx := v8.NewContext(iso, global)

	defer ctx.Close()
	v, err := ctx.RunScript(`
		const calculator = new Calculator()
		calculator.push(7)
		calculator.push(35)
		calculator.add()
		calculator.peek()`, "")
	if err != nil {
		fmt.Println("Error running script", err.Error())
		return
	}
	fmt.Printf("The result is: %d\n", v.Integer())
	// Output:
	// The result is: 42
}
