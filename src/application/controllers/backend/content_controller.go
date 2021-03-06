package backend

import (
	"encoding/json"
	"fmt"
	"html/template"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-xorm/xorm"
	"github.com/microcosm-cc/bluemonday"
	"github.com/xiusin/pine"
	"github.com/xiusin/pinecms/src/application/models/tables"

	"github.com/golang/glog"
	"github.com/xiusin/pinecms/src/application/controllers"

	"github.com/xiusin/pinecms/src/application/models"
	"github.com/xiusin/pinecms/src/common/helper"
)

type ContentController struct {
	pine.Controller
}

func (c *ContentController) RegisterRoute(b pine.IRouterWrapper) {
	b.GET("/content/aside-category", "AsideCategory")
	b.ANY("/content/news-list", "NewsList")
	b.ANY("/content/news-crud", "NewsModelJson")

	b.ANY("/content/page", "Page")
	b.ANY("/content/add", "AddContent")
	b.ANY("/content/edit", "EditContent")
	b.ANY("/content/delete", "DeleteContent")
	b.ANY("/content/order", "OrderContent")
}

func (c *ContentController) AsideCategory() {
	cats := models.NewCategoryModel().GetContentRightCategoryTree(models.NewCategoryModel().GetAll(false), 0)
	helper.Ajax(cats, 0, c.Ctx())
}

func (c *ContentController) NewsList(orm *xorm.Engine) {
	catid, _ := c.Ctx().GetInt64("catid")
	page, _ := c.Ctx().GetInt64("page")
	rows, _ := c.Ctx().GetInt64("perPage")
	catogoryModel := models.NewCategoryModel().GetCategory(catid)
	if catogoryModel == nil {
		helper.Ajax("分类不存在", 1, c.Ctx())
		return
	}
	if catogoryModel.ModelId < 1 {
		helper.Ajax("找不到关联模型", 1, c.Ctx())
		return
	}
	relationDocumentModel := models.NewDocumentModel().GetByID(catogoryModel.ModelId)
	if relationDocumentModel.Id == 0 {
		helper.Ajax("找不到关联模型", 1, c.Ctx())
		return
	}

	// 获取所有字段
	dslFields := models.NewDocumentFieldDslModel().GetList(catogoryModel.ModelId)
	var tMapF = map[string]string{}
	var ff []string
	for _, dsl := range dslFields {
		tMapF[dsl.TableField] = dsl.FormName
		ff = append(ff, dsl.TableField)
	}

	querySqlWhere := []string{"catid=?", "deleted_time IS NULL"}
	var whereHolder = []interface{}{catid}
	getData := c.Ctx().GetData()

	for param, values := range getData {
		if (!strings.HasPrefix(param, "search_")) || len(values) == 0 || len(values[0]) == 0 {
			continue
		}
		field := strings.TrimLeft(param, "search_")
		querySqlWhere = append(querySqlWhere, field+" LIKE ?")
		whereHolder = append(whereHolder, "%"+values[0]+"%")
	}

	offset := (page - 1) * rows
	querySql := "SELECT * FROM `%s` WHERE " + strings.Join(querySqlWhere, " AND ") + " ORDER BY listorder DESC, id DESC LIMIT %d,%d"
	sql := []interface{}{fmt.Sprintf(querySql, controllers.GetTableName(relationDocumentModel.Table), offset, rows)}
	sql = append(sql, whereHolder...)

	contents, err := orm.QueryString(sql...)
	if err != nil {
		pine.Logger().Error("请求列表错误", err)
		helper.Ajax("获取文档列表错误", 1, c.Ctx())
		return
	}

	countSql := "SELECT COUNT(*) total FROM `%s` WHERE " + strings.Join(querySqlWhere, " AND ")
	sql = []interface{}{fmt.Sprintf(countSql, controllers.GetTableName(relationDocumentModel.Table))}
	sql = append(sql, whereHolder)

	totals, _ := orm.QueryString(sql...)
	var total = "0"
	if len(totals) > 0 {
		total = totals[0]["total"]
	}
	if contents == nil {
		contents = []map[string]string{}
	}
	helper.Ajax(pine.H{"rows": contents, "total": total}, 0, c.Ctx())
}

