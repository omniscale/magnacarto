package mss

import (
	"testing"

	"github.com/omniscale/magnacarto/color"

	"github.com/stretchr/testify/assert"
)

func testExpr(t *testing.T, a float64, op codeType, b, c float64) {
	e := expression{}
	e.addValue(a, typeNum)
	e.addValue(b, typeNum)
	e.addOperator(op)
	v, err := e.evaluate()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, c, v)
}

func TestSimpleExpression(t *testing.T) {
	testExpr(t, 1, typeAdd, 2, 3)
	testExpr(t, 1, typeSubtract, 2, -1)
	testExpr(t, 2, typeMultiply, 3, 6)
	testExpr(t, 8, typeDivide, 2, 4)
}

func TestListExpression(t *testing.T) {
	e := expression{}
	e.addValue(float64(4), typeNum)
	e.addValue(float64(10), typeNum)
	e.addValue(float64(2), typeNum)
	v, err := e.evaluate()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, []Value{float64(4), float64(10), float64(2)}, v)
}

func TestMultipleExpression(t *testing.T) {
	e := expression{}
	e.addValue(float64(4), typeNum)
	e.addValue(float64(10), typeNum)
	e.addOperator(typeAdd)
	e.addValue(float64(2), typeNum)
	e.addOperator(typeDivide)
	e.addValue(float64(5), typeNum)
	e.addValue(float64(1), typeNum)
	e.addOperator(typeSubtract)
	e.addOperator(typeMultiply)
	v, err := e.evaluate()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 28, v)
}

func TestFunctionExpression(t *testing.T) {
	e := expression{}
	e.addValue("__echo__", typeFunction)
	e.addValue(float64(4), typeNum)
	e.addValue(float64(10), typeNum)
	e.addOperator(typeAdd)
	e.addValue(nil, typeFunctionEnd)

	v, err := e.evaluate()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 14, v)

	e = expression{}
	e.addValue("__echo__", typeFunction)
	e.addValue("__echo__", typeFunction)
	e.addValue(float64(4), typeNum)
	e.addValue(float64(10), typeNum)
	e.addOperator(typeAdd)
	e.addValue(nil, typeFunctionEnd)
	e.addValue(float64(3), typeNum)
	e.addOperator(typeAdd)
	e.addValue(nil, typeFunctionEnd)
	e.addValue(float64(7), typeNum)
	e.addOperator(typeAdd)

	v, err = e.evaluate()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 24, v)

	e = expression{}
	e.addValue("__echo__", typeFunction)
	e.addValue(float64(4), typeNum)
	e.addValue(float64(10), typeNum)
	e.addOperator(typeAdd)
	e.addValue(float64(1), typeNum)
	e.addValue(nil, typeFunctionEnd)
	e.addOperator(typeMultiply)

	v, err = e.evaluate()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 14, v)
}

func TestLightenFunctionExpression(t *testing.T) {
	e := expression{}
	e.addValue("lighten", typeFunction)
	rgba, err := color.Parse("#ff20f0")
	assert.NoError(t, err)
	e.addValue(rgba, typeColor)
	e.addValue(float64(10), typeNum)
	e.addValue(nil, typeFunctionEnd)

	v, err := e.evaluate()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "#fe53f3", v.(color.RGBA).Hex())
}

func TestExpressionErrors(t *testing.T) {
	e := expression{}
	e.addValue("rgba", typeFunction)
	e.addValue(float64(0), typeNum)
	e.addValue(float64(0), typeNum)
	e.addValue(float64(0), typeNum)
	e.addValue("", typeString)
	e.addValue(nil, typeFunctionEnd)

	_, err := e.evaluate()
	assert.Error(t, err)
}
