package handlers

import (
	"fmt"
	"log"
	"script-kill-backend/internal/database"
	"script-kill-backend/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JoinWaitlistRequest struct {
	CarpoolID uint   `json:"carpool_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Contact   string `json:"contact"`
}

type LeaveCarpoolRequest struct {
	CarpoolID uint   `json:"carpool_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type CreateCarpoolV2Request struct {
	ScriptID      uint       `json:"script_id" binding:"required"`
	HostName      string     `json:"host_name" binding:"required"`
	HostContact   string     `json:"host_contact"`
	StartTime     *time.Time `json:"start_time"`
	DepositAmount float64    `json:"deposit_amount"`
}

func GetWaitlist(c *gin.Context) {
	carpoolID := c.Query("carpool_id")
	var waitlist []models.Waitlist

	query := database.DB.Where("status = ?", "waiting").Order("position asc, created_at asc")
	if carpoolID != "" {
		query = query.Where("carpool_id = ?", carpoolID)
	}

	result := query.Find(&waitlist)
	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(200, waitlist)
}

func JoinWaitlist(c *gin.Context) {
	var req JoinWaitlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(500, gin.H{"error": "开启事务失败"})
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var carpool models.Carpool
	if err := tx.Preload("Script").First(&carpool, req.CarpoolID).Error; err != nil {
		tx.Rollback()
		c.JSON(404, gin.H{"error": "拼车不存在"})
		return
	}

	if carpool.Status == "cancelled" || carpool.Status == "completed" {
		tx.Rollback()
		c.JSON(400, gin.H{"error": "该拼车已结束或已取消"})
		return
	}

	var inCarCount int64
	tx.Model(&models.Player{}).Where("carpool_id = ? AND name = ?", req.CarpoolID, req.Name).Count(&inCarCount)
	if inCarCount > 0 {
		tx.Rollback()
		c.JSON(400, gin.H{"error": "你已在拼车中，无需候补"})
		return
	}

	var alreadyCount int64
	tx.Model(&models.Waitlist{}).Where("carpool_id = ? AND name = ? AND status = ?", req.CarpoolID, req.Name, "waiting").Count(&alreadyCount)
	if alreadyCount > 0 {
		tx.Rollback()
		c.JSON(400, gin.H{"error": "你已在候补队列中"})
		return
	}

	var nextPos int64
	tx.Model(&models.Waitlist{}).Where("carpool_id = ? AND status = ?", req.CarpoolID, "waiting").Count(&nextPos)

	waitlistItem := models.Waitlist{
		CarpoolID: req.CarpoolID,
		Name:      req.Name,
		Contact:   req.Contact,
		Position:  int(nextPos) + 1,
		Status:    "waiting",
	}
	if err := tx.Create(&waitlistItem).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	PushNotification(tx, req.Name, req.CarpoolID, "waitlist_joined",
		fmt.Sprintf("你已成功加入《%s》候补队列，当前第 %d 顺位", carpool.Script.Name, waitlistItem.Position))

	if err := tx.Commit().Error; err != nil {
		c.JSON(500, gin.H{"error": "提交事务失败"})
		return
	}

	c.JSON(200, gin.H{
		"message":  fmt.Sprintf("候补成功，当前第 %d 顺位", waitlistItem.Position),
		"waitlist": waitlistItem,
		"position": waitlistItem.Position,
	})
}

