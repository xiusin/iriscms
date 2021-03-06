# 介绍 

PineCMS 是一个基于 `golang` 的内容管理系统, 可以用来快速构建网站模板, 普通网站可在两个小时内完成整站建设. 其特点如下: 

- 🚀 基于`Go`语言开发, 性能卓越
- 😺内置`hot reload`功能,可在开发过程中不再受<kbd>ctrl</kbd>+<kbd>c</kbd>的困扰
- 🚀自动静态化页面. 可以无侵入式由`nginx`代理服务
- 🏷标签化调用数据, 可以在不写SQL代码的情况下完成页面渲染
- 🍵自动从`dede`导入项目数据以及模板标签替换, 节省`80%`的模板替换时间
- 📚支持SQLite3内嵌数据库
- 🚪支持多主题, 让您的项目不再单调

# 系统模块
PineCMS 包括了以下两个模块.

### CMD  模块
服务的启动和数据导入功能, 您可以直接使用命令`web`服务和`db`的数据导入.

#### SERVE 命令
- `serve` 用于启动服务器的命令
    - `dev` 启动一个开发服务器
    - `run` 启动一个生产服务器
    
#### IMPORT 命令
- `import` 用户其他CMS导入数据和标签替换
    - `dede` 导入织梦数据
    - `dede_tpl` 替换织梦标签,并且在`themes`下生成指定目录名
    


# SRC 模块
源代码模块, 您的核心业务均在这里编写开发.
 
> 非开发者无需关心此模块