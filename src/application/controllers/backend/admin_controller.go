package backend
//
//import (
//	"github.com/go-xorm/xorm"
//	"github.com/xiusin/pine"
//	"github.com/xiusin/pinecms/src/application/controllers"
//	"strconv"
//
//	"github.com/xiusin/pinecms/src/application/models"
//	"github.com/xiusin/pinecms/src/application/models/tables"
//	"github.com/xiusin/pinecms/src/common/helper"
//)
//
//type AdminController struct {
//	pine.Controller
//}
//
//func (c *AdminController) RegisterRoute(b pine.IRouterWrapper) {
//	b.ANY("/admin/info", "AdminInfo")
//	b.ANY("/admin/memberlist", "Memberlist")
//	b.ANY("/admin/public-editpwd", "PublicEditpwd")
//	b.ANY("/admin/public-editinfo", "PublicEditInfo")
//	b.POST("/admin/public-checkEmail", "PublicCheckEmail")
//	b.POST("/admin/public-checkName", "PubicCheckName")
//	b.POST("/admin/check-rolename", "PublicCheckRoleName")
//	b.POST("/admin/check-password", "PublicCheckPassword")
//	b.ANY("/admin/member-add", "MemberAdd")
//	b.ANY("/admin/member-edit", "MemberEdit")
//	b.ANY("/admin/member-delete", "MemberDelete")
//}
//
////用于用户列表数据格式返回
//type memberlist struct {
//	Email         string `json:"email"`
//	Lastloginip   string `json:"lastloginip"`
//	Lastlogintime string `json:"lastlogintime"`
//	Realname      string `json:"realname"`
//	Rolename      string `json:"rolename"`
//	Userid        int64  `json:"userid"`
//	Roleid        int64  `json:"roleid"`
//	Username      string `json:"username"`
//}
//
//func (c *AdminController) AdminInfo() {
//	aid, _ := c.Ctx().Value("adminid").(int64) //检测是否设置过session
//	info, err := models.NewAdminModel().GetUserInfo(aid)
//	if err != nil {
//		helper.Ajax(err.Error(), 1, c.Ctx())
//		return
//	}
//	helper.Ajax(info, 0, c.Ctx())
//}
//
//func (c *AdminController) PublicEditInfo(orm *xorm.Engine) {
//	aid, _ := c.Ctx().Value("adminid").(int64) //检测是否设置过session
//	if c.Ctx().IsPost() {
//		info := tables.Admin{
//			Userid: aid,
//		}
//		has, _ := orm.Get(&info) //读取用户资料
//		if !has {
//			helper.Ajax("用户资料已经不存在", 1, c.Ctx())
//		} else {
//			info.Realname = c.Ctx().PostString("realname")
//			info.Email = c.Ctx().PostString("email")
//			info.Avatar = c.Ctx().PostString("avatar")
//			res, err := orm.Id(aid).MustCols("avatar", "email", "realname").Update(info)
//			if err != nil {
//				helper.Ajax("修改资料失败"+err.Error(), 1, c.Ctx())
//			} else {
//				if res > 0 {
//					helper.Ajax("修改资料成功", 0, c.Ctx())
//				} else {
//					helper.Ajax("修改资料失败", 1, c.Ctx())
//				}
//			}
//		}
//		return
//	}
//	info, err := models.NewAdminModel().GetUserInfo(aid)
//	if err != nil {
//		helper.Ajax(err.Error(), 1, c.Ctx())
//		return
//	}
//	helper.Ajax(info, 0, c.Ctx())
//}
//
//func (c *AdminController) Memberlist() {
//
//	page, err := c.Ctx().GetInt("page")
//	orderField := c.Ctx().GetString("sort")
//	if orderField == "" {
//		orderField = "userid"
//	}
//	orderType := c.Ctx().GetString("sort")
//	if orderType == "" {
//		orderType = "desc"
//	}
//
//	if err != nil {
//		page = 1
//	}
//	data := models.NewAdminModel().GetList("1", page, 10, orderField, orderType)
//	var retData []memberlist
//	//将数据以map的方式返回吧.
//	for _, v := range data {
//		item := memberlist{
//			Email:         v.Email,
//			Lastloginip:   v.Lastloginip,
//			Lastlogintime: helper.FormatTime(v.Lastlogintime),
//			Realname:      v.Realname,
//			Rolename:      "",
//			Userid:        v.Userid,
//			Username:      v.Username,
//			Roleid:        v.Roleid,
//		}
//		roleInfo, err := models.NewAdminRoleModel().GetRoleById(v.Roleid)
//		if err != nil {
//			roleInfo.Rolename = ""
//		}
//		item.Rolename = roleInfo.Rolename
//		retData = append(retData, item)
//	}
//	helper.Ajax(pine.H{"total": len(retData), "rows": retData}, 0, c.Ctx())
//}
//
//func (c *AdminController) PublicEditpwd() {
//	aid, _ := c.Ctx().Value("adminid").(int64)
//	menuid, _ := c.Ctx().GetInt64("menuid")
//	info := tables.Admin{Userid: int64(aid)}
//	has, _ := pine.Make(controllers.ServiceXorm).(*xorm.Engine).Get(&info)
//	if !has {
//		c.Ctx().Write([]byte("没有找到"))
//		return
//	}
//	if c.Ctx().IsPost() {
//		if info.Password != helper.Password(c.Ctx().PostValue("old_password"), info.Encrypt) {
//			helper.Ajax("原密码错误", 1, c.Ctx())
//			return
//		}
//		info.Password = helper.Password(c.Ctx().PostValue("new_password"), info.Encrypt)
//		res, _ := pine.Make(controllers.ServiceXorm).(*xorm.Engine).Id(aid).Update(info)
//		if res > 0 {
//			helper.Ajax("修改资料成功", 0, c.Ctx())
//		} else {
//			helper.Ajax("修改资料失败", 1, c.Ctx())
//		}
//		return
//	}
//	c.Ctx().Render().ViewData("currentpos", models.NewMenuModel().CurrentPos(menuid))
//	c.Ctx().Render().ViewData("admin", info)
//	c.Ctx().Render().HTML("backend/admin_editpwd.html")
//}
//
//func (c *AdminController) PublicCheckEmail() {
//	info := &tables.Admin{Username: c.Ctx().FormValue("name")}
//	has, _ := pine.Make(controllers.ServiceXorm).(*xorm.Engine).Get(info)
//	if !has {
//		helper.Ajax("没有相同的用户名", 0, c.Ctx())
//	} else {
//		helper.Ajax("已经有相同的用户名,请换一个再试", 1, c.Ctx())
//	}
//}
//func (c *AdminController) PublicCheckPassword() {
//	aid, _ := c.Ctx().Value("adminid").(int64)
//	password := c.Ctx().FormValue("password")
//	admin, err := models.NewAdminModel().GetUserInfo(aid)
//	if err != nil {
//		helper.Ajax("无法查找到相关信息", 1, c.Ctx())
//		return
//	}
//	if admin.Password != helper.Password(password, admin.Encrypt) {
//		helper.Ajax("旧密码错误", 1, c.Ctx())
//		return
//	}
//	helper.Ajax("验证密码成功", 0, c.Ctx())
//}
//func (c *AdminController) PubicCheckName() {
//	info := &tables.Admin{Username: c.Ctx().FormValue("name")}
//	uid, _ := c.Ctx().GetInt64("id")
//	has, _ := c.Ctx().Value("orm").(*xorm.Engine).Get(info)
//	if !has || info.Userid == uid {
//		helper.Ajax("没有相同的用户名", 0, c.Ctx())
//	} else {
//		helper.Ajax("已经有相同的用户名,请换一个再试", 1, c.Ctx())
//	}
//}
//
//func (c *AdminController) PublicCheckRoleName() {
//	rolename := c.Ctx().FormValue("rolename")
//	id, _ := c.Ctx().GetInt64("id")
//	if !c.Ctx().IsAjax() || rolename == "" {
//		helper.Ajax("参数错误 ,"+rolename, 1, c.Ctx())
//		return
//	}
//	if models.NewAdminRoleModel().CheckRoleName(id, rolename) {
//		helper.Ajax("角色已存在", 1, c.Ctx())
//		return
//	}
//	helper.Ajax("通过", 0, c.Ctx())
//}
//func (c *AdminController) MemberAdd(orm *xorm.Engine) {
//	if c.Ctx().FormValue("pwdconfirm") != c.Ctx().FormValue("password") || c.Ctx().FormValue("password") == "" {
//		helper.Ajax("两次密码不一致", 1, c.Ctx())
//		return
//	}
//	if c.Ctx().FormValue("roleid") == "" {
//		helper.Ajax("请选择角色", 1, c.Ctx())
//		return
//	}
//	roleid, err := strconv.Atoi(c.Ctx().FormValue("roleid"))
//	if err != nil {
//		helper.Ajax("角色信息错误", 1, c.Ctx())
//		return
//	}
//	str := string(helper.Krand(6, 3))
//
//	newAdmin := &tables.Admin{
//		Username: c.Ctx().FormValue("username"),
//		Password: helper.Password(c.Ctx().FormValue("password"), str),
//		Email:    c.Ctx().FormValue("email"),
//		Encrypt:  str,
//		Realname: c.Ctx().FormValue("realname"),
//		Roleid:   int64(roleid),
//	}
//	// 检查账户是否存在
//	total, _ := orm.Where("username = ?", newAdmin.Username).Table(newAdmin).Count()
//	if total > 0 {
//		helper.Ajax("用户已存在, 请换个再试", 1, c.Ctx())
//		return
//	}
//	id, err := orm.Insert(newAdmin)
//	if id > 0 {
//		helper.Ajax("添加管理员成功", 0, c.Ctx())
//		return
//	}
//	helper.Ajax("添加管理员失败", 1, c.Ctx())
//}
//func (c *AdminController) MemberEdit() {
//	adminid, err := c.Ctx().GetInt64("id")
//	if err != nil {
//		c.Ctx().WriteString("参数错误 : " + err.Error())
//		return
//	}
//	info, err := models.NewAdminModel().GetUserInfo(adminid)
//	if err != nil {
//		c.Ctx().WriteString("没有该管理员信息")
//		return
//	}
//
//	if c.Ctx().FormValue("password") != "" {
//		if c.Ctx().FormValue("pwdconfirm") != c.Ctx().FormValue("password") {
//			helper.Ajax("两次密码不一致", 1, c.Ctx())
//			return
//		}
//	}
//	if c.Ctx().FormValue("roleid") == "" {
//		helper.Ajax("请选择角色", 1, c.Ctx())
//		return
//	}
//	roleid, err := strconv.Atoi(c.Ctx().FormValue("roleid"))
//	if err != nil {
//		helper.Ajax("角色信息错误", 1, c.Ctx())
//		return
//	}
//	info.Username = c.Ctx().FormValue("username")
//	info.Email = c.Ctx().FormValue("email")
//	info.Realname = c.Ctx().FormValue("realname")
//	info.Roleid = int64(roleid)
//	if c.Ctx().FormValue("password") != "" {
//		info.Password = helper.Password(c.Ctx().PostValue("password"), info.Encrypt)
//	}
//	res, err := c.Ctx().Value("orm").(*xorm.Engine).Where("userid = ?", info.Userid).Update(info)
//	if err != nil {
//		helper.Ajax(err.Error(), 0, c.Ctx())
//		return
//	}
//	if res > 0 {
//		helper.Ajax("修改管理员成功", 0, c.Ctx())
//		return
//	}
//	helper.Ajax("修改管理员失败", 1, c.Ctx())
//	return
//
//}
//func (c *AdminController) MemberDelete() {
//	id, err := strconv.Atoi(c.Ctx().FormValue("id"))
//	if err != nil || helper.IsFalse(id) {
//		helper.Ajax("参数错误", 1, c.Ctx())
//		return
//	}
//	if id == 1 {
//		helper.Ajax("禁止删除初始管理员", 1, c.Ctx())
//		return
//	}
//	deleteAdmin := &tables.Admin{Userid: int64(id)}
//	res, err := c.Ctx().Value("orm").(*xorm.Engine).Delete(deleteAdmin)
//	if err != nil || helper.IsFalse(res) {
//		helper.Ajax("删除失败", 1, c.Ctx())
//		return
//	}
//	helper.Ajax("删除成功", 0, c.Ctx())
//}
