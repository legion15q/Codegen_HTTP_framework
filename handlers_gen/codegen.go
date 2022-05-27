package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type ApiError struct {
	HTTPStatus int
	Err        error
}

func main() {
	fset := token.NewFileSet()
	//os_Args_1 := "D:/Projects/Go/coursera/part_2/hw5_codegen/api.go"
	//os_Args_2 := "D:/Projects/Go/coursera/part_2/hw5_codegen/api_handlers.go"
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	out, _ := os.Create(os.Args[2])
	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out) // empty line
	fmt.Fprintln(out, `import "net/http"`)
	fmt.Fprintln(out, `import "strconv"`)
	fmt.Fprintln(out, `import "context"`)
	fmt.Fprintln(out, `import "fmt"`)
	fmt.Fprintln(out) // empty line

	multiplexers_for_ServeHTTP := make(map[string][]map[string]string)
	fmt.Println("Collecting Api_struct and method for api")
	fmt.Println("--------------------------------------")
	//в первом проходе собираем все api_struct и методы для них
	for _, node := range node.Decls {
		FD, ok := node.(*ast.FuncDecl)
		if !ok {
			fmt.Printf("SKIP %T is not *ast.FuncDecl\n", node)
			continue
		}
		if FD.Doc == nil {
			fmt.Printf("SKIP func %#v doesnt have comments\n", FD.Name.Name)
			continue
		}
		//
		needCodegen := false
		for _, comment := range FD.Doc.List {
			needCodegen = needCodegen || strings.HasPrefix(comment.Text, "// apigen:api")
		}
		if !needCodegen {
			fmt.Printf("SKIP func %#v doesnt have apigen mark\n", FD.Name.Name)
			continue
		}

		api_struct_name := FD.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
		comment := FD.Doc.List[0].Text
		comment = strings.Trim(comment, "// apigen:api ")

		type to_generate struct {
			Url    string `json:"url"`
			Auth   bool   `json:"auth"`
			Method string `json:"method"`
		}
		to_gen := to_generate{}
		json.Unmarshal([]byte(comment), &to_gen)

		fmt.Println("Find api struct:", api_struct_name)
		multiplexers_for_ServeHTTP[api_struct_name] = append(multiplexers_for_ServeHTTP[api_struct_name], map[string]string{"Handler" + FD.Name.Name: to_gen.Url})
		fmt.Println("Find new multiplexer", "Handler"+FD.Name.Name, "for api", api_struct_name)

		var need_struct_type *ast.Ident
		for _, func_param := range FD.Type.Params.List {
			if id, ok := func_param.Type.(*ast.Ident); ok {
				need_struct_type = id
			} else {
				fmt.Println("SKIP Context")
			}
		}
		fmt.Fprintln(out, "func (obj *"+api_struct_name+") "+"Handler"+FD.Name.Name+"(w http.ResponseWriter, r *http.Request) {")
		if to_gen.Method == "POST" {
			out.Write([]byte("\n" + `	if r.Method != "POST" {
				w.WriteHeader(http.StatusNotAcceptable)
				w.Write([]byte(` + "`" + `{"error": "bad method"}` + "`" + `))
				return
			}` + "\n"))
		}
		if to_gen.Auth {
			out.Write([]byte("\n" + `if r.Header.Get("X-Auth") != "100500" {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(` + "`" + `{"error": "unauthorized"}` + "`" + `))
				return
			}` + "\n"))
		}

		fmt.Println("need struct:", need_struct_type.Name)

		fmt.Fprintln(out, "ctx := context.Background()")
		fmt.Fprintln(out, "in, api_err :=Validate"+need_struct_type.Name+"(r)")
		out.Write([]byte(`
		if api_err != nil {
			w.WriteHeader(api_err.HTTPStatus)
			w.Write([]byte(` + "`" + `{"error": "` + "`" + `+ api_err.Error() +` + "`" + `"}` + "`" + `))
			return
		}` + "\n"))
		fmt.Fprintln(out, "resp, err := obj."+FD.Name.Name+"(ctx, *in)")
		out.Write([]byte(`if err != nil {
			switch t := err.(type) {
			case ApiError:
				w.WriteHeader(t.HTTPStatus)
				w.Write([]byte(` + "`" + `{"error": "` + "`" + `+ t.Error() +` + "`" + `"}` + "`" + `))
			case error:
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(` + "`" + `{"error": "` + "`" + `+ err.Error() +` + "`" + `"}` + "`" + `))
			}
			return
		}` + "\n"))
		return_struct_name := FD.Type.Results.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name

		fmt.Fprintln(out, "json_out, _ := CompileJSON"+return_struct_name+"("+"resp"+")")
		fmt.Fprintln(out, "w.Write(json_out)")
		fmt.Fprintln(out, "}")

	}
	fmt.Println("--------------------------------------")
	fmt.Println("Generating validator and json compiler")
	fmt.Println("--------------------------------------")
	// во втором проходе генерируем создатель json-а и валидаторы структур
	for _, node := range node.Decls {
		GD, ok := node.(*ast.GenDecl)
		if !ok {
			fmt.Printf("SKIP %T is not *ast.GenDecl\n", node)
			continue
		}
		for _, spec := range GD.Specs {
			currType, ok := spec.(*ast.TypeSpec)
			if !ok {
				fmt.Printf("SKIP %T is not ast.TypeSpec\n", spec)
				continue
			}

			currStruct, ok := currType.Type.(*ast.StructType)
			if !ok {
				fmt.Printf("SKIP %T is not ast.StructType\n", currStruct)
				continue
			}

			needCodegen_validator := false
			needCodegen_json_compiler := false
			for _, field := range currStruct.Fields.List {
				if field.Tag != nil {
					//fmt.Println(field.Tag.Value)
					if strings.HasPrefix(field.Tag.Value, "`apivalidator:") {
						needCodegen_validator = true
					}
					if strings.HasPrefix(field.Tag.Value, "`json:") {
						needCodegen_json_compiler = true
					}
				}

			}
			if !needCodegen_validator && !needCodegen_json_compiler {
				fmt.Printf("SKIP struct %#v doesnt have apivalidator or json mark\n", currType.Name.Name)
				continue
			}
			if needCodegen_validator {
				fmt.Fprintln(out, "func Validate"+currType.Name.Name+"("+"r *http.Request) (*"+currType.Name.Name+", *ApiError) {")
				fmt.Fprintln(out, "obj := "+currType.Name.Name+"{}")
				fmt.Fprintln(out, "var req int")
				fmt.Fprintln(out, "_ = req")
				fmt.Fprintln(out, "var err_conv error")
				fmt.Fprintln(out, "_ = err_conv")
				fmt.Fprintln(out, `default_val := ""`)
				for _, field := range currStruct.Fields.List {
					tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
					tag_value := tag.Get("apivalidator")
					if tag_value == "" || tag_value == "-" {
						continue
					}
					commands := strings.Split(tag_value, ",")
					field_name := field.Names[0].Name
					request_name := strings.ToLower(field_name)
					field_type_name := field.Type.(*ast.Ident).Name
					var int_value bool = false
					if strings.Contains(tag_value, "paramname=") {
						temp := strings.Split(tag_value, "paramname=")
						param_val := strings.Split(temp[1], ",")
						if param_val[0] == temp[1] {
							param_val = strings.Split(temp[1], `"`)
							request_name = param_val[0]
						} else {
							request_name = param_val[0]
						}
						fmt.Println(request_name)

					}
					if strings.Contains(tag_value, "default") {
						temp := strings.Split(tag_value, "default=")
						param_val := strings.Split(temp[1], ",")
						default_val := ""
						if param_val[0] == temp[1] {
							param_val = strings.Split(temp[1], `"`)
							default_val = param_val[0]
						} else {
							default_val = param_val[0]
						}
						fmt.Println(default_val)
						fmt.Fprintln(out, "if r.FormValue("+`"`+request_name+`"`+")"+`== ""`+"{")
						fmt.Fprintln(out, "default_val"+`="`+default_val+`"`)
						fmt.Fprintln(out, "}")

					}
					for _, command := range commands {

						if strings.HasPrefix(command, "required") {
							fmt.Fprintln(out, "if r.FormValue("+`"`+request_name+`"`+")"+` == ""`+"{")
							fmt.Fprintln(out, `err := ApiError{http.StatusBadRequest, fmt.Errorf("`+request_name+` must be not empty")}`)
							fmt.Fprintln(out, `return nil, &err`)
							fmt.Fprintln(out, "}")
						}

						if strings.HasPrefix(command, "min") {
							command_value := strings.Split(command, "=")
							if len(command_value) > 1 {
								if field_type_name == "string" {
									fmt.Fprintln(out, "if len(r.FormValue("+`"`+request_name+`"`+")"+")<"+command_value[1]+"{")
									fmt.Fprintln(out, `err := ApiError{http.StatusBadRequest, fmt.Errorf("`+request_name+` len must be >= `+command_value[1]+`")}`)
									fmt.Fprintln(out, `return nil, &err`)
									fmt.Fprintln(out, "}")
									int_value = false
								} else if field_type_name == "int" {
									fmt.Fprintln(out, `req, err_conv = strconv.Atoi(r.FormValue(`+`"`+request_name+`"`+`))`)
									fmt.Fprintln(out, "if err_conv != nil {")
									fmt.Fprintln(out, `err := ApiError{http.StatusBadRequest, fmt.Errorf("`+request_name+` must be int")}`)
									fmt.Fprintln(out, `return nil, &err`)
									fmt.Fprintln(out, "}")
									fmt.Fprintln(out, "if req"+"<"+command_value[1]+"{")
									fmt.Fprintln(out, `err := ApiError{http.StatusBadRequest, fmt.Errorf("`+request_name+` must be >= `+command_value[1]+`")}`)
									fmt.Fprintln(out, `return nil, &err`)
									fmt.Fprintln(out, "}")
									int_value = true
								}

							} else {
								log.Fatal("error, need min value of field ", field_name, " of struct ", currType.Name.Name)
								return
							}

						}
						if strings.HasPrefix(command, "max") {
							command_value := strings.Split(command, "=")
							if len(command_value) > 1 {
								if field_type_name == "string" {
									fmt.Fprintln(out, "if len(r.FormValue("+`"`+request_name+`"`+")"+")>"+command_value[1]+"{")
									fmt.Fprintln(out, `err := ApiError{http.StatusBadRequest, fmt.Errorf("`+request_name+` len must be <= `+command_value[1]+`")}`)
									fmt.Fprintln(out, `return nil, &err`)
									fmt.Fprintln(out, "}")
									int_value = false
								} else if field_type_name == "int" {
									fmt.Fprintln(out, `req, err_conv = strconv.Atoi(r.FormValue(`+`"`+request_name+`"`+`))`)
									fmt.Fprintln(out, "if err_conv != nil {")
									fmt.Fprintln(out, `err := ApiError{http.StatusBadRequest, fmt.Errorf("`+request_name+` must be int")}`)
									fmt.Fprintln(out, `return nil, &err`)
									fmt.Fprintln(out, "}")
									fmt.Fprintln(out, "if req >"+command_value[1]+`{`)
									fmt.Fprintln(out, `err := ApiError{http.StatusBadRequest, fmt.Errorf("`+request_name+` must be <= `+command_value[1]+`")}`)
									fmt.Fprintln(out, `return nil, &err`)
									fmt.Fprintln(out, "}")
									int_value = true
								}
							} else {
								log.Fatal("error, need min value of field ", field_name, " of struct ", currType.Name.Name)
								return
							}

						}

						if strings.HasPrefix(command, "enum") {
							enum := strings.Split(command, "=")
							enum_values := strings.Split(enum[1], "|")
							fmt.Fprintln(out, "k := 0")
							enum_values_str := ""
							for idx, val := range enum_values {
								fmt.Fprintln(out, "if r.FormValue("+`"`+request_name+`"`+")"+` != "`+val+`" && default_val != "`+val+`"{`)
								fmt.Fprintln(out, "k++")
								if idx != len(enum_values)-1 {
									enum_values_str = enum_values_str + val + ", "
								} else {
									enum_values_str = enum_values_str + val
								}

								fmt.Fprintln(out, "}")
							}
							fmt.Fprintln(out, "if k =="+strconv.Itoa(len(enum_values))+"{")
							fmt.Fprintln(out, `err := ApiError{http.StatusBadRequest, fmt.Errorf(`+`"`+request_name+` must be one of [`+enum_values_str+`]")}`)
							fmt.Fprintln(out, `return nil, &err`)
							fmt.Fprintln(out, "}")
						}

					}
					if int_value {
						fmt.Fprintln(out, "obj."+field_name+"=req")
					} else {
						fmt.Fprintln(out, `if default_val == ""{`)
						fmt.Fprintln(out, "obj."+field_name+"=r.FormValue("+`"`+request_name+`"`+")")
						fmt.Fprintln(out, "}else{")
						fmt.Fprintln(out, "obj."+field_name+"=default_val")
						fmt.Fprintln(out, "}")

					}

				}
				fmt.Fprintln(out, "return &obj, nil")
				fmt.Fprintln(out, "}")
			}

			if needCodegen_json_compiler {
				fmt.Fprintln(out, "func CompileJSON"+currType.Name.Name+"("+"obj *"+currType.Name.Name+") ([]byte, error) {")
				fmt.Fprintln(out, `result := "{" + "\n"`)
				fmt.Fprintln(out, "result = result +"+`"\"`+"error"+`\""`+`+ ":" + `+`"\""`+`+ "\""`+`+ ","`+`+ "\n"`)
				fmt.Fprintln(out, "result = result +"+`"\"`+"response"+`\""`+`+ ":" + `+`"{"`+`+ "\n"`)
				for idx, field := range currStruct.Fields.List {
					tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
					tag_value := tag.Get("json")
					if tag_value == "" || tag_value == "-" {
						continue
					}
					field_name := field.Names[0].Name
					field_type_name := field.Type.(*ast.Ident).Name
					new_field_name := "obj." + field_name
					if field_type_name == "uint64" {
						new_field_name = "strconv.FormatUint(" + "obj." + field_name + ",10)"
					} else if field_type_name == "int64" {
						new_field_name = "strconv.FormatInt(" + "obj." + field_name + ",10)"
					} else if field_type_name == "int" {
						new_field_name = "strconv.FormatInt(int64(" + "obj." + field_name + "),10)"
					}
					if idx != len(currStruct.Fields.List)-1 {
						if field_type_name == "string" {
							fmt.Fprintln(out, "result = result +"+`"\"`+tag_value+`\""`+`+ ":" + `+`"\"" +`+new_field_name+`+ "\""`+`+ ","`+`+ "\n"`)
						} else {
							fmt.Fprintln(out, "result = result +"+`"\"`+tag_value+`\""`+`+ ":" + `+new_field_name+`+ ","`+`+ "\n"`)
						}
					} else {
						if field_type_name == "string" {
							fmt.Fprintln(out, "result = result +"+`"\"`+tag_value+`\""`+`+ ":" + `+`"\"" +`+new_field_name+`+ "\""`+`+ "\n"`)
						} else {
							fmt.Fprintln(out, "result = result +"+`"\"`+tag_value+`\""`+`+ ":" + `+new_field_name+`+ "\n"`)
						}

					}

				}
				fmt.Fprintln(out, `result = result  + "}"`)
				fmt.Fprintln(out, `result = result  + "}"`)
				fmt.Fprintln(out, `return []byte(result), nil`)
				fmt.Fprintln(out, "}")
			}

		}
	}
	fmt.Println("--------------------------------------")
	fmt.Println(multiplexers_for_ServeHTTP)
	for api, slice_of_multiplexer_to_url_map := range multiplexers_for_ServeHTTP {
		fmt.Fprintln(out, "func (obj *"+api+") ServeHTTP(w http.ResponseWriter, r *http.Request) {")
		fmt.Fprintln(out, "switch r.URL.Path {")
		for _, multiplexer_to_url_map := range slice_of_multiplexer_to_url_map {
			for key, val := range multiplexer_to_url_map {
				fmt.Fprintln(out, "case "+`"`+val+`"`+":"+"\n"+"obj."+key+"(w,r)")
			}
		}
		out.Write([]byte("\n" + `	default:
			// 404
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(` + "`" + `{"error": "unknown method"}` + "`" + `))
			return
		}`))
		fmt.Fprintln(out, "}")
	}

}
