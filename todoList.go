package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"net/http"
	"os"
)

var DB *gorm.DB
var ERR error

func initDataBase() (*gorm.DB, error) {
	dns := "root:123456@(0.0.0.0:3306)/todolist?charset=utf8&parseTime=True&loc=Local"
	var err error
	DB, err = gorm.Open("mysql", dns)
	return DB, err
}

type ToDoList struct {
	ID     int    `json:"id" gorm:"primary_key"`
	Title  string `json:"title"`
	Status bool   `json:"status"`
}

func main() {
	//连接数据库
	DB, ERR = initDataBase()
	handeErr(ERR)
	//建表
	DB.AutoMigrate(&ToDoList{})
	//获取驱动
	engine := gin.Default()
	engine.Static("/static", "./static")
	//加载解析template
	engine.LoadHTMLGlob("./templates/*")

	//渲染模板
	engine.GET("/", func(context *gin.Context) {
		context.HTML(http.StatusOK, "index.html", nil)
	})

	routeGroup := engine.Group("/v1")
	{
		//新增一个代办事项
		routeGroup.POST("/todo", func(context *gin.Context) {
			ptrToDoList := new(ToDoList)
			//绑定数据
			ERR = context.ShouldBind(ptrToDoList)
			handeErr(ERR)
			//存入数据库
			err := DB.Create(ptrToDoList).Error
			if err != nil {
				context.JSON(http.StatusBadRequest, gin.H{
					"err": err,
				})
			} else {
				context.JSON(http.StatusOK, ptrToDoList)
			}
		})

		//查看所有的待办事项
		routeGroup.GET("/todo", func(context *gin.Context) {
			toDoLists := new([]ToDoList)
			err := DB.Find(toDoLists).Error
			if err != nil {
				context.JSON(http.StatusBadRequest, gin.H{
					"err": err,
				})
			} else {
				context.JSON(http.StatusOK, toDoLists)
			}

		})

		//更新待办事项的状态
		routeGroup.PUT("/todo/:id", func(context *gin.Context) {
			getParam, ok := context.Params.Get("id")

			if !ok {
				context.JSON(http.StatusBadRequest, gin.H{
					"err": "获取更新参数失败",
				})
				return
			}
			todo := new(ToDoList)
			DB.Where("id=?", getParam).First(todo)

			err := context.BindJSON(todo)
			if err != nil {
				context.JSON(http.StatusBadRequest, gin.H{
					"err": err,
				})
				return
			}
			err = DB.Save(todo).Error

			if err != nil {
				context.JSON(http.StatusBadRequest, gin.H{
					"err": err,
				})
				return
			}
			context.JSON(http.StatusOK, todo)
		})

		//删除待办事项
		routeGroup.DELETE("/todo/:id", func(context *gin.Context) {
			getParam, ok := context.Params.Get("id")
			if !ok {
				context.JSON(http.StatusBadRequest, gin.H{
					"err": "获取删除参数id失败",
				})
				return
			}
			err := DB.Delete(&ToDoList{}, getParam).Error

			if err != nil {
				context.JSON(http.StatusBadRequest, gin.H{
					"err": "数据库中记录删除失败",
				})
				return
			}
			context.JSON(http.StatusOK, gin.H{
				"status": "删除成功",
			})

		})
	}

	engine.Run("0.0.0.0:8081")
}

func handeErr(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
