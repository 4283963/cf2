package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
)

const BaseURL = "http://localhost:8081/api"

type Carpool struct {
	ID              uint   `json:"id"`
	CurrentPlayers  int    `json:"current_players"`
	RequiredPlayers int    `json:"required_players"`
	Status          string `json:"status"`
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func httpPost(path string, body interface{}, out interface{}) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return err
	}
	resp, err := http.Post(BaseURL+path, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		if len(data) > 0 {
			json.Unmarshal(data, &errResp)
		}
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

func httpGet(path string, out interface{}) error {
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

func main() {
	fmt.Println("=== 高并发拼车抢座位 Bug 复现 & 验证测试 ===")
	fmt.Println()

	scriptID := 3
	fmt.Printf("[Step 1] 创建新拼车（剧本 #%d，窗边的女人，6人本）\n", scriptID)

	var created Carpool
	must(httpPost("/carpools", map[string]interface{}{
		"script_id":    scriptID,
		"host_name":    "车主小王",
		"host_contact": "13800000000",
	}, &created))
	fmt.Printf("        ✓ 拼车 #%d 创建成功：1/%d 人\n", created.ID, created.RequiredPlayers)

	fmt.Println()
	fmt.Println("[Step 2] 顺序加入 4 个玩家，做成 5缺1 的状态...")
	for i := 1; i <= 4; i++ {
		name := fmt.Sprintf("玩家%d", i)
		var res Carpool
		err := httpPost("/carpools/join", map[string]interface{}{
			"carpool_id": created.ID,
			"name":       name,
			"contact":    "",
		}, &res)
		if err != nil {
			panic(fmt.Sprintf("%s 加入失败: %v", name, err))
		}
		fmt.Printf("        ✓ %s 加入：%d/%d 人\n", name, res.CurrentPlayers, res.RequiredPlayers)
	}

	var before Carpool
	must(httpGet(fmt.Sprintf("/carpools?status=all"), nil))
	must(httpGet(fmt.Sprintf("/carpools"), nil))

	var allCarpools []Carpool
	must(httpGet("/carpools", &allCarpools))
	for _, c := range allCarpools {
		if c.ID == created.ID {
			before = c
			break
		}
	}
	fmt.Printf("        当前状态：%d/%d（还差 %d 人满员）\n",
		before.CurrentPlayers, before.RequiredPlayers,
		before.RequiredPlayers-before.CurrentPlayers)

	fmt.Println()
	fmt.Println("[Step 3] 同时启动 10 个 goroutine 并发抢最后 1 个座位！")
	var (
		successCount int32
		failCount    int32
		wg           sync.WaitGroup
	)

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("并发玩家%d", idx)
			var res Carpool
			err := httpPost("/carpools/join", map[string]interface{}{
				"carpool_id": created.ID,
				"name":       name,
				"contact":    "",
			}, &res)
			if err == nil {
				atomic.AddInt32(&successCount, 1)
				fmt.Printf("        [Goroutine %2d] ✓ %s 加入成功！%d/%d\n",
					idx, name, res.CurrentPlayers, res.RequiredPlayers)
			} else {
				atomic.AddInt32(&failCount, 1)
				fmt.Printf("        [Goroutine %2d] ✗ %s 被拒绝：%v\n",
					idx, name, err)
			}
		}(i)
	}
	wg.Wait()

	fmt.Println()
	fmt.Println("[Step 4] 从数据库重新查询最终状态，验证数据一致性...")
	var finalList []struct {
		ID              uint   `json:"id"`
		CurrentPlayers  int    `json:"current_players"`
		RequiredPlayers int    `json:"required_players"`
		Status          string `json:"status"`
		Players         []struct {
			Name string `json:"name"`
		} `json:"players"`
	}
	must(httpGet("/carpools", &finalList))

	var final struct {
		ID              uint   `json:"id"`
		CurrentPlayers  int    `json:"current_players"`
		RequiredPlayers int    `json:"required_players"`
		Status          string `json:"status"`
		Players         []struct {
			Name string `json:"name"`
		} `json:"players"`
	}
	for _, c := range finalList {
		if c.ID == created.ID {
			final = c
			break
		}
	}

	fmt.Println()
	fmt.Println("========== 测试结果 ==========")
	fmt.Printf("并发加入成功: %d 人（预期 1 人）\n", successCount)
	fmt.Printf("并发加入失败: %d 人（预期 9 人）\n", failCount)
	fmt.Printf("数据库最终人数: %d / %d（上限 %d 人）\n",
		final.CurrentPlayers, final.RequiredPlayers, final.RequiredPlayers)
	fmt.Printf("实际玩家列表长度: %d 人\n", len(final.Players))
	fmt.Printf("拼车状态: %s\n", final.Status)

	fmt.Println()
	pass := true
	if final.CurrentPlayers > final.RequiredPlayers {
		fmt.Printf("❌ 严重失败：current_players(%d) > required_players(%d)，超卖 Bug 仍存在！\n",
			final.CurrentPlayers, final.RequiredPlayers)
		pass = false
	} else {
		fmt.Printf("✅ 人数未超卖：%d ≤ %d\n", final.CurrentPlayers, final.RequiredPlayers)
	}

	if len(final.Players) > final.RequiredPlayers {
		fmt.Printf("❌ 严重失败：实际玩家数(%d) > required_players(%d)\n",
			len(final.Players), final.RequiredPlayers)
		pass = false
	} else {
		fmt.Printf("✅ 实际玩家数正常：%d ≤ %d\n", len(final.Players), final.RequiredPlayers)
	}

	if final.CurrentPlayers != len(final.Players) {
		fmt.Printf("⚠️  警告：current_players(%d) != 实际玩家数(%d)，计数不一致\n",
			final.CurrentPlayers, len(final.Players))
		pass = false
	} else {
		fmt.Printf("✅ 计数字段与实际数据一致\n")
	}

	if successCount != 1 {
		fmt.Printf("⚠️  并发成功数：%d（严格上只能有 1 人成功，若为 0 可能是顺序加入了 5 人）\n", successCount)
	}
	if failCount != 9 {
		fmt.Printf("⚠️  并发失败数：%d（应有 9 人被拒绝）\n", failCount)
	}

	if final.Status != "full" {
		fmt.Printf("❌ 状态应为 'full'，实际为 '%s'\n", final.Status)
		pass = false
	} else {
		fmt.Printf("✅ 拼车状态正确：已满员\n")
	}

	fmt.Println()
	if pass {
		fmt.Println("🎉🎉🎉 全部通过！并发抢座位 Bug 已彻底修复！")
	} else {
		fmt.Println("❌ 测试未通过，仍有问题")
	}
}
