package parquet_go

import (
	//"log"
	. "SchemaHandler"
	"reflect"
)

type Node struct {
	Val  reflect.Value
	Path []string
	RL   int32
	DL   int32
}

func Marshal(srcInterface interface{}, bgn int, end int, schemaHandler *SchemaHandler) *map[string]*Table {
	src := reflect.ValueOf(srcInterface)
	res := make(map[string]*Table)
	for i := 0; i < len(schemaHandler.SchemaElements); i++ {
		schema := schemaHandler.SchemaElements[i]
		pathStr := schemaHandler.IndexMap[i]
		numChildren := schema.GetNumChildren()
		if numChildren == 0 {
			res[pathStr] = new(Table)
			res[pathStr].Path = StrToPath(pathStr)
			res[pathStr].MaxDefinitionLevel, _ = schemaHandler.MaxDefinitionLevel(res[pathStr].Path)
			res[pathStr].MaxRepetitionLevel, _ = schemaHandler.MaxRepetitionLevel(res[pathStr].Path)
			res[pathStr].Repetition_Type = schema.GetRepetitionType()
		}
	}

	rootName := schemaHandler.GetRootName()
	for i := bgn; i < end; i++ {
		stack := make([]*Node, 0)
		node := new(Node)
		node.Val = src.Index(i)
		node.Path = append(node.Path, rootName)
		stack = append(stack, node)

		for len(stack) > 0 {
			ln := len(stack)
			node := stack[ln-1]
			stack = stack[:ln-1]

			if node.Val.Type().Kind() == reflect.Ptr {
				if node.Val.IsNil() {
					pathStr := PathToStr(node.Path)
					for key, table := range res {
						if len(key) >= len(pathStr) && key[:len(pathStr)] == pathStr {
							table.Values = append(table.Values, nil)
							table.DefinitionLevels = append(table.DefinitionLevels, node.DL)
							table.RepetitionLevels = append(table.RepetitionLevels, node.RL)
						}
					}
				} else {
					node.Val = node.Val.Elem()
					node.DL++
					stack = append(stack, node)
				}
			} else if node.Val.Type().Kind() == reflect.Struct {
				numField := node.Val.Type().NumField()
				for j := 0; int32(j) < numField; j++ {
					tf := node.Val.Type().Field(j)
					name := tf.Name
					newNode := new(Node)
					newNode.Path = append(newNode.Path, node.Path...)
					newNode.Path = append(newNode.Path, name)
					newNode.Val = node.Val.FieldByName(name)
					newNode.RL = node.RL
					newNode.DL = node.DL
					stack = append(stack, newNode)
				}
			} else if node.Val.Type().Kind() == reflect.Slice {
				ln := node.Val.Len()
				path := make([]string, 0)
				path = append(path, node.Path...)
				path = append(path, "list", "element")

				if ln <= 0 {
					pathStr := PathToStr(node.Path)
					for key, table := range res {
						if len(key) >= len(pathStr) && key[:len(pathStr)] == pathStr {
							table.Values = append(table.Values, nil)
							table.DefinitionLevels = append(table.DefinitionLevels, node.DL)
							table.RepetitionLevels = append(table.RepetitionLevels, node.RL)
						}
					}
				}

				rl, _ := schemaHandler.MaxRepetitionLevel(path)
				for j := ln - 1; j >= 0; j-- {
					newNode := new(Node)
					newNode.Path = path
					newNode.Val = node.Val.Index(j)
					if j == 0 {
						newNode.RL = node.RL
					} else {
						newNode.RL = rl
					}
					newNode.DL = node.DL + 1 //list is repeated
					stack := append(stack, newNode)
				}
			} else if node.Val.Type().Kind() == reflect.Map {
				path := make([]string, 0)
				path = append(path, node.Path...)
				path = append(path, "key_value")

			} else {
			}

		}
	}

}