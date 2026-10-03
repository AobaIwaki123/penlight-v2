package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

type FieldInfo struct {
	Name    string
	Type    string
	IsPK    bool
	IsFK    bool
	IsUK    bool
	Comment string
}

type EntityInfo struct {
	Name   string
	Fields []FieldInfo
}

type RelationInfo struct {
	From  string
	To    string
	Label string
}

func main() {
	modelDir := filepath.Join("pkg", "model")
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, modelDir, nil, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse model dir: %v\n", err)
		os.Exit(1)
	}

	entities := make(map[string]*EntityInfo)
	var relations []RelationInfo

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.TYPE {
					continue
				}

				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}

					entityName := strings.ToUpper(toSnakeCase(typeSpec.Name.Name))
					// スキップ対象（組み込み小オブジェクト等）
					if entityName == "PENLIGHT_PAIR" || entityName == "QUIZ_OPTION" {
						continue
					}

					entity := &EntityInfo{Name: entityName}

					for _, field := range structType.Fields.List {
						if len(field.Names) == 0 {
							continue
						}
						fieldName := field.Names[0].Name
						fieldType := exprToString(field.Type)

						var tagValue string
						if field.Tag != nil {
							tagValue = strings.Trim(field.Tag.Value, "`")
						}

						dbTag := reflect.StructTag(tagValue).Get("db")
						jsonTag := reflect.StructTag(tagValue).Get("json")

						if dbTag == "-" {
							continue
						}

						// JSON名またはスネークケースをカラム名に
						colName := toSnakeCase(fieldName)
						if jsonTag != "" {
							parts := strings.Split(jsonTag, ",")
							if parts[0] != "" {
								colName = parts[0]
							}
						}

						isPK := strings.Contains(dbTag, "pk") || colName == "id"
						isFK := strings.Contains(dbTag, "fk") || (strings.HasSuffix(colName, "_id") && !isPK)
						isUK := strings.Contains(dbTag, "uk") || colName == "slug" || colName == "google_sub"

						comment := ""
						if field.Comment != nil {
							comment = strings.TrimSpace(field.Comment.Text())
						}

						entity.Fields = append(entity.Fields, FieldInfo{
							Name:    colName,
							Type:    mapTypeToMermaid(fieldType),
							IsPK:    isPK,
							IsFK:    isFK,
							IsUK:    isUK,
							Comment: comment,
						})

						// 外部キーからリレーションを自動推論
						if isFK {
							targetEntity := inferTargetEntity(colName)
							if targetEntity != "" {
								relations = append(relations, RelationInfo{
									From:  targetEntity,
									To:    entityName,
									Label: colName,
								})
							}
						}
					}

					entities[entityName] = entity
				}
			}
		}
	}

	// Mermaid ER 図の構築
	var buf bytes.Buffer
	buf.WriteString("erDiagram\n")

	// 1. リレーション出力（決定論的ソート）
	sort.Slice(relations, func(i, j int) bool {
		if relations[i].From != relations[j].From {
			return relations[i].From < relations[j].From
		}
		if relations[i].To != relations[j].To {
			return relations[i].To < relations[j].To
		}
		return relations[i].Label < relations[j].Label
	})
	for _, rel := range relations {
		buf.WriteString(fmt.Sprintf("    %s ||--o{ %s : \"%s\"\n", rel.From, rel.To, rel.Label))
	}
	buf.WriteString("\n")

	// 2. エンティティ定義出力
	entityOrder := []string{"GROUP", "COLOR", "MEMBER", "USER", "ANSWER_LOG", "QUIZ_QUESTION"}
	for _, name := range entityOrder {
		ent, ok := entities[name]
		if !ok {
			continue
		}
		buf.WriteString(fmt.Sprintf("    %s {\n", ent.Name))
		for _, f := range ent.Fields {
			attrType := f.Type
			keyModifier := ""
			if f.IsPK {
				keyModifier = " PK"
			} else if f.IsFK {
				keyModifier = " FK"
			} else if f.IsUK {
				keyModifier = " UK"
			}

			commentStr := ""
			if f.Comment != "" {
				// 特殊文字をエスケープしてダブルクォートで囲む
				safeComment := strings.ReplaceAll(f.Comment, "\"", "'")
				commentStr = fmt.Sprintf(" \"%s\"", safeComment)
			}

			buf.WriteString(fmt.Sprintf("        %s %s%s%s\n", attrType, f.Name, keyModifier, commentStr))
		}
		buf.WriteString("    }\n\n")
	}

	outDir := filepath.Join("assets", "schema")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create out dir: %v\n", err)
		os.Exit(1)
	}

	mermaidPath := filepath.Join(outDir, "er-diagram.mermaid")
	if err := os.WriteFile(mermaidPath, buf.Bytes(), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write mermaid: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully generated from Go AST: %s\n", mermaidPath)

	mdPath := filepath.Join(outDir, "er-diagram.md")
	mdContent := fmt.Sprintf("# 自動生成データモデル ER 図 (Auto-Generated Data Model)\n\n"+
		"> 本ドキュメントは `scripts/gen-er-diagram.go` が `pkg/model/*.go` の Go AST（抽象構文木）を直接静的解析して完全自動生成した資産 (Asset) です。\n\n"+
		"```mermaid\n%s```\n", buf.String())

	if err := os.WriteFile(mdPath, []byte(mdContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write md: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully generated from Go AST: %s\n", mdPath)
}

func mapTypeToMermaid(goType string) string {
	switch {
	case strings.Contains(goType, "int"):
		return "int"
	case strings.Contains(goType, "bool"):
		return "boolean"
	case strings.Contains(goType, "Time"):
		return "datetime"
	default:
		return "string"
	}
}

func inferTargetEntity(fkCol string) string {
	switch fkCol {
	case "group_id":
		return "GROUP"
	case "left_color_id", "right_color_id", "color_id":
		return "COLOR"
	case "target_member_id", "member_id":
		return "MEMBER"
	case "user_id":
		return "USER"
	case "quiz_question_id":
		return "QUIZ_QUESTION"
	default:
		return ""
	}
}

func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	default:
		return fmt.Sprintf("%v", expr)
	}
}

func toSnakeCase(s string) string {
	var buf bytes.Buffer
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			buf.WriteByte('_')
		}
		buf.WriteRune(r)
	}
	return strings.ToLower(buf.String())
}
