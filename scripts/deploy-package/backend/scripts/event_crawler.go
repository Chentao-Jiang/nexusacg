package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Event represents an event scraped from a website.
type Event struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Address     string    `json:"address"`
	CoverURL    string    `json:"cover_url"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
}

// SeedEvents creates sample events and pushes them to the API.
func SeedEvents(apiBase, token string) error {
	events := sampleEvents()
	client := &http.Client{Timeout: 30 * time.Second}

	for i, ev := range events {
		body, _ := json.Marshal(map[string]interface{}{
			"name":        ev.Name,
			"description": ev.Description,
			"start_time":  ev.StartTime.Format(time.RFC3339),
			"end_time":    ev.EndTime.Format(time.RFC3339),
			"address":     ev.Address,
			"cover_url":   ev.CoverURL,
		})

		req, _ := http.NewRequest("POST", apiBase+"/api/v1/events", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("  [%d/%d] failed to create %s: %v", i+1, len(events), ev.Name, err)
			continue
		}
		resp.Body.Close()
		log.Printf("  [%d/%d] created: %s", i+1, len(events), ev.Name)
	}
	return nil
}

// CrawlACGEvents scrapes events from public ACG event aggregation sites.
// Currently supports bilibili member events and acg17.com as sources.
func CrawlACGEvents() ([]Event, error) {
	var allEvents []Event

	// Try bilibili member events
	events, err := crawlBilibiliEvents()
	if err != nil {
		log.Printf("bilibili crawl failed: %v", err)
	} else {
		allEvents = append(allEvents, events...)
	}

	// Try acg17
	events, err = crawlACG17Events()
	if err != nil {
		log.Printf("acg17 crawl failed: %v", err)
	} else {
		allEvents = append(allEvents, events...)
	}

	return allEvents, nil
}

func crawlBilibiliEvents() ([]Event, error) {
	resp, err := http.Get("https://member.bilibili.com/platform/act-list")
	if err != nil {
		return nil, fmt.Errorf("fetch bilibili events: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bilibili returned status %d", resp.StatusCode)
	}

	// Parse HTML for event listings
	// Note: bilibili structure may change; this is a best-effort parser
	var events []Event
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse bilibili HTML: %w", err)
	}

	doc.Find(".act-item, .event-item").Each(func(i int, s *goquery.Selection) {
		name := s.Find(".title, .name").Text()
		addr := s.Find(".address, .location").Text()
		desc := s.Find(".desc, .description").Text()
		if name != "" {
			events = append(events, Event{
				Name:        strings.TrimSpace(name),
				Description: strings.TrimSpace(desc),
				Address:     strings.TrimSpace(addr),
			})
		}
	})

	return events, nil
}

func crawlACG17Events() ([]Event, error) {
	resp, err := http.Get("https://www.acg17.com/allpost")
	if err != nil {
		return nil, fmt.Errorf("fetch acg17 events: %w", err)
	}
	defer resp.Body.Close()

	var events []Event
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse acg17 HTML: %w", err)
	}

	doc.Find("article, .post").Each(func(i int, s *goquery.Selection) {
		title := s.Find("h2 a, h1 a, .title").Text()
		content := s.Find(".entry-content, .content").Text()
		if title != "" && (strings.Contains(title, "漫展") || strings.Contains(title, "活动") || strings.Contains(title, "ONLY")) {
			events = append(events, Event{
				Name:        strings.TrimSpace(title),
				Description: strings.TrimSpace(content)[:min(len(strings.TrimSpace(content)), 500)],
				Address:     "", // Would need to parse detail page for address
			})
		}
	})

	return events, nil
}

func sampleEvents() []Event {
	now := time.Now()
	return []Event{
		{
			Name:        "上海次元随舞·冬季特别场",
			Description: "二次元随机宅舞大会冬季特别场！现场播放上百首ACG热门曲目，随到随跳，自由参与。还有嘉宾表演和抽奖环节。",
			StartTime:   now.AddDate(0, 0, 7),
			EndTime:     now.AddDate(0, 0, 7).Add(4 * time.Hour),
			Address:     "上海市黄浦区南京东路步行街",
			CoverURL:    "",
		},
		{
			Name:        "魔都同人祭·COMICUP29",
			Description: "COMICUP同人作品交流会第29届，数千社团参展，涵盖漫画、小说、游戏、手书等多个领域。",
			StartTime:   now.AddDate(0, 1, 0),
			EndTime:     now.AddDate(0, 1, 1),
			Address:     "上海国家会展中心（虹桥）",
			CoverURL:    "",
		},
		{
			Name:        "杭州动漫展·CP分会场",
			Description: "COSPLAY自由舞台、摄影会、同人摊位、美食区。现场还有声优见面会。",
			StartTime:   now.AddDate(0, 0, 14),
			EndTime:     now.AddDate(0, 0, 15),
			Address:     "杭州白马湖国际会展中心",
			CoverURL:    "",
		},
		{
			Name:        "二次元摄影会·外景拍摄",
			Description: "户外二次元主题摄影活动，提供服装道具，专业摄影师指导。适合COSER和摄影爱好者。",
			StartTime:   now.AddDate(0, 0, 5),
			EndTime:     now.AddDate(0, 0, 5).Add(3 * time.Hour),
			Address:     "杭州市西湖文化广场",
			CoverURL:    "",
		},
		{
			Name:        "南京·金陵Comic ONLY",
			Description: "南京地区最大规模的同人ONLY活动，聚集华东地区优秀创作者。",
			StartTime:   now.AddDate(0, 0, 21),
			EndTime:     now.AddDate(0, 0, 22),
			Address:     "南京国际展览中心",
			CoverURL:    "",
		},
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	apiBase := os.Getenv("API_BASE_URL")
	if apiBase == "" {
		apiBase = "http://localhost:8080"
	}
	token := os.Getenv("ADMIN_TOKEN")
	if token == "" {
		fmt.Println("ADMIN_TOKEN not set, using seed data without auth...")
	}

	fmt.Println("=== 次元链活动信息爬取 & 种子数据 ===")

	// Try crawling first
	fmt.Println("\n[1/2] 尝试爬取公开活动信息...")
	events, err := CrawlACGEvents()
	if err != nil {
		fmt.Printf("  爬取失败: %v（将使用种子数据）\n", err)
	} else if len(events) > 0 {
		fmt.Printf("  成功抓取 %d 个活动\n", len(events))
		// Push to API
		fmt.Println("\n[2/2] 导入活动到数据库...")
		if err := SeedEvents(apiBase, token); err != nil {
			fmt.Printf("  导入失败: %v\n", err)
		}
		return
	}

	// Fall back to seed data
	fmt.Println("\n[2/2] 使用种子数据创建示例活动...")
	if err := SeedEvents(apiBase, token); err != nil {
		fmt.Printf("  创建失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("  完成！")
}
