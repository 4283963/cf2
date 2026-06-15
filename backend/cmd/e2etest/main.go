package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const BaseURL = "http://localhost:8081/api"

func post(path string, body interface{}, out interface{}) error {
	jsonData, _ := json.Marshal(body)
	resp, err := http.Post(BaseURL+path, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		var errResp struct{ Error string }
		json.Unmarshal(data, &errResp)
		if errResp.Error != "" {
			return fmt.Errorf("[%d] %s", resp.StatusCode, errResp.Error)
		}
		return fmt.Errorf("[%d] %s", resp.StatusCode, string(data))
	}
	if out != nil && len(data) > 0 {
		return json.Unmarshal(data, out)
	}
	return nil
}

func get(path string, out interface{}) error {
	resp, err := http.Get(BaseURL + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("[%d] %s", resp.StatusCode, string(data))
	}
	if out != nil && len(data) > 0 {
		return json.Unmarshal(data, out)
	}
	return nil
}

type Carpool struct {
	ID             uint   `json:"id"`
	CurrentPlayers int    `json:"current_players"`
	RequiredPlayers int   `json:"required_players"`
	Status         string `json:"status"`
}

func main() {
	fmt.Println("========== 候补 & 退车自动补位 端到端测试 ==========")
	fmt.Println()

	scriptID := 1 // 古木吟 6人本
	fmt.Printf("[Step 1] 车主小王创建《古木吟》拼车（剧本 #%d，6 人本）\n", scriptID)
	var created Carpool
	must(post("/carpools", map[string]interface{}{
		"script_id":    scriptID,
		"host_name":    "小王",
		"host_contact": "13800000000",
	}, &created))
	carpoolID := created.ID
	fmt.Printf("        ✓ 拼车 #%d 创建成功：1/%d\n", carpoolID, created.RequiredPlayers)

	fmt.Println()
	fmt.Println("[Step 2] 让 5 个玩家顺序加入，做成 5缺1 的状态")
	players := []string{"小李", "小张", "小赵", "小钱", "小孙"}
	for i, name := range players[:4] {
		var res Carpool
		must(post("/carpools/join", map[string]interface{}{
			"carpool_id": carpoolID,
			"name":       name,
		}, &res))
		fmt.Printf("        ✓ %s 加入：%d/%d\n", name, res.CurrentPlayers, res.RequiredPlayers)
		_ = i
	}

	fmt.Println()
	fmt.Println("[Step 3] 让 3 个玩家分别申请候补（候补 A、候补 B、候补 C）")
	waiters := []string{"候补A", "候补B", "候补C"}
	for i, name := range waiters {
		var res struct {
			Message  string `json:"message"`
			Position int    `json:"position"`
		}
		must(post("/waitlist", map[string]interface{}{
			"carpool_id": carpoolID,
			"name":       name,
		}, &res))
		fmt.Printf("        ✓ %s %s\n", name, res.Message)
		_ = i
	}

	fmt.Println()
	fmt.Println("[Step 4] 让第 5 个玩家（小孙）加入，使拼车变满")
	{
		var res Carpool
		must(post("/carpools/join", map[string]interface{}{
			"carpool_id": carpoolID,
			"name":       players[4],
		}, &res))
		fmt.Printf("        ✓ %s 加入后：%d/%d，状态=%s\n",
			players[4], res.CurrentPlayers, res.RequiredPlayers, res.Status)
	}

	fmt.Println()
	fmt.Println("[Step 5] 让玩家「小赵」退车，触发自动补位")
	var leaveRes struct {
		Message  string `json:"message"`
		Promoted *struct {
			Name     string `json:"name"`
			Position int    `json:"position"`
		} `json:"promoted"`
		Carpool Carpool `json:"carpool"`
	}
	must(post("/carpools/leave", map[string]interface{}{
		"carpool_id": carpoolID,
		"name":       "小赵",
	}, &leaveRes))
	fmt.Printf("        ✓ 小赵退车：%s\n", leaveRes.Message)
	if leaveRes.Promoted != nil {
		fmt.Printf("        🎉 自动补位成功：%s（第 %d 顺位）已补入！\n",
			leaveRes.Promoted.Name, leaveRes.Promoted.Position)
	} else {
		fmt.Printf("        ❌ 未触发自动补位（Bug！）\n")
	}
	fmt.Printf("        当前状态：%d/%d，拼车状态=%s\n",
		leaveRes.Carpool.CurrentPlayers, leaveRes.Carpool.RequiredPlayers, leaveRes.Carpool.Status)

	fmt.Println()
	fmt.Println("[Step 6] 检查「候补A」的通知")
	var notifs []struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}
	must(get("/notifications?user="+waiters[0], &notifs))
	fmt.Printf("        候补A 共收到 %d 条通知：\n", len(notifs))
	for i, n := range notifs {
		fmt.Printf("          [%d] [%s] %s\n", i+1, n.Type, n.Message)
	}

	fmt.Println()
	fmt.Println("[Step 7] 检查「小王」（车主）的通知")
	must(get("/notifications?user=小王", &notifs))
	fmt.Printf("        车主小王 共收到 %d 条通知：\n", len(notifs))
	for i, n := range notifs {
		fmt.Printf("          [%d] [%s] %s\n", i+1, n.Type, n.Message)
	}

	fmt.Println()
	fmt.Println("[Step 8] 验证「小赵」（退车玩家）的押金退还通知")
	must(get("/notifications?user=小赵", &notifs))
	fmt.Printf("        小赵 共收到 %d 条通知：\n", len(notifs))
	for i, n := range notifs {
		fmt.Printf("          [%d] [%s] %s\n", i+1, n.Type, n.Message)
	}

	fmt.Println()
	fmt.Println("[Step 9] 再让一个人退车，验证候补B的补位")
	must(post("/carpools/leave", map[string]interface{}{
		"carpool_id": carpoolID,
		"name":       "小钱",
	}, &leaveRes))
	fmt.Printf("        ✓ 小钱退车\n")
	if leaveRes.Promoted != nil {
		fmt.Printf("        🎉 自动补位成功：%s（第 %d 顺位）已补入！\n",
			leaveRes.Promoted.Name, leaveRes.Promoted.Position)
	} else {
		fmt.Printf("        ❌ 未触发自动补位（Bug！）\n")
	}
	fmt.Printf("        当前状态：%d/%d，拼车状态=%s\n",
		leaveRes.Carpool.CurrentPlayers, leaveRes.Carpool.RequiredPlayers, leaveRes.Carpool.Status)

	time.Sleep(1200 * time.Millisecond)
	fmt.Println()
	fmt.Println("========== 测试完成 ==========")

	passCount, failCount := 0, 0
	if leaveRes.Promoted != nil && leaveRes.Promoted.Name == "候补B" {
		passCount++
		fmt.Printf("✅ 候补B补位正确\n")
	} else {
		failCount++
		fmt.Printf("❌ 候补B未正确补位\n")
	}
	if leaveRes.Carpool.CurrentPlayers == leaveRes.Carpool.RequiredPlayers {
		passCount++
		fmt.Printf("✅ 补位后人数恢复到满员：%d/%d\n",
			leaveRes.Carpool.CurrentPlayers, leaveRes.Carpool.RequiredPlayers)
	} else {
		failCount++
		fmt.Printf("❌ 补位后人数不对：%d/%d\n",
			leaveRes.Carpool.CurrentPlayers, leaveRes.Carpool.RequiredPlayers)
	}
	if leaveRes.Carpool.Status == "full" {
		passCount++
		fmt.Printf("✅ 补位后状态=full\n")
	} else {
		failCount++
		fmt.Printf("❌ 补位后状态不是full：%s\n", leaveRes.Carpool.Status)
	}

	fmt.Println()
	fmt.Printf("通过: %d / 失败: %d\n", passCount, failCount)
	if failCount == 0 {
		fmt.Println("🎉🎉🎉 候补自动补位全部测试通过！")
	} else {
		fmt.Println("❌ 存在测试失败")
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
