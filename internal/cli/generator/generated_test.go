package generator

import (
	"bytes"
	"context"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
	"time"
)

// Compile and exercise the emitted code, not a hand-maintained facsimile.
// Only the project's Go dependencies and focused platform source are copied;
// no app modules, configuration or external services enter the fixture.
func TestGeneratedModules(t *testing.T) {
	if testing.Short() {
		t.Skip("generated-code subprocess integration")
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Clean(filepath.Join(cwd, "../../.."))
	for _, tt := range []struct {
		name         string
		crud, tenant bool
	}{
		{"normal", false, false}, {"crud", true, false},
		{"tenant normal", false, true}, {"tenant crud", true, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, name := range []string{"go.mod", "go.sum"} {
				copyFixtureFile(t, root, dir, name)
			}
			for _, pkg := range []string{"shared", "platform/identity", "platform/apperror", "platform/messages", "platform/validator", "platform/response", "platform/query", "platform/gormutil", "platform/tenant"} {
				rel := filepath.Join("internal", pkg)
				entries, err := os.ReadDir(filepath.Join(root, rel))
				if err != nil {
					t.Fatal(err)
				}
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
						copyFixtureFile(t, root, dir, filepath.Join(rel, entry.Name()))
					}
				}
			}
			if err := os.Chdir(dir); err != nil {
				t.Fatal(err)
			}
			err = GenerateModule(ModuleOptions{Name: "widgets", CRUD: tt.crud, Tenant: tt.tenant})
			if chdirErr := os.Chdir(cwd); chdirErr != nil {
				t.Fatal(chdirErr)
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, path := range []string{"core/widgets/model.go", "core/widgets/dtos.go", "core/contracts.go", "core/errors.go", "container.go", "routes.go", "slices/list_widgets/handler.go", "slices/get_widgets/repository.go"} {
				if _, err := os.Stat(filepath.Join(dir, "internal/modules/widgets", path)); err != nil {
					t.Fatalf("new-code layout missing %s: %v", path, err)
				}
			}
			if !tt.crud {
				if _, err := os.Stat(filepath.Join(dir, "internal/modules/widgets/slices/create_widgets")); !os.IsNotExist(err) {
					t.Fatalf("normal scaffold unexpectedly includes create: %v", err)
				}
			}
			data := map[string]any{"crud": tt.crud, "tenant": tt.tenant}
			writeFixtureTest(t, dir, "internal/modules/widgets/core/widgets/dtos_test.go", generatedDTOTests, data)
			writeFixtureTest(t, dir, "internal/modules/widgets/repository_test.go", generatedBehaviorTests, data)
			actions := []string{"get", "list"}
			if tt.crud {
				actions = []string{"create", "get", "list", "update", "delete"}
			}
			for _, action := range actions {
				handlerData := map[string]any{"tenant": tt.tenant, "action": action, "useCase": action + "_widgets"}
				writeFixtureTest(t, dir, "internal/modules/widgets/slices/"+action+"_widgets/handler_test.go", generatedHandlerTests, handlerData)
				writeFixtureTest(t, dir, "internal/modules/widgets/slices/"+action+"_widgets/repository_test.go", generatedRepositoryTests, handlerData)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "go", "test", "-mod=readonly", "-count=1", "-timeout", "60s", "./internal/modules/widgets/...")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "GOWORK=off", "GOPROXY=off", "GOSUMDB=off")
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("generated %s tests: %v\n%s", tt.name, err, output)
			}
			t.Logf("generated %s compiled behavior:\n%s", tt.name, output)
		})
	}
}

func copyFixtureFile(t *testing.T, root, dir, rel string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func writeFixtureTest(t *testing.T, dir, rel, source string, data any) {
	t.Helper()
	tmpl, err := template.New(rel).Parse(source)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		t.Fatal(err)
	}
	formatted, err := format.Source(out.Bytes())
	if err != nil {
		t.Fatalf("format %s: %v\n%s", rel, err, out.String())
	}
	target := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, formatted, 0644); err != nil {
		t.Fatal(err)
	}
}

