package generator

const generatedRepositoryTests = `package {{.useCase}}
import (
 "errors"
 "fmt"
 "testing"
 "github.com/enviniom/nexokit/internal/modules/widgets/core"
 {{if eq .action "update"}}
 "context"
 entity "github.com/enviniom/nexokit/internal/modules/widgets/core/widgets"
 "gorm.io/driver/sqlite"
 {{if .tenant}}"github.com/enviniom/nexokit/internal/platform/tenant"{{end}}
 {{end}}
 "gorm.io/gorm"
)
func TestPersistenceErrorIdentity(t *testing.T) {
 for _,tt:=range []struct{name string;cause,target error}{
  {"absence",gorm.ErrRecordNotFound,core.ErrNotFound},
  {"duplicate",gorm.ErrDuplicatedKey,core.ErrConflict},
  {"reference",gorm.ErrForeignKeyViolated,core.ErrInvalidReference},
 } {
  t.Run(tt.name,func(t *testing.T){
   cause:=fmt.Errorf("database operation: %w",tt.cause);got:=persistenceError(cause)
   if !errors.Is(got,tt.target) || !errors.Is(got,tt.cause) {t.Fatalf("lost module identity/cause: %v",got)}
  })
 }
 unknown:=errors.New("private driver failure");if got:=persistenceError(unknown);got!=unknown {t.Fatalf("unexpected error changed: %v",got)}
}
{{if eq .action "update"}}
func TestUpdateReadFailureRollsBack(t *testing.T) {
 db,err:=gorm.Open(sqlite.Open(":memory:"),&gorm.Config{});if err!=nil {t.Fatal(err)}
 sqlDB,err:=db.DB();if err!=nil {t.Fatal(err)};sqlDB.SetMaxOpenConns(1);t.Cleanup(func(){sqlDB.Close()})
 if err:=db.AutoMigrate(&entity.Widgets{});err!=nil {t.Fatal(err)}
 m:=entity.Widgets{Name:"Original"{{if .tenant}},CompanyID:1{{end}}};if err:=db.Create(&m).Error;err!=nil {t.Fatal(err)}
 failure:=errors.New("private read failure")
 if err:=db.Callback().Query().Before("gorm:query").Register("fail_read",func(tx *gorm.DB){tx.AddError(failure)});err!=nil {t.Fatal(err)}
 _,err=NewRepository(db).Update(context.Background(),{{if .tenant}}tenant.NewScoped(1,"one"),{{end}}m.PublicID,"Changed")
 if !errors.Is(err,failure) {t.Fatalf("read failure lost: %v",err)}
 if err:=db.Callback().Query().Remove("fail_read");err!=nil {t.Fatal(err)}
 var fresh entity.Widgets;if err:=db.First(&fresh,m.ID).Error;err!=nil {t.Fatal(err)}
 if fresh.Name!="Original" {t.Fatalf("failed transaction persisted %q",fresh.Name)}
}
{{end}}
`
