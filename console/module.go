package console

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/anand-tan/goja_nodejs/require"
	"github.com/anand-tan/goja_nodejs/util"
)

const ModuleName = "console"

type Console struct {
	runtime *goja.Runtime
	util    *goja.Object
	printer Printer
}

type Printer interface {
	Log(string)
	Warn(string)
	Error(string)
}

func (c *Console) log(p func(string)) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if format, ok := goja.AssertFunction(c.util.Get("format")); ok {
			ret, err := format(c.util, call.Arguments...)
			if err != nil {
				panic(err)
			}

			p(ret.String())
		} else {
			panic(c.runtime.NewTypeError("util.format is not a function"))
		}

		return nil
	}
}

func Require(runtime *goja.Runtime, module *goja.Object) {
	requireWithPrinter(defaultStdPrinter)(runtime, module)
}

func RequireWithPrinter(printer Printer) require.ModuleLoader {
	return requireWithPrinter(printer)
}

func requireWithPrinter(printer Printer) require.ModuleLoader {
	return func(runtime *goja.Runtime, module *goja.Object) {
		c := &Console{
			runtime: runtime,
			printer: printer,
		}

		c.util = require.Require(runtime, util.ModuleName).(*goja.Object)

		o := module.Get("exports").(*goja.Object)
		o.Set("log", c.log(c.printer.Log))
		o.Set("error", c.log(c.printer.Error))
		o.Set("warn", c.log(c.printer.Warn))
		o.Set("info", c.log(c.printer.Log))
		o.Set("debug", c.log(c.printer.Log))
	}
}

func Enable(runtime *goja.Runtime) {
	runtime.Set("console", require.Require(runtime, ModuleName))
}

func init() {
	require.RegisterCoreModule(ModuleName, Require)
}

func SetGlobal(runtime *goja.Runtime, printer Printer) error {
	p := printer
	if p == nil {
		p = defaultStdPrinter
	}
	c := &Console{
		runtime: runtime,
		printer: p,
	}

	if u := runtime.Get("util"); u != nil {
		c.util = u.ToObject(runtime)
		o := runtime.NewObject()
		o.Set("log", c.log(c.printer.Log))
		o.Set("error", c.log(c.printer.Error))
		o.Set("warn", c.log(c.printer.Warn))
		o.Set("info", c.log(c.printer.Log))
		o.Set("debug", c.log(c.printer.Log))
		runtime.Set(ModuleName, o)
	} else {
		return fmt.Errorf("util module is not available in the runtime")
	}
	return nil
}
