package gen

const templateRouterContent = `
// ============================================================================
// This is auto-generated by gf cli tool only once. Fill this file as you wish.
// ============================================================================

package router

{TplPackageImports}
// Fill with you ideas below.

//加载路由
func init() {
	g1 := router.New("/api/v1", "/{TplPackageName}", auth.Auth)	
	g1.POST("/list","",{TplPackageName}.List)	
	g1.POST("/create","",{TplPackageName}.Create)
	g1.POST("/delete","",{TplPackageName}.Delete)
	g1.POST("/edit", "",{TplPackageName}.Edit)
	
}


`


