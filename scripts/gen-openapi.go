//go:build ignore

// Generate the OpenAPI document from the implemented HTTP surface and the Go
// model types (Ref: ADR-0004, ADR-0036, ADR-0037).
package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/aobaiwaki/penlight-v2/pkg/model"
)

type schemaGenerator struct {
	modelPackage string
	seen         map[string]reflect.Type
	order        []string
}

func newSchemaGenerator() *schemaGenerator {
	return &schemaGenerator{
		modelPackage: reflect.TypeOf(model.Member{}).PkgPath(),
		seen:         make(map[string]reflect.Type),
	}
}

func (g *schemaGenerator) register(t reflect.Type) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		g.register(t.Elem())
		return
	}
	if t.Kind() == reflect.Map {
		g.register(t.Key())
		g.register(t.Elem())
		return
	}
	if t == reflect.TypeOf(time.Time{}) || t.PkgPath() != g.modelPackage || t.Name() == "" {
		return
	}
	if _, ok := g.seen[t.Name()]; ok {
		return
	}
	g.seen[t.Name()] = t
	g.order = append(g.order, t.Name())
	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath == "" && jsonName(field) != "-" {
				g.register(field.Type)
			}
		}
	}
}

func jsonName(field reflect.StructField) string {
	name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
	if name == "" {
		return field.Name
	}
	return name
}

func isOmitEmpty(field reflect.StructField) bool {
	_, options, _ := strings.Cut(field.Tag.Get("json"), ",")
	return strings.Contains(","+options+",", ",omitempty,")
}

func (g *schemaGenerator) writeType(w *strings.Builder, indent string, t reflect.Type) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == reflect.TypeOf(time.Time{}) {
		fmt.Fprintf(w, "%stype: string\n%sformat: date-time\n", indent, indent)
		return
	}
	if t.PkgPath() == g.modelPackage && t.Name() != "" {
		g.register(t)
		fmt.Fprintf(w, "%s$ref: \"#/components/schemas/%s\"\n", indent, t.Name())
		return
	}
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		fmt.Fprintf(w, "%stype: array\n%sitems:\n", indent, indent)
		g.writeType(w, indent+"  ", t.Elem())
	case reflect.Map:
		fmt.Fprintf(w, "%stype: object\n%sadditionalProperties: true\n", indent, indent)
	case reflect.Bool:
		fmt.Fprintf(w, "%stype: boolean\n", indent)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fmt.Fprintf(w, "%stype: integer\n", indent)
	case reflect.Float32, reflect.Float64:
		fmt.Fprintf(w, "%stype: number\n", indent)
	case reflect.String:
		fmt.Fprintf(w, "%stype: string\n", indent)
	default:
		fmt.Fprintf(w, "%stype: object\n", indent)
	}
}

func (g *schemaGenerator) writeSchema(w *strings.Builder, name string, t reflect.Type) {
	fmt.Fprintf(w, "    %s:\n", name)
	if name == "ID" {
		w.WriteString("      type: string\n      pattern: \"^(ser|grp|col|mem|sng|img|pht|quiz|usr|ans|prp)_[0-9a-f]{32}$\"\n")
		return
	}
	if name == "Prefix" {
		w.WriteString("      type: string\n      enum: [ser, grp, col, mem, sng, img, pht, quiz, usr, ans, prp]\n")
		return
	}
	if name == "MemberStatus" {
		w.WriteString("      type: string\n      enum: [active, graduated, hiatus]\n")
		return
	}
	if name == "MetadataEditProposalStatus" {
		w.WriteString("      type: string\n      enum: [pending, approved, rejected]\n")
		return
	}
	if name == "TargetType" {
		w.WriteString("      type: string\n      enum: [member, song]\n")
		return
	}
	if t.Kind() != reflect.Struct {
		g.writeType(w, "      ", t)
		return
	}
	w.WriteString("      type: object\n      properties:\n")
	required := make([]string, 0)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		name := jsonName(field)
		if name == "-" {
			continue
		}
		fmt.Fprintf(w, "        %s:\n", name)
		g.writeType(w, "          ", field.Type)
		if !isOmitEmpty(field) {
			required = append(required, name)
		}
	}
	if len(required) > 0 {
		w.WriteString("      required: [")
		for i, name := range required {
			if i > 0 {
				w.WriteString(", ")
			}
			w.WriteString(name)
		}
		w.WriteString("]\n")
	}
}