const generatedDTOTests = `package widgets
import (
 "encoding/json"
 "strings"
 "testing"
 "github.com/enviniom/nexokit/internal/shared"
 "github.com/oklog/ulid/v2"
)
func TestPublicIDHook(t *testing.T) {
 if (Widgets{}).TableName()!="widgets" {t.Fatal("wrong table name")}
 m:=Widgets{}; if err:=m.BeforeCreate(nil);err!=nil {t.Fatal(err)}
 if _,err:=ulid.ParseStrict(m.PublicID);err!=nil {t.Fatalf("not a ULID: %v",err)}
 id:=m.PublicID; if err:=m.BeforeCreate(nil);err!=nil || m.PublicID!=id {t.Fatalf("valid ID changed: %q %v",m.PublicID,err)}
 m.PublicID="not-a-public-id"; if err:=m.BeforeCreate(nil);err==nil {t.Fatal("invalid explicit ID accepted")}
}
func TestRequestsNormalizeBeforeRules(t *testing.T) {
 for _, tt := range []struct{name, value, normalized string; invalid bool}{
  {"trim", "  Alpha  ", "Alpha", false}, {"blank", "   ", "", true}, {"short", " x ", "x", true},
 } {
  t.Run(tt.name,func(t *testing.T){
   c := CreateWidgetsRequest{Name:tt.value}; u := UpdateWidgetsRequest{Name:tt.value}
   if got := c.Validate().HasErrors(); got != tt.invalid { t.Errorf("create invalid=%v",got) }
   if got := u.Validate().HasErrors(); got != tt.invalid { t.Errorf("update invalid=%v",got) }
   if c.Name != tt.normalized || u.Name != tt.normalized { t.Errorf("normalization lost: %q %q",c.Name,u.Name) }
  })
 }
}
func TestResponseOwnsSafeMapping(t *testing.T) {
 m := Widgets{BaseModel:shared.BaseModel{ID:123,PublicID:"public-id",CreatedBy:new(uint)},Name:"Alpha" {{if .tenant}},CompanyID:987{{end}}}
 var dto WidgetsResponse; dto.FromModel(m)
 b,err := json.Marshal(dto); if err != nil { t.Fatal(err) }
 var fields map[string]any; if err := json.Unmarshal(b,&fields); err != nil { t.Fatal(err) }
 if fields["id"] != "public-id" || fields["name"] != "Alpha" { t.Fatalf("mapping=%s",b) }
 for _, key := range []string{"company_id","created_by","updated_by","deleted_at","public_id"} {
  if _,ok := fields[key]; ok { t.Errorf("internal field %s exposed",key) }
 }
 if strings.Contains(string(b),"123") || strings.Contains(string(b),"987") { t.Fatalf("internal ID exposed: %s",b) }
}
`