func LeaveCarpool(c *gin.Context) {
	var req LeaveCarpoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(500, gin.H{"error": "开启事务失败"})
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var carpool models.Carpool
	if err := tx.Preload("Script").First(&carpool, req.CarpoolID).Error; err != nil {
		tx.Rollback()
		c.JSON(404, gin.H{"error": "拼车不存在"})
		return
	}

	if carpool.Status == "cancelled" || carpool.Status == "completed" {
		tx.Rollback()
		c.JSON(400, gin.H{"error": "该拼车已结束或已取消"})
		return
	}

	var player models.Player
	if err := tx.Where("carpool_id = ? AND name = ?", req.CarpoolID, req.Name).First(&player).Error; err != nil {
		tx.Rollback()
		c.JSON(400, gin.H{"error": "你不在该拼车中"})
		return
	}

	if player.IsHost {
		tx.Rollback()
		c.JSON(400, gin.H{"error": "车主不能退车，请先转让车主或直接取消拼车"})
		return
	}

	if err := tx.Delete(&player).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if player.DepositPaid {
		log.Printf("[Refund] 玩家 %s 退车，押金 %.2f 已自动退回", player.Name, carpool.DepositAmount)
		PushNotification(tx, player.Name, req.CarpoolID, "deposit_refunded",
			fmt.Sprintf("你已退车，押金 %.2f 元已退回", carpool.DepositAmount))
	}

	PushNotification(tx, carpool.HostName, req.CarpoolID, "player_left",
		fmt.Sprintf("玩家 %s 已退出《%s》拼车", player.Name, carpool.Script.Name))

	decResult := tx.Model(&models.Carpool{}).
		Where("id = ?", req.CarpoolID).
		Updates(map[string]interface{}{
			"current_players": gorm.Expr("CASE WHEN current_players > 0 THEN current_players - 1 ELSE 0 END"),
			"status":          "recruiting",
		})
	if decResult.Error != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": decResult.Error.Error()})
		return
	}

	var promoted *models.Waitlist
	for {
		var next models.Waitlist
		err := tx.Where("carpool_id = ? AND status = ?", req.CarpoolID, "waiting").
			Order("position asc, created_at asc").
			First(&next).Error
		if err != nil {
			break
		}

		var dupCheck int64
		tx.Model(&models.Player{}).Where("carpool_id = ? AND name = ?", req.CarpoolID, next.Name).Count(&dupCheck)
		if dupCheck > 0 {
			tx.Model(&next).Update("status", "skipped_duplicate")
			continue
		}

		promoteResult := tx.Model(&models.Carpool{}).
			Where("id = ? AND status = ? AND current_players < required_players", req.CarpoolID, "recruiting").
			Updates(map[string]interface{}{
				"current_players": gorm.Expr("current_players + 1"),
				"status": gorm.Expr(
					"CASE WHEN current_players + 1 >= required_players THEN ? ELSE status END",
					"full",
				),
			})
		if promoteResult.Error != nil || promoteResult.RowsAffected == 0 {
			break
		}

		newPlayer := models.Player{
			CarpoolID: req.CarpoolID,
			Name:      next.Name,
			Contact:   next.Contact,
			IsHost:    false,
		}
		if err := tx.Create(&newPlayer).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		tx.Model(&next).Updates(map[string]interface{}{
			"status": "promoted",
		})

		promoted = &next

		var allWaiters []models.Waitlist
		tx.Where("carpool_id = ? AND status = ? AND id > ?", req.CarpoolID, "waiting", next.ID).
			Order("position asc, created_at asc").
			Find(&allWaiters)
		for i, w := range allWaiters {
			tx.Model(&w).Update("position", i+2)
		}

		PushNotification(tx, next.Name, req.CarpoolID, "promoted_from_waitlist",
			fmt.Sprintf("🎉 恭喜！有玩家退车，你已自动补位加入《%s》拼车！", carpool.Script.Name))

		PushNotification(tx, carpool.HostName, req.CarpoolID, "waitlist_promoted",
			fmt.Sprintf("候补玩家 %s 已自动补位加入", next.Name))

		break
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(500, gin.H{"error": "提交事务失败"})
		return
	}

	var finalCarpool models.Carpool
	database.DB.Preload("Script").Preload("Players").Preload("Waitlist", "status = ?", "waiting").
		First(&finalCarpool, req.CarpoolID)

	response := gin.H{
		"message": "退车成功",
		"carpool": finalCarpool,
	}
	if promoted != nil {
		response["promoted"] = gin.H{
			"name":     promoted.Name,
			"position": promoted.Position,
		}
	} else {
		response["promoted"] = nil
	}
	c.JSON(200, response)
}