func writeHeader(w *strings.Builder) {
	w.WriteString(`openapi: 3.1.0
info:
  title: Penlight Quiz v2 API
  version: 1.0.0
  description: Go model and implemented HTTP contract (Ref: ADR-0004, ADR-0036, ADR-0037).
servers:
  - url: https://penlight.aooba.net
    description: Production Environment
  - url: http://localhost:8080
    description: Local Development Server
paths:
  /healthz:
    get:
      summary: ヘルスチェック
      responses:
        "200":
          description: 正常稼働
          content:
            text/plain:
              schema:
                type: string
  /api/v1/sync/bootstrap:
    get:
      summary: オフライン用マスタ一括取得
      parameters:
        - name: include_graduated
          in: query
          required: false
          schema:
            type: boolean
            default: false
          description: true の場合は卒業生も含める
      responses:
        "200":
          description: 最新マスタデータ
          headers:
            ETag:
              schema:
                type: string
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/BootstrapResponse"
        "304":
          description: 更新なし
        "400":
          $ref: "#/components/responses/InvalidParams"
  /api/v1/quiz/answers/batch:
    post:
      summary: オフライン回答ログ一括同期
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/BatchAnswerRequest"
      responses:
        "200":
          description: 同期結果
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/BatchAnswerResponse"
  /api/v1/quiz/statistics:
    get:
      summary: クイズ回答統計取得
      parameters:
        - name: user_id
          in: query
          schema:
            $ref: "#/components/schemas/ID"
        - name: group_id
          in: query
          schema:
            $ref: "#/components/schemas/ID"
        - name: target_type
          in: query
          schema:
            $ref: "#/components/schemas/TargetType"
        - name: limit
          in: query
          schema:
            type: integer
            default: 5
      responses:
        "200":
          description: 統計データ
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/QuizStatisticsResponse"
  /api/v1/members/{id}/metadata-edit-proposals:
    post:
      summary: メンバー編集提案の送信
      parameters:
        - name: id
          in: path
          required: true
          schema:
            $ref: "#/components/schemas/ID"
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/SubmitMetadataEditProposalRequest"
      responses:
        "200":
          description: 保存された提案（同一内容の再送は既存結果）
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/MetadataEditProposal"
        "400":
          $ref: "#/components/responses/InvalidParams"
        "404":
          $ref: "#/components/responses/NotFound"
  /api/v1/admin/metadata-edit-proposals:
    get:
      summary: メタデータ編集提案一覧
      parameters:
        - name: status
          in: query
          required: false
          schema:
            type: string
            enum: [pending, approved, rejected]
            default: pending
      responses:
        "200":
          description: 提案一覧
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/MetadataEditProposal"
        "400":
          $ref: "#/components/responses/InvalidParams"
  /api/v1/admin/metadata-edit-proposals/{id}/approve:
    post:
      summary: メタデータ編集提案の承認
      parameters:
        - name: id
          in: path
          required: true
          schema:
            $ref: "#/components/schemas/ID"
      responses:
        "200":
          description: 承認済み提案（再承認は既存結果）
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/MetadataEditProposal"
        "400":
          $ref: "#/components/responses/InvalidParams"
        "404":
          $ref: "#/components/responses/NotFound"
  /api/v1/admin/metadata-edit-proposals/{id}/reject:
    post:
      summary: メタデータ編集提案の却下
      parameters:
        - name: id
          in: path
          required: true
          schema:
            $ref: "#/components/schemas/ID"
      requestBody:
        required: false
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/RejectMetadataEditProposalRequest"
      responses:
        "200":
          description: 却下済み提案（再却下は既存結果）
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/MetadataEditProposal"
        "400":
          $ref: "#/components/responses/InvalidParams"
        "404":
          $ref: "#/components/responses/NotFound"
  /images/{key}:
    get:
      summary: 不変画像取得
      parameters:
        - name: key
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: 画像バイナリ
          content:
            image/webp:
              schema:
                type: string
                format: binary
        "404":
          $ref: "#/components/responses/NotFound"
components:
  responses:
    InvalidParams:
      description: パラメータ不正
      content:
        application/problem+json:
          schema:
            $ref: "#/components/schemas/AppError"
    NotFound:
      description: 対象が存在しない
      content:
        application/problem+json:
          schema:
            $ref: "#/components/schemas/AppError"
  schemas:
`)
}

func main() {
	generator := newSchemaGenerator()
	roots := []reflect.Type{
		reflect.TypeOf(model.AppError{}),
		reflect.TypeOf(model.BootstrapResponse{}),
		reflect.TypeOf(model.BatchAnswerRequest{}),
		reflect.TypeOf(model.BatchAnswerResponse{}),
		reflect.TypeOf(model.QuizStatisticsResponse{}),
		reflect.TypeOf(model.SubmitMetadataEditProposalRequest{}),
		reflect.TypeOf(model.RejectMetadataEditProposalRequest{}),
		reflect.TypeOf(model.MetadataEditProposal{}),
	}
	for _, root := range roots {
		generator.register(root)
	}

	var output strings.Builder
	writeHeader(&output)
	for _, name := range generator.order {
		generator.writeSchema(&output, name, generator.seen[name])
	}
	if err := os.WriteFile("api/openapi.yaml", []byte(output.String()), 0o644); err != nil {
		panic(fmt.Errorf("write api/openapi.yaml: %w", err))
	}
	fmt.Printf("Generated api/openapi.yaml: %d model schemas\n", len(generator.order))
}
