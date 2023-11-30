package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func main() {
	Auto()
}

func Auto() {
	docmap := make(map[string]string)
	// 获取文件注释
	// 解析文件
	workPath := "./"
	fileList := []string{}
	err := filepath.Walk(workPath+"./controllers", func(path string, f os.FileInfo, err error) error {
		if filepath.Ext(path) == ".go" {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	fset := token.NewFileSet()
	for _, file := range fileList {
		f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			fmt.Println(err)
			continue
		}
		for _, decl := range f.Decls {
			f2, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			docmap[strings.ToUpper(getFullName(f2))] = strings.ReplaceAll(f2.Doc.Text(), "\n", "  ")
		}
	}
	// 从文件中解析 AST
	var content []string
	content = append(content, "package controllers\n")
	fileName := workPath + "./controllers/output" + fmt.Sprint(time.Now().Unix()) + ".go"
	num := 1
	fset = token.NewFileSet()
	f, err := parser.ParseFile(fset, workPath+"./routers/router.go", nil, parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}

	// 遍历 AST 查找函数调用
	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			funcName, ok := node.Fun.(*ast.Ident)
			if ok && funcName.Name == "Post" {
				group := fmt.Sprint(node.Args[0])
				met := fmt.Sprint(node.Args[2]) // &{contr GetUser}
				met = met[strings.Index(met, " ")+1 : len(met)-1]
				chi := strings.ToLower(string(met[0])) + met[1:]
				fmt.Printf("index: %v type: Post path: /%s/%s  method: %s \n", num, group, chi, met)
				content = append(content, fmt.Sprintf("// @Tags %v\n", group))
				content = append(content, "// @Accept json\n")
				content = append(content, "// @Produce json\n")
				content = append(content, "// @Security ApiKeyAuth\n")
				content = append(content, fmt.Sprintf("// @Param data body req%s true \"raw\"\n", met))
				content = append(content, fmt.Sprintf("// @Success 200 {object} ret%s\n", met))
				content = append(content, fmt.Sprintf("// @Router /%s/%s [post]\n", group, chi))
				key := strings.ToUpper(group + "Controller." + met)
				content = append(content, fmt.Sprintf("// @Summary %s\n", docmap[key]))
				content = append(content, fmt.Sprintf("func Test%v() {}\n\n\n", num))
				num += 1
			}
		}
		return true
	})

	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			funcName, ok := node.Fun.(*ast.Ident)
			if ok && funcName.Name == "Get" {
				group := fmt.Sprint(node.Args[0])
				met := fmt.Sprint(node.Args[2]) // &{contr GetUser}
				met = met[strings.Index(met, " ")+1 : len(met)-1]
				chi := strings.ToLower(string(met[0])) + met[1:]
				fmt.Printf("index: %v type: Post path: /%s/%s  method: %s \n", num, group, chi, met)
				content = append(content, fmt.Sprintf("// @Tags %v\n", group))
				content = append(content, "// @Accept json\n")
				content = append(content, "// @Produce json\n")
				content = append(content, "// @Security ApiKeyAuth\n")
				content = append(content, fmt.Sprintf("// @Param data body req%s true \"raw\"\n", met))
				content = append(content, fmt.Sprintf("// @Success 200 {object} ret%s\n", met))
				content = append(content, fmt.Sprintf("// @Router /%s/%s [post]\n", group, chi))
				key := strings.ToUpper(group + "Controller." + met)
				content = append(content, fmt.Sprintf("// @Summary %s\n", docmap[key]))
				content = append(content, fmt.Sprintf("func Test%v() {}\n\n\n", num))
				num += 1
			}
		}
		return true
	})
	Write(content, fileName)
	SwagInit()
	err2 := os.Remove(fileName)
	if err2 != nil {
		fmt.Printf("删除文件失败,文件名:%s,err:%v", fileName, err2)
	}
}

func Write(content []string, fileName string) {
	// 创建文件
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// 创建带缓冲的写入器
	writer := bufio.NewWriter(file)

	// 写入字符串
	for _, v := range content {
		writer.WriteString(v)
	}
	// 确保所有缓冲的数据都写入文件
	err = writer.Flush()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// 关闭文件
	err = file.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func SwagInit() {
	fmt.Println("\n\n exec swag init")
	cmd := exec.Command("swag", "init")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println(string(out))
}

func CreateJwt(uid string) (string, error) {
	claims := jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour * 24 * 100).Unix(),
		Id:        uid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte("lizenghui"))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func getFullName(f *ast.FuncDecl) string {
	var name string
	if f.Recv != nil {
		if recv, ok := f.Recv.List[0].Type.(*ast.Ident); ok {
			name = recv.Name + "."
		}
	}
	return name + f.Name.Name
}