func CancelCarpool(c *gin.Context) {
	var req struct {
		CarpoolID uint   `json:"carpool_id" binding:"required"`
		Name      string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(500, gin.H{"error": "开启事务失败"})
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var carpool models.Carpool
	if err := tx.Preload("Script").First(&carpool, req.CarpoolID).Error; err != nil {
		tx.Rollback()
		c.JSON(404, gin.H{"error": "拼车不存在"})
		return
	}

	if carpool.HostName != req.Name {
		tx.Rollback()
		c.JSON(403, gin.H{"error": "只有车主可以取消拼车"})
		return
	}

	if carpool.Status == "cancelled" || carpool.Status == "completed" {
		tx.Rollback()
		c.JSON(400, gin.H{"error": "拼车已结束"})
		return
	}

	var players []models.Player
	tx.Where("carpool_id = ?", req.CarpoolID).Find(&players)
	for _, p := range players {
		if p.DepositPaid && carpool.DepositAmount > 0 {
			log.Printf("[Refund] 拼车取消，玩家 %s 押金 %.2f 已退回", p.Name, carpool.DepositAmount)
			PushNotification(tx, p.Name, req.CarpoolID, "deposit_refunded",
				fmt.Sprintf("拼车已取消，你的押金 %.2f 元已退回", carpool.DepositAmount))
		} else {
			PushNotification(tx, p.Name, req.CarpoolID, "carpool_cancelled",
				fmt.Sprintf("《%s》拼车已被车主取消", carpool.Script.Name))
		}
	}

	var waiters []models.Waitlist
	tx.Where("carpool_id = ? AND status = ?", req.CarpoolID, "waiting").Find(&waiters)
	for _, w := range waiters {
		tx.Model(&w).Update("status", "cancelled")
		PushNotification(tx, w.Name, req.CarpoolID, "carpool_cancelled",
			fmt.Sprintf("《%s》拼车已取消，候补资格失效", carpool.Script.Name))
	}

	tx.Model(&carpool).Update("status", "cancelled")

	if err := tx.Commit().Error; err != nil {
		c.JSON(500, gin.H{"error": "提交事务失败"})
		return
	}

	c.JSON(200, gin.H{"message": "拼车已取消，押金已自动退回", "carpool_id": req.CarpoolID})
}

func GetNotifications(c *gin.Context) {
	user := c.Query("user")
	if user == "" {
		c.JSON(400, gin.H{"error": "请指定 user 参数"})
		return
	}
	var notifs []models.Notification
	database.DB.Where("\"user\" = ?", user).Order("created_at desc").Limit(50).Find(&notifs)
	c.JSON(200, notifs)
}

func PushNotification(tx *gorm.DB, user string, carpoolID uint, notifType string, message string) {
	notif := models.Notification{
		User:      user,
		CarpoolID: carpoolID,
		Type:      notifType,
		Message:   message,
		IsRead:    false,
	}
	tx.Create(&notif)
	log.Printf("[Notify] %s | [%s] %s", user, notifType, message)
}

func CheckTimeoutAndCancel() {
	log.Println("[Scheduler] 扫描即将超时的拼车...")

	now := time.Now()
	cutoff := now.Add(1 * time.Hour)

	var carpools []models.Carpool
	database.DB.Preload("Script").Preload("Players").Preload("Waitlist").
		Where("status IN ? AND start_time IS NOT NULL AND start_time <= ?",
			[]string{"recruiting", "full"}, cutoff).
		Find(&carpools)

	for _, carpool := range carpools {
		if carpool.Status == "full" {
			continue
		}

		log.Printf("[Scheduler] 拼车 #%d 《%s》距开局不足1小时，仍差 %d 人，自动解散",
			carpool.ID, carpool.Script.Name, carpool.RequiredPlayers-carpool.CurrentPlayers)

		tx := database.DB.Begin()

		for _, p := range carpool.Players {
			if p.DepositPaid && carpool.DepositAmount > 0 {
				log.Printf("[Refund][Auto] 玩家 %s 押金 %.2f 已退回", p.Name, carpool.DepositAmount)
				PushNotification(tx, p.Name, carpool.ID, "deposit_refunded",
					fmt.Sprintf("拼车超时自动解散，押金 %.2f 元已退回", carpool.DepositAmount))
			} else {
				PushNotification(tx, p.Name, carpool.ID, "carpool_cancelled",
					fmt.Sprintf("《%s》距开局不足1小时仍未满员，已自动解散", carpool.Script.Name))
			}
		}

		for _, w := range carpool.Waitlist {
			if w.Status == "waiting" {
				tx.Model(&w).Update("status", "cancelled")
				PushNotification(tx, w.Name, carpool.ID, "carpool_cancelled",
					fmt.Sprintf("《%s》已自动解散，候补资格失效", carpool.Script.Name))
			}
		}

		tx.Model(&carpool).Update("status", "cancelled")
		tx.Commit()
	}

	log.Printf("[Scheduler] 扫描完成，处理 %d 个拼车", len(carpools))
}

func StartScheduler() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		log.Println("[Scheduler] 超时自动解散任务已启动（每分钟执行）")
		for range ticker.C {
			CheckTimeoutAndCancel()
		}
	}()
}
