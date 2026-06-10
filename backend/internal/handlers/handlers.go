package handlers

import (
	"net/http"
	"script-kill-backend/internal/database"
	"script-kill-backend/internal/models"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateCarpoolRequest struct {
	ScriptID    uint   `json:"script_id" binding:"required"`
	HostName    string `json:"host_name" binding:"required"`
	HostContact string `json:"host_contact"`
}

type JoinCarpoolRequest struct {
	CarpoolID uint   `json:"carpool_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Contact   string `json:"contact"`
}

func GetScripts(c *gin.Context) {
	var scripts []models.Script
	result := database.DB.Find(&scripts)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, scripts)
}

func GetCarpools(c *gin.Context) {
	status := c.Query("status")
	var carpools []models.Carpool

	query := database.DB.Preload("Script").Preload("Room").Preload("Players")

	if status != "" {
		query = query.Where("status = ?", status)
	} else {
		query = query.Where("status IN ?", []string{"recruiting", "full"})
	}

	result := query.Order("created_at desc").Find(&carpools)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, carpools)
}

func CreateCarpool(c *gin.Context) {
	var req CreateCarpoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var script models.Script
	if err := database.DB.First(&script, req.ScriptID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "剧本不存在"})
		return
	}

	tx := database.DB.Begin()

	carpool := models.Carpool{
		ScriptID:        req.ScriptID,
		HostName:        req.HostName,
		HostContact:     req.HostContact,
		CurrentPlayers:  1,
		RequiredPlayers: script.PlayerCount,
		Status:          "recruiting",
	}

	if err := tx.Create(&carpool).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	player := models.Player{
		CarpoolID: carpool.ID,
		Name:      req.HostName,
		Contact:   req.HostContact,
		IsHost:    true,
	}

	if err := tx.Create(&player).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()

	database.DB.Preload("Script").Preload("Players").First(&carpool, carpool.ID)
	c.JSON(http.StatusCreated, carpool)
}

func JoinCarpool(c *gin.Context) {
	var req JoinCarpoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := database.DB.Begin()

	var carpool models.Carpool
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&carpool, req.CarpoolID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "拼车不存在"})
		return
	}

	if carpool.Status != "recruiting" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "该拼车已停止招募"})
		return
	}

	if carpool.CurrentPlayers >= carpool.RequiredPlayers {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "该拼车人数已满"})
		return
	}

	var existingCount int64
	tx.Model(&models.Player{}).Where("carpool_id = ? AND name = ?", req.CarpoolID, req.Name).Count(&existingCount)
	if existingCount > 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "你已加入该拼车"})
		return
	}

	player := models.Player{
		CarpoolID: req.CarpoolID,
		Name:      req.Name,
		Contact:   req.Contact,
		IsHost:    false,
	}

	if err := tx.Create(&player).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	carpool.CurrentPlayers++
	if carpool.CurrentPlayers >= carpool.RequiredPlayers {
		carpool.Status = "full"
	}

	if err := tx.Save(&carpool).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()

	database.DB.Preload("Script").Preload("Players").First(&carpool, carpool.ID)
	c.JSON(http.StatusOK, carpool)
}

func SeedData() {
	var count int64
	database.DB.Model(&models.Script{}).Count(&count)
	if count > 0 {
		return
	}

	scripts := []models.Script{
		{Name: "古木吟", PlayerCount: 6, Type: "情感/校园/变格", Difficulty: "中等", Duration: 240, Description: "20世纪末，岷江水畔的银杏高中，流传着一个关于古木的传说..."},
		{Name: "暗黑者", PlayerCount: 7, Type: "硬核/推理/机制", Difficulty: "困难", Duration: 300, Description: "法外制裁者暗黑者重现江湖，一场智慧与正义的较量..."},
		{Name: "窗边的女人", PlayerCount: 6, Type: "惊悚/推理/本格", Difficulty: "中等", Duration: 210, Description: "根据真实案件改编，天黑请闭眼，死神在窗外窥视..."},
		{Name: "年轮", PlayerCount: 5, Type: "硬核/推理/变格", Difficulty: "困难", Duration: 300, Description: "位于T市郊外的一个破旧仓库内，一场诡异的聚会..."},
		{Name: "漓川怪谈簿", PlayerCount: 7, Type: "日式/推理/变格", Difficulty: "中等", Duration: 270, Description: "常世是这片土地的名字，也是土地中心那片诡异湖泊的名字..."},
		{Name: "破晓", PlayerCount: 6, Type: "情感/沉浸/现代", Difficulty: "简单", Duration: 270, Description: "禁毒大队的年轻警员们，面对的不仅是毒贩，还有人性的考验..."},
		{Name: "纸帆", PlayerCount: 6, Type: "情感/推理/沉浸", Difficulty: "中等", Duration: 240, Description: "一艘纸帆，承载着怎样的秘密？那年夏天，我们失去了什么..."},
		{Name: "就像水消失在水中", PlayerCount: 6, Type: "情感/沉浸/现代", Difficulty: "简单", Duration: 240, Description: "我们都像水，消失在时代的洪流中，却又无处不在..."},
	}

	for _, s := range scripts {
		database.DB.Create(&s)
	}

	rooms := []models.Room{
		{Name: "古风厅", Capacity: 8, Status: "available"},
		{Name: "现代厅", Capacity: 8, Status: "available"},
		{Name: "恐怖厅", Capacity: 6, Status: "available"},
		{Name: "日式厅", Capacity: 7, Status: "available"},
		{Name: "欧式厅", Capacity: 8, Status: "available"},
	}

	for _, r := range rooms {
		database.DB.Create(&r)
	}

	now := time.Now()
	demoCarpools := []struct {
		Carpool models.Carpool
		Players []models.Player
	}{
		{
			Carpool: models.Carpool{
				ScriptID:        2,
				HostName:        "张三",
				HostContact:     "13800138001",
				CurrentPlayers:  5,
				RequiredPlayers: 7,
				Status:          "recruiting",
				CreatedAt:       now,
			},
			Players: []models.Player{
				{Name: "张三", IsHost: true},
				{Name: "李四", IsHost: false},
				{Name: "王五", IsHost: false},
				{Name: "赵六", IsHost: false},
				{Name: "钱七", IsHost: false},
			},
		},
		{
			Carpool: models.Carpool{
				ScriptID:        1,
				HostName:        "小美",
				HostContact:     "13800138002",
				CurrentPlayers:  4,
				RequiredPlayers: 6,
				Status:          "recruiting",
				CreatedAt:       now.Add(-30 * time.Minute),
			},
			Players: []models.Player{
				{Name: "小美", IsHost: true},
				{Name: "小丽", IsHost: false},
				{Name: "小芳", IsHost: false},
				{Name: "小华", IsHost: false},
			},
		},
		{
			Carpool: models.Carpool{
				ScriptID:        4,
				HostName:        "阿杰",
				HostContact:     "13800138003",
				CurrentPlayers:  5,
				RequiredPlayers: 5,
				Status:          "full",
				CreatedAt:       now.Add(-1 * time.Hour),
			},
			Players: []models.Player{
				{Name: "阿杰", IsHost: true},
				{Name: "阿明", IsHost: false},
				{Name: "阿丽", IsHost: false},
				{Name: "阿强", IsHost: false},
				{Name: "阿红", IsHost: false},
			},
		},
	}

	for _, item := range demoCarpools {
		tx := database.DB.Begin()
		carpool := item.Carpool
		tx.Create(&carpool)
		for i := range item.Players {
			player := item.Players[i]
			player.CarpoolID = carpool.ID
			tx.Create(&player)
		}
		tx.Commit()
	}

	println("Sample data seeded")
}

func AutoMigrate() {
	database.DB.AutoMigrate(&models.Script{}, &models.Room{}, &models.Carpool{}, &models.Player{})
}