// NewsModelJson 动态json表单
func (c *ContentController) NewsModelJson(orm *xorm.Engine) {
	catid, _ := c.Ctx().GetInt64("catid")
	catogoryModel := models.NewCategoryModel().GetCategory(catid)
	if catogoryModel == nil {
		helper.Ajax("分类不存在", 1, c.Ctx())
		return
	}
	// 根据类型展示不同的页面
	if catogoryModel.Type == 2 { // 外部链接

	} else if catogoryModel.Type == 1 { // 单页发布

	}
	rd := models.NewDocumentModel().GetByID(catogoryModel.ModelId)
	if rd == nil || rd.Id == 0 {
		helper.Ajax("找不到关联模型", 1, c.Ctx())
		return
	}
	var fields []tables.DocumentModelDsl
	orm.Table(new(tables.DocumentModelDsl)).Where("mid = ?", catogoryModel.ModelId).OrderBy("listorder").Find(&fields) // 按排序查字段
	var forms []FormControl
	var formColums []FormControl
	fm := models.NewDocumentModelFieldModel().GetMap()
	for _, field := range fields {
		form := FormControl{Type: fm[field.FieldType].AmisType, Name: field.TableField, Label: field.FormName}
		if field.ShowInList {
			formColums = append(formColums, FormControl{Type: "text", Name: field.TableField, Label: field.FormName})
		}
		rf := reflect.ValueOf(&form)
		if fm[field.FieldType].Opt != "" { // 判断是否需要合并属性
			optArr := strings.Split(fm[field.FieldType].Opt, "\r\n")
			for _, v := range optArr {
				opts := strings.SplitN(v, ":", 3)
				opts[0] = ucwords(opts[0])
				switch opts[1] {
				case "bool":
					v, _ := strconv.ParseBool(opts[2])
					rf.Elem().FieldByName(opts[0]).SetBool(v)
				case "int":
					v, _ := strconv.Atoi(opts[2])
					rf.Elem().FieldByName(opts[0]).SetInt(int64(v))
				case "array", "object":
					var val []KV
					json.Unmarshal([]byte(opts[2]), &val)
					switch opts[0] {
					case "Options":
						form.Options = val
						if form.Type == "checkboxes" {
							form.Multiple = true
						}
					}
				default:
				}
			}
		}
		if field.Required == 1 {
			form.Required = true
			form.ValidationErrors = field.RequiredTips
		}
		if field.Validator != "" {
			form.Validations = field.Validator
		}
		if field.Default != "" {
			form.Value = field.Default
		}
		forms = append(forms, form)
	}

	action := map[string]interface{}{
		"type":       "button",
		"align":      "right",
		"actionType": "drawer",
		"label":      "添加",
		"icon":       "fa fa-plus pull-left",
		"size":       "sm",
		"primary":    true,
		"drawer": map[string]interface{}{
			"position": "right",
			"size":     "xl",
			"title":    "发布内容",
			"body": map[string]interface{}{
				"type":     "form",
				"mode":     "horizontal",
				"api":      fmt.Sprintf("POST content/add?mid=%d&catid=%d&table_name=%s", rd.Id, catogoryModel.Catid, rd.Table),
				"controls": forms,
			},
		},
	}

	formColums = append(formColums, FormControl{
		Type:        "operation",
		Label:       "操作",
		LimitsLogic: "or",
		Limits:      []string{"edit", "del"},
		Buttons: []interface{}{
			map[string]interface{}{
				"type":       "action",
				"limits":     "edit",
				"actionType": "drawer",
				"tooltip":    "修改",
				"icon":       "fa fa-edit text-info",
				"drawer": map[string]interface{}{
					"position": "right",
					"size":     "xl",
					"title":    "修改内容",
					"body": map[string]interface{}{
						"type":     "form",
						"mode":     "horizontal",
						"api":      fmt.Sprintf("POST content/edit?id=$id&mid=%d&catid=%d&table_name=%s", rd.Id, catogoryModel.Catid, rd.Table),
						"controls": forms,
					},
				},
			},
			map[string]interface{}{
				"limits":      "del",
				"type":        "action",
				"icon":        "fa fa-times text-danger",
				"actionType":  "ajax",
				"api":         fmt.Sprintf("POST content/delete?id=$id&catid=%d", catogoryModel.Catid),
				"tooltip":     "删除",
				"confirmText": "您确认要删除?",
			},
		},
	})

	helper.Ajax(pine.H{
		"type":            "crud",
		"columns":         formColums,
		"filterTogglable": true,
		"filter":          "$preset.forms.filter",
		"api":             fmt.Sprintf("GET content/news-list?catid=%d", catogoryModel.Catid),
		"headerToolbar": []interface{}{
			"filter-toggler",
			map[string]string{
				"type":  "columns-toggler",
				"align": "left",
			},
			map[string]string{
				"type":  "pagination",
				"align": "left",
			},
			action,
		},
		"footerToolbar": []string{"statistics", "switch-per-page", "pagination"},
	}, 0, c.Ctx())
}

