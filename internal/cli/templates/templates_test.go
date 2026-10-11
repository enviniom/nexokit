package templates

import (
	"bytes"
	"go/format"
	"testing"
)

func TestModuleFiles(t *testing.T) {
	for _, crud := range []bool{false, true} {
		data := ModuleData{Name: "order_items", Package: "order_items", Struct: "OrderItems", Table: "order_items", CRUD: crud}
		files := ModuleFiles(data)
		wantCount := 10
		if crud {
			wantCount = 16
		}
		if len(files) != wantCount {
			t.Fatalf("crud=%v files=%d want=%d", crud, len(files), wantCount)
		}
		seen := map[string]bool{}
		for _, file := range files {
			if seen[file.Path] {
				t.Fatalf("duplicate output %s", file.Path)
			}
			seen[file.Path] = true
			for _, tenant := range []bool{false, true} {
				tmpl, err := ParseModuleTemplate(file.Template)
				if err != nil {
					t.Fatal(err)
				}
				fileData := data
				fileData.Action, fileData.UseCase, fileData.Tenant = file.Action, file.UseCase, tenant
				var out bytes.Buffer
				if err := tmpl.Execute(&out, fileData); err != nil {
					t.Fatal(err)
				}
				if _, err := format.Source(out.Bytes()); err != nil {
					t.Fatalf("%s tenant=%v: %v\n%s", file.Path, tenant, err, out.String())
				}
			}
		}
		for _, path := range []string{"container.go", "routes.go", "core/contracts.go", "core/errors.go", "core/order_items/model.go", "core/order_items/dtos.go", "slices/get_order_items/handler.go", "slices/list_order_items/repository.go"} {
			if !seen[path] {
				t.Errorf("missing %s", path)
			}
		}
		if seen["slices/create_order_items/handler.go"] != crud {
			t.Errorf("mutation generation does not respect CRUD")
		}
	}
}

func TestNormalizePackage(t *testing.T) {
	if got := NormalizePackage("Products"); got != "products" {
		t.Errorf("NormalizePackage(Products) = %q, want products", got)
	}
	if got := NormalizePackage(" order_items "); got != "order_items" {
		t.Errorf("NormalizePackage( order_items ) = %q, want order_items", got)
	}
}

func TestStructName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"products", "Products"},
		{"order_items", "OrderItems"},
		{"user", "User"},
		{"status", "Status"},
		{"address", "Address"},
		{"process", "Process"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := StructName(tt.input); got != tt.want {
				t.Errorf("StructName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPluralStructName(t *testing.T) {
	if got := PluralStructName("products"); got != "Products" {
		t.Errorf("PluralStructName(products) = %q, want Products", got)
	}
}

func TestTableName(t *testing.T) {
	if got := TableName("OrderItems"); got != "orderitems" {
		t.Errorf("TableName(OrderItems) = %q, want orderitems", got)
	}
}
