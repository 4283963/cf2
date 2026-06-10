package handlers

import (
	"errors"
	"log"
	"net/http"
	"script-kill-backend/internal/database"
	"script-kill-backend/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "剧本不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if script.PlayerCount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "剧本人数配置错误"})
		return
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "开启事务失败"})
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

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

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败"})
		return
	}

	var finalCarpool models.Carpool
	database.DB.Preload("Script").Preload("Players").First(&finalCarpool, carpool.ID)
	c.JSON(http.StatusCreated, finalCarpool)
}

func JoinCarpool(c *gin.Context) {
	var req JoinCarpoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "开启事务失败"})
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var carpool models.Carpool
	if err := tx.Clauses().First(&carpool, req.CarpoolID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "拼车不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if carpool.Status != "recruiting" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "该拼车已停止招募"})
		return
	}

	var existingCount int64
	if err := tx.Model(&models.Player{}).Where("carpool_id = ? AND name = ?", req.CarpoolID, req.Name).Count(&existingCount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingCount > 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "你已加入该拼车"})
		return
	}

	result := tx.Model(&models.Carpool{}).
		Where("id = ? AND status = ? AND current_players < required_players", req.CarpoolID, "recruiting").
		Updates(map[string]interface{}{
			"current_players": gorm.Expr("current_players + 1"),
			"status": gorm.Expr(
				"CASE WHEN current_players + 1 >= required_players THEN ? ELSE status END",
				"full",
			),
		})

	if result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		var recheck models.Carpool
		database.DB.First(&recheck, req.CarpoolID)
		if recheck.CurrentPlayers >= recheck.RequiredPlayers {
			c.JSON(http.StatusBadRequest, gin.H{"error": "手慢了！该拼车人数刚刚已满"})
		} else if recheck.Status != "recruiting" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "该拼车已停止招募"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "加入失败，请重试"})
		}
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

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败"})
		return
	}

	var finalCarpool models.Carpool
	database.DB.Preload("Script").Preload("Players").First(&finalCarpool, req.CarpoolID)
	c.JSON(http.StatusOK, finalCarpool)
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
	EnforceDataConsistency()
}

func EnforceDataConsistency() {
	logPrefix := "[Consistency]"

	var overflows []models.Carpool
	database.DB.Raw(`
		SELECT * FROM carpools 
		WHERE current_players > required_players 
		   OR (status = 'full' AND current_players < required_players)
		   OR (status = 'recruiting' AND current_players >= required_players)
		AND deleted_at IS NULL
	`).Scan(&overflows)

	if len(overflows) > 0 {
		log.Printf("%s Found %d inconsistent carpool records, fixing...", logPrefix, len(overflows))
	}

	for _, c := range overflows {
		fixedCurrent := c.CurrentPlayers
		fixedStatus := c.Status

		if fixedCurrent > c.RequiredPlayers {
			log.Printf("%s Carpool #%d: current_players=%d > required=%d, truncating to %d",
				logPrefix, c.ID, fixedCurrent, c.RequiredPlayers, c.RequiredPlayers)
			fixedCurrent = c.RequiredPlayers
		}

		if fixedCurrent >= c.RequiredPlayers && fixedStatus != "full" {
			log.Printf("%s Carpool #%d: current >= required, marking as 'full'", logPrefix, c.ID)
			fixedStatus = "full"
		} else if fixedCurrent < c.RequiredPlayers && fixedStatus == "full" {
			log.Printf("%s Carpool #%d: current < required but marked 'full', reverting to 'recruiting'", logPrefix, c.ID)
			fixedStatus = "recruiting"
		}

		database.DB.Model(&models.Carpool{}).Where("id = ?", c.ID).Updates(map[string]interface{}{
			"current_players": fixedCurrent,
			"status":          fixedStatus,
		})

		var playerCount int64
		database.DB.Model(&models.Player{}).Where("carpool_id = ? AND deleted_at IS NULL", c.ID).Count(&playerCount)
		if int64(fixedCurrent) != playerCount {
			log.Printf("%s Carpool #%d: current_players(%d) != actual players(%d), syncing",
				logPrefix, c.ID, fixedCurrent, playerCount)
			actualCount := int(playerCount)
			if actualCount > c.RequiredPlayers {
				actualCount = c.RequiredPlayers
			}
			actualStatus := "recruiting"
			if actualCount >= c.RequiredPlayers {
				actualStatus = "full"
			}
			database.DB.Model(&models.Carpool{}).Where("id = ?", c.ID).Updates(map[string]interface{}{
				"current_players": actualCount,
				"status":          actualStatus,
			})
		}
	}

	if len(overflows) > 0 {
		log.Printf("%s Consistency repair complete", logPrefix)
	}
}
