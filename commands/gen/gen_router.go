package gen

import (
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/gogf/gf-cli/library/allyes"
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/text/gstr"
	_ "github.com/lib/pq"

	"strings"
)

const (
	DEFAULT_GEN_ROUTER_PATH = "./app/router"
)

// doGenModel implements the "gen model" command.
func doGenRouter(parser *gcmd.Parser) {
	var err error
	genPath := parser.GetArg(3, DEFAULT_GEN_ROUTER_PATH)
	if !gfile.IsEmpty(genPath) && !allyes.Check() {
		s := gcmd.Scanf("path '%s' is not empty, files might be overwrote, continue? [y/n]: ", genPath)
		if strings.EqualFold(s, "n") {
			return
		}
	}
	tableOpt := parser.GetOpt("table")
	linkInfo := parser.GetOpt("link")
	configFile := parser.GetOpt("config")
	configGroup := parser.GetOpt("group", gdb.DEFAULT_GROUP_NAME)
	prefixArray := gstr.SplitAndTrim(parser.GetOpt("prefix"), ",")

	if linkInfo != "" {
		path := gfile.TempDir() + gfile.Separator + "config.toml"
		if err := gfile.PutContents(path, fmt.Sprintf("[database]\n\tlink=\"%s\"", linkInfo)); err != nil {
			mlog.Fatalf("write configuration file to '%s' failed: %v", path, err)
		}
		defer gfile.Remove(path)
		if err := g.Cfg().SetPath(gfile.TempDir()); err != nil {
			mlog.Fatalf("set configuration path '%s' failed: %v", gfile.TempDir(), err)
		}
	}
	// Custom configuration file.
	if configFile != "" {
		path, err := gfile.Search(configFile)
		if err != nil {
			mlog.Fatalf("search configuration file '%s' failed: %v", configFile, err)
		}
		if err := g.Cfg().SetPath(gfile.Dir(path)); err != nil {
			mlog.Fatalf("set configuration path '%s' failed: %v", path, err)
		}
		g.Cfg().SetFileName(gfile.Basename(path))
	}

	db := g.DB(configGroup)
	if db == nil {
		mlog.Fatal("database initialization failed")
	}

	if err := gfile.Mkdir(genPath); err != nil {
		mlog.Fatalf("mkdir for generating path '%s' failed: %v", genPath, err)
	}

	tables := ([]string)(nil)
	if tableOpt != "" {
		tables = gstr.SplitAndTrim(tableOpt, ",")
	} else {
		tables, err = db.Tables()
		if err != nil {
			mlog.Fatalf("fetching tables failed: \n %v", err)
		}
	}

	for _, table := range tables {
		variable := table
		for _, v := range prefixArray {
			variable = gstr.TrimLeftStr(variable, v)
		}
		generateRouterContentFile(db, table, variable, genPath, configGroup)
	}
	mlog.Print("done!")
}

// generateModelContentFile generates the model content of given table.
// The parameter <variable> specifies the variable name for the table, which
// is the prefix-stripped name of the table.
//
// Note that, this function will generate 3 files under <folderPath>/<packageName>/:
// file.go        : the package index go file, developer can fill the file with model logic;
// file_entity.go : the entity definition go file, it can be overwrote by gf-cli tool, don't edit it;
// file_model.go  : the active record design model definition go file, it can be overwrote by gf-cli tool, don't edit it;
func generateRouterContentFile(db gdb.DB, table, variable, folderPath, groupName string) {
	fieldMap, err := db.TableFields(table)
	if err != nil {
		mlog.Fatalf("fetching tables fields failed for table '%s':\n%v", table, err)
	}
	camelName := gstr.CamelCase(variable)
	structDefine := generateStructDefinition(fieldMap)
	packageName := gstr.SnakeCase(variable)
	fileName := gstr.Trim(packageName, "-_.")
	if len(fileName) > 5 && fileName[len(fileName)-5:] == "_test" {
		// Add suffix to avoid the table name which contains "_test",
		// which would make the go file a testing file.
		fileName += "_table"
	}


	// index
	path := gfile.Join(folderPath, fileName+".go")
	if !gfile.Exists(path) {


		modelPackageImports := "import ( \n" +
			`	"suyuan/app/controller/`+packageName+`"` + "\n"+
			`	"suyuan/app/utils/middleware/auth"` + "\n" +
			`	"suyuan/app/utils/middleware/router"` + "\n" +
			"\n)"



		//空白model
		indexContent := gstr.ReplaceByMap(templateRouterContent, g.MapStrStr{
			"{TplTableName}":      table,
			"{TplModelName}":      camelName,
			"{TplGroupName}":      groupName,
			"{TplPackageName}":    packageName,
			"{TplPackageImports}": modelPackageImports,
			"{TplStructDefine}":   structDefine,
		})


		if err := gfile.PutContents(path, strings.TrimSpace(indexContent)); err != nil {
			mlog.Fatalf("writing content to '%s' failed: %v", path, err)
		} else {
			mlog.Print("generated:", path)
		}

	}

}