func (c *ContentController) Page() {
	catid, _ := c.Ctx().GetInt64("catid")
	if catid == 0 {
		helper.Ajax("页面错误", 1, c.Ctx())
		return
	}
	pageModel := models.NewPageModel()
	page := pageModel.GetPage(catid)
	hasPage := page != nil
	if page == nil {
		page = &tables.Page{}
	}
	var res bool
	if c.Ctx().IsPost() {
		page.Title = c.Ctx().FormValue("title")
		page.Content = c.Ctx().FormValue("content")
		page.Keywords = c.Ctx().FormValue("keywords")
		page.Description = c.Ctx().FormValue("description")
		page.Updatetime = int64(helper.GetTimeStamp())
		if !hasPage {
			page.Catid = catid
			res = pageModel.AddPage(page)
		} else {
			res = pageModel.UpdatePage(page)
		}
		if res {
			helper.Ajax("发布成功", 0, c.Ctx())
		} else {
			helper.Ajax("发布失败", 1, c.Ctx())
		}
		return
	}

	c.Ctx().Render().ViewData("catid", catid)
	c.Ctx().Render().ViewData("info", page)
	c.Ctx().Render().HTML("backend/content_page.html")

}

type customForm map[string]string

func (c customForm) MustCheck() bool {
	var ok bool
	if _, ok = c["catid"]; !ok {
		return false
	}
	if _, ok = c["mid"]; !ok {
		return false
	}
	if _, ok = c["table_name"]; !ok {
		return false
	}
	return true
}

