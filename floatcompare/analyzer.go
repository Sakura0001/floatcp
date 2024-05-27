package floatcompare

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

type floatcompare struct {
	equalOnly bool //仅进行相等比较
	skipTests bool //跳过测试
}

// NewAnalyzer create a new analyzer for float compare
func NewAnalyzer() *analysis.Analyzer {
	fc := floatcompare{}
	var flagSet flag.FlagSet

	flagSet.BoolVar(&fc.equalOnly, "equalOnly", false, "should the linter only search for == and !=")
	flagSet.BoolVar(&fc.skipTests, "skipTests", false, "should the linter execute on test files as well")

	return &analysis.Analyzer{
		Name:  "floatcompare",
		Doc:   "Search for float comparison, since these are potential errors",
		Run:   fc.run,
		Flags: flagSet,
	}
}

func (fc *floatcompare) isCheckExpr(node ast.Node, pass *analysis.Pass) {
	switch expr := node.(type) {
	//如果是二元表达式，调用checkBinExpr方法检查是否是浮点数比较。
	case *ast.BinaryExpr:
		fc.checkBinExpr(expr, pass)
	//如果是switch语句，则检查标签是否为浮点数
	case *ast.SwitchStmt:
		if fc.isFloat(expr.Tag, pass) {
			pass.Reportf(expr.Tag.Pos(), "float comparison with switch statement")
			return
		}
	}
}

func (fc *floatcompare) isFloat(expr ast.Expr, pass *analysis.Pass) bool {
	//检查是否为浮点数
	t := pass.TypesInfo.TypeOf(expr)
	if t == nil {
		return false
	}
	bt, ok := t.Underlying().(*types.Basic)
	if !ok {
		return false
	}
	if (bt.Info() & types.IsFloat) == 0 {
		return false
	}
	return true
}

func (fc *floatcompare) checkBinExpr(binExpr *ast.BinaryExpr, pass *analysis.Pass) {
	//检查 equalOnly 标志和操作符 是否为==和！=
	if fc.equalOnly && !(binExpr.Op == token.EQL || binExpr.Op == token.NEQ) {
		return
	}

	//检查是否为常见的比较操作符
	if !(binExpr.Op == token.EQL || binExpr.Op == token.LEQ || binExpr.Op == token.LSS || binExpr.Op == token.GEQ || binExpr.Op == token.GTR || binExpr.Op == token.NEQ) {
		return
	}

	//检查是否为浮点数
	if !fc.isFloat(binExpr.X, pass) || !fc.isFloat(binExpr.Y, pass) {
		return
	}

	//全部通过
	pass.Reportf(binExpr.Pos(), "float comparison found %q",
		render(pass.Fset, binExpr))

}

func (fc *floatcompare) run(pass *analysis.Pass) (interface{}, error) {
	//遍历文件，检查是否为test结尾
	for _, f := range pass.Files {
		if fc.skipTests && strings.HasSuffix(pass.Fset.Position(f.Pos()).Filename, "_test.go") {
			continue
		}
		//遍历AST树，查找是否为二元运算
		ast.Inspect(f, func(node ast.Node) bool {
			fc.isCheckExpr(node, pass)
			return true
		})
	}
	return nil, nil
}
func render(fset *token.FileSet, x interface{}) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, x); err != nil {
		return fmt.Sprintf("ERROR during token parsing: %v", err)
	}
	return buf.String()
}
