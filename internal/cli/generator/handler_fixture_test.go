package generator

const generatedHandlerTests = `package {{.useCase}}
import (
 "context"
 "errors"
 "fmt"
 "net/http/httptest"
 "strings"
 "testing"
 "github.com/enviniom/nexokit/internal/modules/widgets/core"
 {{if ne .action "delete"}}entity "github.com/enviniom/nexokit/internal/modules/widgets/core/widgets"{{end}}
 {{if eq .action "list"}}"github.com/enviniom/nexokit/internal/platform/query"{{end}}
 {{if .tenant}}"github.com/enviniom/nexokit/internal/platform/tenant"{{end}}
 "github.com/gin-gonic/gin"
)
type fakeRepository struct {calls int; err error; name string}
{{if eq .action "create"}}
func (f *fakeRepository) Create(_ context.Context, {{if .tenant}}_ tenant.TenantContext,{{end}}m *entity.Widgets)(*entity.Widgets,error) {f.calls++;f.name=m.Name;return m,f.err}
{{else if eq .action "get"}}
func (f *fakeRepository) Get(_ context.Context,{{if .tenant}}_ tenant.TenantContext,{{end}}_ string)(*entity.Widgets,error) {f.calls++;return &entity.Widgets{Name:"Alpha"},f.err}
{{else if eq .action "list"}}
func (f *fakeRepository) List(_ context.Context,{{if .tenant}}_ tenant.TenantContext,{{end}}p query.ListParams)([]entity.Widgets,int64,error) {f.calls++;if p.Pagination.Page!=1 || p.Pagination.PerPage!=100 {return nil,0,fmt.Errorf("pagination not normalized: %+v",p)};return []entity.Widgets{ {Name:"Alpha"} },1,f.err}
{{else if eq .action "update"}}
func (f *fakeRepository) Update(_ context.Context,{{if .tenant}}_ tenant.TenantContext,{{end}}_,name string)(*entity.Widgets,error) {f.calls++;f.name=name;return &entity.Widgets{Name:name},f.err}
{{else if eq .action "delete"}}
func (f *fakeRepository) Delete(_ context.Context,{{if .tenant}}_ tenant.TenantContext,{{end}}_ string)error {f.calls++;return f.err}
{{end}}
func execute(f *fakeRepository, body string{{if .tenant}},tc *tenant.TenantContext{{end}})*httptest.ResponseRecorder {
 gin.SetMode(gin.TestMode);r:=gin.New()
 {{if .tenant}}if tc!=nil {r.Use(func(c *gin.Context){tenant.SetGin(c,*tc)})}{{end}}
 h:=NewHandler(f);r.POST("/:id",h.Handle)
 w:=httptest.NewRecorder();req:=httptest.NewRequest("POST","/public-id?page=-1&per_page=999",strings.NewReader(body));req.Header.Set("Content-Type","application/json");r.ServeHTTP(w,req);return w
}
func TestHandlerErrors(t *testing.T) {
 {{if .tenant}}tc:=tenant.NewScoped(1,"one"){{end}}
 for _,tt:=range []struct{name string;err error;status int}{
  {"unexpected",errors.New("private database diagnostic"),500},
  {"known",fmt.Errorf("wrapped: %w",core.ErrNotFound),404},
 } {
  t.Run(tt.name,func(t *testing.T){f:=&fakeRepository{err:tt.err};w:=execute(f,"{\"name\":\"  Alpha  \"}"{{if .tenant}},&tc{{end}})
   if f.calls!=1 || w.Code!=tt.status {t.Fatalf("calls=%d status=%d body=%s",f.calls,w.Code,w.Body)}
   if strings.Contains(w.Body.String(),"private database diagnostic") || strings.Contains(w.Body.String(),"debug") {t.Fatalf("unsafe response: %s",w.Body)}
  })
 }
}
{{if or (eq .action "create") (eq .action "update")}}
func TestBindingValidationAndPersistedNormalization(t *testing.T) {
 {{if .tenant}}tc:=tenant.NewScoped(1,"one"){{end}}
 for _,tt:=range []struct{name,body string;status int}{
  {"malformed","{",400},{"wrong type","{\"name\":42}",400},{"whitespace","{\"name\":\"  \"}",422},
 } {
  t.Run(tt.name,func(t *testing.T){f:=&fakeRepository{};w:=execute(f,tt.body{{if .tenant}},&tc{{end}})
   if f.calls!=0 || w.Code!=tt.status {t.Fatalf("calls=%d status=%d body=%s",f.calls,w.Code,w.Body)}
  })
 }
 f:=&fakeRepository{};w:=execute(f,"{\"name\":\"  Alpha  \",\"company_id\":999}"{{if .tenant}},&tc{{end}})
 if f.calls!=1 || f.name!="Alpha" || w.Code!={{if eq .action "create"}}201{{else}}200{{end}} {t.Fatalf("normalization calls=%d name=%q response=%d %s",f.calls,f.name,w.Code,w.Body)}
 if strings.Contains(w.Body.String(),"company_id") {t.Fatal("company key exposed")}
}
{{end}}
{{if .tenant}}
func TestTenantContextShortCircuits(t *testing.T) {
 invalid:=tenant.TenantContext{};inconsistent:=tenant.TenantContext{IsRootScope:true,CompanyID:1};root:=tenant.NewRoot()
 for _,tc:=range []*tenant.TenantContext{nil,&invalid,&inconsistent{{if ne .action "list"}},&root{{end}}} {
  f:=&fakeRepository{};w:=execute(f,"{\"name\":\"Valid\"}",tc)
  if f.calls!=0 || w.Code!=400 {t.Fatalf("missing/invalid scope calls=%d response=%d %s",f.calls,w.Code,w.Body)}
 }
 {{if eq .action "list"}}{f:=&fakeRepository{};w:=execute(f,"",&root);if f.calls!=1 || w.Code!=200 {t.Fatalf("explicit global list calls=%d status=%d",f.calls,w.Code)}}{{else}}_ = root{{end}}
 scoped:=tenant.NewScoped(1,"one");f:=&fakeRepository{};w:=execute(f,"{\"name\":\"Valid\"}",&scoped)
 if f.calls!=1 || w.Code>=400 {t.Fatalf("selected scope calls=%d response=%d %s",f.calls,w.Code,w.Body)}
}
{{end}}
`