const generatedBehaviorTests = `package widgets_test
import (
 "context"
 "encoding/json"
 "errors"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"

 module "github.com/enviniom/nexokit/internal/modules/widgets"
 "github.com/enviniom/nexokit/internal/modules/widgets/core"
 entity "github.com/enviniom/nexokit/internal/modules/widgets/core/widgets"
 get "github.com/enviniom/nexokit/internal/modules/widgets/slices/get_widgets"
 list "github.com/enviniom/nexokit/internal/modules/widgets/slices/list_widgets"
 {{if .crud}}
 create "github.com/enviniom/nexokit/internal/modules/widgets/slices/create_widgets"
 {{if .tenant}}
 update "github.com/enviniom/nexokit/internal/modules/widgets/slices/update_widgets"
 del "github.com/enviniom/nexokit/internal/modules/widgets/slices/delete_widgets"
 {{end}}
 {{end}}
 "github.com/enviniom/nexokit/internal/platform/query"
 {{if .tenant}}"github.com/enviniom/nexokit/internal/platform/tenant"{{end}}
 "github.com/gin-gonic/gin"
 "github.com/oklog/ulid/v2"
 "gorm.io/driver/sqlite"
 "gorm.io/gorm"
)
func database(t *testing.T) *gorm.DB {
 t.Helper()
 db,err := gorm.Open(sqlite.Open(":memory:"),&gorm.Config{TranslateError:true}); if err != nil {t.Fatal(err)}
 sqlDB,err := db.DB(); if err != nil {t.Fatal(err)}; sqlDB.SetMaxOpenConns(1); t.Cleanup(func(){sqlDB.Close()})
 if err := db.AutoMigrate(&entity.Widgets{}); err != nil {t.Fatal(err)}
 return db
}
func request(r http.Handler, method,path,body string) *httptest.ResponseRecorder {
 w:=httptest.NewRecorder(); req:=httptest.NewRequest(method,path,strings.NewReader(body)); req.Header.Set("Content-Type","application/json"); r.ServeHTTP(w,req); return w
}
func router(db *gorm.DB{{if .tenant}}, tc *tenant.TenantContext{{end}}) *gin.Engine {
 gin.SetMode(gin.TestMode); r:=gin.New()
 {{if .tenant}}if tc != nil {r.Use(func(c *gin.Context){tenant.SetGin(c,*tc)})}{{end}}
 module.Register(r.Group("/v1"),module.NewContainer(db)); return r
}
func seed(t *testing.T,db *gorm.DB,name string{{if .tenant}}, company uint{{end}}) entity.Widgets {
 t.Helper(); m:=entity.Widgets{Name:name{{if .tenant}},CompanyID:company{{end}}}
 if err:=db.Create(&m).Error; err != nil {t.Fatal(err)}
 if _,err:=ulid.ParseStrict(m.PublicID); err != nil {t.Fatalf("public ID not ULID: %q: %v",m.PublicID,err)}
 return m
}
func TestListSortingCountAndSoftDelete(t *testing.T) {
 db:=database(t); ctx:=context.Background()
 z:=seed(t,db,"Zulu"{{if .tenant}},1{{end}}); seed(t,db,"Alpha"{{if .tenant}},1{{end}})
 deleted:=seed(t,db,"Gone"{{if .tenant}},1{{end}}); if err:=db.Delete(&deleted).Error; err != nil {t.Fatal(err)}
 {{if .tenant}}seed(t,db,"Foreign",2); tc:=tenant.NewScoped(1,"one"){{end}}
 repo:=list.NewRepository(db)
 params:=query.ListParams{Pagination:query.PaginationParams{Page:1,PerPage:1},Sort:query.SortParams{Sort:"name",Order:"asc"}}
 items,total,err:=repo.List(ctx,{{if .tenant}}tc,{{end}}params)
 if err!=nil || total!=2 || len(items)!=1 || items[0].Name!="Alpha" {t.Fatalf("list=%+v total=%d err=%v",items,total,err)}
 for _, sort:=range []string{"company_id","unknown_column","name; DROP TABLE widgets;--"} {
  params.Sort.Sort=sort; params.Pagination.PerPage=100
  items,total,err=repo.List(ctx,{{if .tenant}}tc,{{end}}params)
  if err!=nil || total!=2 || len(items)!=2 {t.Fatalf("guarded sort %q: %v %d %v",sort,items,total,err)}
 }
 params.Search.Query="Alpha"; params.Sort.Sort="name"
 items,total,err=repo.List(ctx,{{if .tenant}}tc,{{end}}params)
 if err!=nil || total!=1 || len(items)!=1 {t.Fatalf("filtered count=%d items=%d err=%v",total,len(items),err)}
 _,err=get.NewRepository(db).Get(ctx,{{if .tenant}}tc,{{end}}deleted.PublicID)
 if !errors.Is(err,core.ErrNotFound) {t.Fatalf("deleted lookup=%v",err)}
 w:=request(router(db{{if .tenant}},&tc{{end}}),"GET","/v1/widgets/"+z.PublicID,"")
 if w.Code!=200 || strings.Contains(w.Body.String(),"company_id") {t.Fatalf("get response=%d %s",w.Code,w.Body)}
 w=request(router(db{{if .tenant}},&tc{{end}}),"GET","/v1/widgets?sort=name&order=asc&per_page=1","")
 var envelope struct{Data []entity.WidgetsResponse; Meta struct{Pagination struct{Total int64}}}
 if err:=json.Unmarshal(w.Body.Bytes(),&envelope); err!=nil {t.Fatal(err)}
 if w.Code!=200 || len(envelope.Data)!=1 || envelope.Data[0].Name!="Alpha" || envelope.Meta.Pagination.Total!=2 {t.Fatalf("list response=%d %s",w.Code,w.Body)}
}
func TestUnexpectedErrorsStayInternal(t *testing.T) {
 db:=database(t)
 // Missing table supplies a real technical failure; GET must not disguise it as 404.
 if err:=db.Migrator().DropTable(&entity.Widgets{}); err!=nil {t.Fatal(err)}
 {{if .tenant}}tc:=tenant.NewScoped(1,"one"){{end}}
 r:=router(db{{if .tenant}},&tc{{end}})
 for _, path:=range []string{"/v1/widgets","/v1/widgets/unknown"} {
  w:=request(r,"GET",path,"")
  if w.Code!=500 || strings.Contains(w.Body.String(),"no such table") || strings.Contains(w.Body.String(),"SELECT") || strings.Contains(w.Body.String(),"debug") {t.Fatalf("unsafe error=%d %s",w.Code,w.Body)}
 }
}
{{if .crud}}
func TestCRUDHTTPAndKnownFailures(t *testing.T) {
 db:=database(t); ctx:=context.Background()
 {{if .tenant}}tc:=tenant.NewScoped(1,"one"){{end}}
 r:=router(db{{if .tenant}},&tc{{end}})
 for _, tt:=range []struct{body string; status int}{ {"{",400},{"{\"name\":42}",400},{"{\"name\":\"   \"}",422}} {
  w:=request(r,"POST","/v1/widgets",tt.body); if w.Code!=tt.status {t.Fatalf("invalid request=%d %s",w.Code,w.Body)}
 }
 var count int64; db.Model(&entity.Widgets{}).Count(&count); if count!=0 {t.Fatal("invalid requests persisted")}
 w:=request(r,"POST","/v1/widgets","{\"name\":\"  Alpha  \",\"company_id\":999}")
 var envelope struct{Data entity.WidgetsResponse}; if err:=json.Unmarshal(w.Body.Bytes(),&envelope);err!=nil {t.Fatal(err)}
 if w.Code!=201 || envelope.Data.Name!="Alpha" || strings.Contains(w.Body.String(),"company_id") {t.Fatalf("create=%d %s",w.Code,w.Body)}
 id:=envelope.Data.ID; if _,err:=ulid.ParseStrict(id);err!=nil {t.Fatal(err)}
 var persisted entity.Widgets; if err:=db.First(&persisted,"public_id = ?",id).Error;err!=nil {t.Fatal(err)}
 {{if .tenant}}if persisted.CompanyID!=1 {t.Fatalf("JSON tenant trusted: %d",persisted.CompanyID)}{{end}}
 // A duplicate public ID must preserve module identity and the GORM cause.
 duplicate:=entity.Widgets{Name:"duplicate",BaseModel:persisted.BaseModel{{if .tenant}},CompanyID:1{{end}}}; duplicate.ID=0
 _,err:=create.NewRepository(db).Create(ctx,{{if .tenant}}tc,{{end}}&duplicate)
 if !errors.Is(err,core.ErrConflict) || !errors.Is(err,gorm.ErrDuplicatedKey) {t.Fatalf("duplicate mapping=%v",err)}
 w=request(r,"PUT","/v1/widgets/"+id,"{\"name\":\"  Beta  \",\"company_id\":999}")
 if w.Code!=200 || !strings.Contains(w.Body.String(),"Beta") {t.Fatalf("update=%d %s",w.Code,w.Body)}
 {{if .tenant}}db.First(&persisted,"public_id = ?",id); if persisted.CompanyID!=1 {t.Fatal("update reassigned tenant")}{{end}}
 w=request(r,"DELETE","/v1/widgets/"+id,""); if w.Code!=204 || w.Body.Len()!=0 {t.Fatalf("delete=%d %s",w.Code,w.Body)}
 w=request(r,"GET","/v1/widgets/"+id,""); if w.Code!=404 {t.Fatalf("deleted GET=%d %s",w.Code,w.Body)}
 w=request(r,"DELETE","/v1/widgets/"+id,""); if w.Code!=404 {t.Fatalf("deleted DELETE=%d",w.Code)}
 w=request(r,"PUT","/v1/widgets/"+id,"{\"name\":\"Changed\"}"); if w.Code!=404 {t.Fatalf("deleted UPDATE=%d",w.Code)}
 w=request(r,"PUT","/v1/widgets/missing","{\"name\":\"Valid\"}"); if w.Code!=404 {t.Fatalf("missing update=%d",w.Code)}
}
{{end}}
{{if .tenant}}
func TestTenantIsolationAndContext(t *testing.T) {
 db:=database(t); ctx:=context.Background()
 own:=seed(t,db,"Own",1); foreign:=seed(t,db,"Foreign",2)
 scoped:=tenant.NewScoped(1,"one"); root:=tenant.NewRoot(); invalid:=tenant.TenantContext{}
 inconsistent:=tenant.TenantContext{IsRootScope:true,CompanyID:1}
 params:=query.ListParams{Pagination:query.PaginationParams{Page:1,PerPage:20}}
 for _, tc:=range []tenant.TenantContext{invalid,inconsistent} {
  _,_,err:=list.NewRepository(db).List(ctx,tc,params); if !errors.Is(err,core.ErrTenantRequired) {t.Fatalf("invalid scope accepted: %v",err)}
  _,err=get.NewRepository(db).Get(ctx,tc,own.PublicID); if !errors.Is(err,core.ErrTenantRequired) {t.Fatalf("invalid get=%v",err)}
 }
 _,err:=get.NewRepository(db).Get(ctx,scoped,foreign.PublicID); if !errors.Is(err,core.ErrNotFound) {t.Fatalf("cross-tenant get=%v",err)}
 _,err=get.NewRepository(db).Get(ctx,root,foreign.PublicID); if !errors.Is(err,core.ErrTenantRequired) {t.Fatalf("global get=%v",err)}
 items,total,err:=list.NewRepository(db).List(ctx,root,params); if err!=nil || total!=2 || len(items)!=2 {t.Fatalf("explicit global list=%v %d %v",items,total,err)}
 for _, tc:=range []*tenant.TenantContext{nil,&invalid,&inconsistent} {
  r:=router(db,tc)
  paths:=[]struct{method,path,body string}{ {"GET","/v1/widgets",""},{"GET","/v1/widgets/"+own.PublicID,""}{{if .crud}},{"POST","/v1/widgets","{\"name\":\"Valid\"}"},{"PUT","/v1/widgets/"+own.PublicID,"{\"name\":\"Valid\"}"},{"DELETE","/v1/widgets/"+own.PublicID,""}{{end}}}
  for _, p:=range paths {w:=request(r,p.method,p.path,p.body); if w.Code!=400 {t.Fatalf("missing/invalid context %s=%d %s",p.method,w.Code,w.Body)}}
 }
 w:=request(router(db,&root),"GET","/v1/widgets",""); if w.Code!=200 {t.Fatalf("global list=%d",w.Code)}
 w=request(router(db,&scoped),"GET","/v1/widgets/"+foreign.PublicID,""); if w.Code!=404 {t.Fatalf("foreign get=%d",w.Code)}
 {{if .crud}}
 _,err=create.NewRepository(db).Create(ctx,root,&entity.Widgets{Name:"No company"}); if !errors.Is(err,core.ErrTenantRequired) {t.Fatalf("global create=%v",err)}
 for _,tc:=range []tenant.TenantContext{invalid,inconsistent,root} {
  _,err=update.NewRepository(db).Update(ctx,tc,own.PublicID,"Changed"); if !errors.Is(err,core.ErrTenantRequired) {t.Fatalf("invalid update=%v",err)}
  err=del.NewRepository(db).Delete(ctx,tc,own.PublicID); if !errors.Is(err,core.ErrTenantRequired) {t.Fatalf("invalid delete=%v",err)}
 }
 _,err=update.NewRepository(db).Update(ctx,scoped,foreign.PublicID,"Changed"); if !errors.Is(err,core.ErrNotFound) {t.Fatalf("foreign update=%v",err)}
 err=del.NewRepository(db).Delete(ctx,scoped,foreign.PublicID); if !errors.Is(err,core.ErrNotFound) {t.Fatalf("foreign delete=%v",err)}
 w=request(router(db,&root),"POST","/v1/widgets","{\"name\":\"Valid\",\"company_id\":1}"); if w.Code!=400 {t.Fatalf("global create HTTP=%d",w.Code)}
 for _,p:=range []struct{method,body string}{ {"PUT","{\"name\":\"Changed\"}"},{"DELETE",""}} {
  w=request(router(db,&scoped),p.method,"/v1/widgets/"+foreign.PublicID,p.body); if w.Code!=404 {t.Fatalf("foreign %s=%d",p.method,w.Code)}
 }
 var fresh entity.Widgets; db.First(&fresh,foreign.ID); if fresh.Name!="Foreign" || fresh.DeletedAt.Valid {t.Fatal("foreign row mutated")}
 // Repository create overwrites caller-supplied company; it cannot assign another tenant.
 m:=entity.Widgets{Name:"Assigned",CompanyID:2}
 if _,err:=create.NewRepository(db).Create(ctx,scoped,&m);err!=nil {t.Fatal(err)}
 if m.CompanyID!=1 {t.Fatal("repository trusted caller company")}
 {{end}}
}
{{end}}
`