//AddContent 添加内容
func (c *ContentController) AddContent(orm *xorm.Engine) {
	if c.Ctx().IsPost() {
		mid, _ := strconv.Atoi(c.Ctx().FormValue("mid"))
		if mid < 1 {
			helper.Ajax("模型参数错误， 无法确定所属模型", 1, c.Ctx())
			return
		}
		var data = customForm{}
		postData := c.Ctx().PostData()
		model := models.NewDocumentModel().GetByID(int64(mid))
		//table := orm.TableInfo(controllers.GetTableName(model.Table))
		for formName, values := range postData {
			if formName == "attrs" {
				data[formName] = strings.Join(values, ",")
			} else {
				data[formName] = values[0]
			}
			//col := table.GetColumn(formName)
			//if data[formName] == "" && col.Nullable{
			//	if col.SQLType.IsNumeric() {
			//		data[formName] = "0"
			//	}
			//}
		}
		data["mid"] = c.Ctx().FormValue("mid")
		data["catid"] = c.Ctx().FormValue("catid")
		data["table_name"] = c.Ctx().FormValue("table_name")

		if !data.MustCheck() {
			helper.Ajax("缺少必要参数", 1, c.Ctx())
			return
		}

		//if _, ok := data["status"]; ok {
		//	data["status"] = "1"
		//} else {
		//	data["status"] = "0"
		//}
		//
		//if data["description"] == "" {
		//	cont := bluemonday.NewPolicy().Sanitize(data["content"])
		//	if len(cont) > 250 {
		//		data["description"] = cont[:250]
		//	} else {
		//		data["description"] = cont
		//	}
		//}
		//
		//data["created_time"] = time.Now().In(helper.GetLocation()).Format(helper.TimeFormat)

		var fields []string
		var values []interface{}
		for k, v := range data {
			if k == "table_name" {
				continue
			}
			fields = append(fields, "`"+k+"`")
			values = append(values, v)
		}

		params := append([]interface{}{fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", controllers.GetTableName(model.Table), strings.Join(fields, ","), strings.TrimRight(strings.Repeat("?,", len(values)), ","))}, values...)
		// 先直接入库对应表内
		insertID, err := c.Ctx().Value("orm").(*xorm.Engine).Exec(params...)
		if err != nil {
			glog.Error(err)
			helper.Ajax("添加失败:"+err.Error(), 1, c.Ctx())
			return
		}
		id, _ := insertID.LastInsertId()
		if id > 0 {
			helper.Ajax(id, 0, c.Ctx())
		} else {
			helper.Ajax("添加失败", 1, c.Ctx())
		}
		return
	}
	//根据catid读取出相应的添加模板
	catid, _ := c.Ctx().GetInt64("catid")
	if catid == 0 {
		helper.Ajax("参数错误", 1, c.Ctx())
		return
	}
	cat := models.NewCategoryModel().GetCategory(catid)
	if cat == nil {
		helper.Ajax("分类不存在", 1, c.Ctx())
		return
	}
	if cat.Catid == 0 {
		helper.Ajax("不存在的分类", 1, c.Ctx())
		return
	}
	c.Ctx().Render().ViewData("category", cat)
	c.Ctx().Render().ViewData("submitURL", template.HTML("/b/content/add"))
	c.Ctx().Render().ViewData("preview", 0)
	c.Ctx().Render().HTML("backend/model_publish.html")
}

//EditContent 修改内容
func (c *ContentController) EditContent(orm *xorm.Engine) {
	//根据catid读取出相应的添加模板
	catid, _ := c.Ctx().GetInt64("catid")
	id, _ := c.Ctx().GetInt64("id")
	if catid < 1 || id < 1 {
		helper.Ajax("参数错误", 1, c.Ctx())
		return
	}
	catogoryModel := models.NewCategoryModel().GetCategory(catid)
	if catogoryModel == nil {
		helper.Ajax("分类不存在", 1, c.Ctx())
		return
	}
	if catogoryModel.ModelId < 1 {
		helper.Ajax("找不到关联模型", 1, c.Ctx())
		return
	}
	relationDocumentModel := models.NewDocumentModel().GetByID(catogoryModel.ModelId)
	if relationDocumentModel.Id == 0 {
		helper.Ajax("找不到关联模型", 1, c.Ctx())
		return
	}
	sql := []interface{}{fmt.Sprintf("SELECT * FROM `%s` WHERE catid=? and deleted_time IS NULL AND id = ? LIMIT 1", controllers.GetTableName(relationDocumentModel.Table)), catid, id}
	contents, err := orm.QueryString(sql...)
	if err != nil {
		c.Logger().Error(err)
		helper.Ajax("获取文章内容错误", 1, c.Ctx())
		return
	}
	if len(contents) == 0 {
		helper.Ajax("文章不存在或已删除", 1, c.Ctx())
		return
	}
	if c.Ctx().IsPost() {
		var data = customForm{}
		postData := c.Ctx().PostData()
		for formName, values := range postData {
			if formName == "flag" {
				data[formName] = strings.Join(values, ",")
			} else {
				data[formName] = values[0]
			}
		}
		data["catid"] = c.Ctx().GetString("catid")
		if !data.MustCheck() {
			helper.Ajax("缺少必要参数", 1, c.Ctx())
			return
		}
		delete(data, "id")

		if _, ok := data["status"]; ok {
			data["status"] = "1"
		} else {
			data["status"] = "0"
		}

		data["updated_time"] = time.Now().In(helper.GetLocation()).Format(helper.TimeFormat)
		var sets []string
		var values []interface{}

		if data["description"] == "" {
			cont := bluemonday.NewPolicy().Sanitize(data["content"])
			if len(cont) > 250 {
				data["description"] = cont[:250]
			} else {
				data["description"] = cont
			}
		}

		for k, v := range data {
			if k == "table_name" {
				continue
			}
			sets = append(sets, "`"+k+"`= ?")
			values = append(values, v)
		}

		values = append(values, id, catid)
		params := append([]interface{}{fmt.Sprintf("UPDATE `%s` SET %s WHERE id=? and catid=?", controllers.GetTableName(relationDocumentModel.Table), strings.Join(sets, ", "))}, values...)
		insertID, err := c.Ctx().Value("orm").(*xorm.Engine).Exec(params...)
		if err != nil {
			glog.Error(err)
			helper.Ajax("修改失败:"+err.Error(), 1, c.Ctx())
			return
		}
		res, _ := insertID.RowsAffected()
		if res > 0 {
			helper.Ajax("修改成功", 0, c.Ctx())
		} else {
			helper.Ajax("修改失败", 1, c.Ctx())
		}
		return
	}
	c.Ctx().Render().ViewData("category", catogoryModel)
	c.Ctx().Render().ViewData("submitURL", template.HTML("/b/content/edit"))
	c.Ctx().Render().ViewData("preview", 0)
	c.Ctx().Render().HTML("backend/model_publish.html")
}

//DeleteContent 删除内容
func (c *ContentController) DeleteContent(orm *xorm.Engine) {
	catid, _ := c.Ctx().GetInt64("catid")
	id, _ := c.Ctx().GetInt64("id")
	if catid < 1 || id < 1 {
		helper.Ajax("参数错误", 1, c.Ctx())
		return
	}
	catogoryModel := models.NewCategoryModel().GetCategory(catid)
	if catogoryModel == nil {
		helper.Ajax("分类不存在", 1, c.Ctx())
		return
	}
	if catogoryModel.ModelId < 1 {
		helper.Ajax("找不到关联模型", 1, c.Ctx())
		return
	}
	relationDocumentModel := models.NewDocumentModel().GetByID(catogoryModel.ModelId)
	if relationDocumentModel.Id == 0 {
		helper.Ajax("找不到关联模型", 1, c.Ctx())
		return
	}
	sqlOrArgs := []interface{}{fmt.Sprintf("UPDATE `%s` SET `deleted_time`='"+time.Now().In(helper.GetLocation()).Format(helper.TimeFormat)+"' WHERE id = ? and catid=?", controllers.GetTableName(relationDocumentModel.Table)), id, catid}
	res, err := orm.Exec(sqlOrArgs...)
	if err != nil {
		c.Logger().Error(err.Error())
		helper.Ajax("删除失败", 1, c.Ctx())
		return
	}
	if ret, _ := res.RowsAffected(); ret > 0 {
		helper.Ajax("删除成功", 0, c.Ctx())
	} else {
		helper.Ajax("删除失败", 1, c.Ctx())
	}
}

//排序内容
func (c *ContentController) OrderContent() {
	data := c.Ctx().PostData()
	var order = map[string]string{}
	for k, v := range data {
		order[strings.ReplaceAll(strings.ReplaceAll(k, "order[", ""), "]", "")] = v[0]
	}
	id, _ := c.Ctx().GetInt64("catid")
	if id < 1 {
		helper.Ajax("参数错误", 1, c.Ctx())
		return
	}
	catogoryModel := models.NewCategoryModel().GetCategory(id)
	if catogoryModel == nil {
		helper.Ajax("分类不存在", 1, c.Ctx())
		return
	}
	if catogoryModel.ModelId < 1 {
		helper.Ajax("找不到关联模型", 1, c.Ctx())
		return
	}
	relationDocumentModel := models.NewDocumentModel().GetByID(catogoryModel.ModelId)
	if relationDocumentModel.Id == 0 {
		helper.Ajax("找不到关联模型", 1, c.Ctx())
		return
	}
	for artID, orderNum := range order {
		sqlOrArgs := []interface{}{fmt.Sprintf("UPDATE `%s` SET `listorder`=? , updated_time = '"+time.Now().In(helper.GetLocation()).Format(helper.TimeFormat)+"' WHERE id = ? and catid=?", controllers.GetTableName(relationDocumentModel.Table)), orderNum, artID, id}
		if _, err := c.Ctx().Value("orm").(*xorm.Engine).Exec(sqlOrArgs...); err != nil {
			c.Logger().Error(err)
			helper.Ajax("更新文档排序失败", 1, c.Ctx())
			return
		}
	}
	helper.Ajax("更新排序成功", 0, c.Ctx())
}
