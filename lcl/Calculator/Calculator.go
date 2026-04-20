package main

import (
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"strconv"
)

type TCalculatorForm struct {
	lcl.TEngForm
	Display   lcl.IEdit
	Result    float64
	Operator  string
	NewNumber bool
}

var CalcForm TCalculatorForm

func main() {
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForms(&CalcForm)
	lcl.Application.Run()
}

func (c *TCalculatorForm) FormCreate(sender lcl.IObject) {
	c.SetCaption("简易计算器")
	c.SetPosition(types.PoScreenCenter)
	c.SetWidth(320)
	c.SetHeight(450)
	c.SetColor(colors.ClWhite)

	c.Display = lcl.NewEdit(c)
	c.Display.SetParent(c)
	c.Display.SetLeft(10)
	c.Display.SetTop(10)
	c.Display.SetWidth(290)
	c.Display.SetHeight(40)
	c.Display.SetReadOnly(true)
	c.Display.SetText("0")
	c.Display.SetAlignment(types.TaRightJustify)

	buttons := []string{
		"7", "8", "9", "/",
		"4", "5", "6", "*",
		"1", "2", "3", "-",
		"0", ".", "=", "+",
	}

	for i, caption := range buttons {
		row := i / 4
		col := i % 4
		btn := lcl.NewButton(c)
		btn.SetParent(c)
		btn.SetCaption(caption)
		btn.SetLeft(int32(10 + col*75))
		btn.SetTop(int32(60 + row*75))
		btn.SetWidth(65)
		btn.SetHeight(65)

		if caption == "=" {
			btn.SetColor(colors.ClLightblue)
		}

		btn.SetOnClick(func(sender lcl.IObject) {
			c.OnButtonClick(lcl.AsButton(sender).Caption())
		})
	}

	clearBtn := lcl.NewButton(c)
	clearBtn.SetParent(c)
	clearBtn.SetCaption("C")
	clearBtn.SetLeft(10)
	clearBtn.SetTop(360)
	clearBtn.SetWidth(290)
	clearBtn.SetHeight(50)
	clearBtn.SetColor(colors.ClLightgray)
	clearBtn.SetOnClick(func(sender lcl.IObject) {
		c.Result = 0
		c.Operator = ""
		c.NewNumber = true
		c.Display.SetText("0")
	})

	c.NewNumber = true
}

func (c *TCalculatorForm) OnButtonClick(value string) {
	switch value {
	case "+", "-", "*", "/":
		if c.Operator != "" && !c.NewNumber {
			c.Calculate()
		}
		current, _ := strconv.ParseFloat(c.Display.Text(), 64)
		c.Result = current
		c.Operator = value
		c.NewNumber = true

	case "=":
		if c.Operator != "" {
			c.Calculate()
			c.Operator = ""
			c.NewNumber = true
		}

	default:
		if c.NewNumber {
			c.Display.SetText(value)
			c.NewNumber = false
		} else {
			if c.Display.Text() == "0" {
				c.Display.SetText(value)
			} else {
				c.Display.SetText(c.Display.Text() + value)
			}
		}
	}
}

func (c *TCalculatorForm) Calculate() {
	current, _ := strconv.ParseFloat(c.Display.Text(), 64)
	var result float64

	switch c.Operator {
	case "+":
		result = c.Result + current
	case "-":
		result = c.Result - current
	case "*":
		result = c.Result * current
	case "/":
		if current != 0 {
			result = c.Result / current
		} else {
			c.Display.SetText("错误")
			return
		}
	}

	c.Display.SetText(strconv.FormatFloat(result, 'f', -1, 64))
	c.Result = result
}
