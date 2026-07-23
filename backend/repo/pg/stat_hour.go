package pg

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/samber/lo"

	v1 "github.com/chaitin/panda-wiki/api/stat/v1"
	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/utils"
)

func (r *StatRepository) GetConversationCountOneHour(ctx context.Context, kbID string) (int64, error) {
	var conversationCount int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Conversation{}).
		Where("kb_id = ?", kbID).
		Where("created_at >= ? AND created_at < ?", utils.GetTimeHourOffset(-1), utils.GetTimeHourOffset(0)).
		Count(&conversationCount).Error; err != nil {
		return conversationCount, err
	}
	return conversationCount, nil
}

func (r *StatRepository) GetStatPageOneHour(ctx context.Context, kbID string) (*domain.StatPageHour, error) {
	var statPageHour domain.StatPageHour
	err := r.db.WithContext(ctx).Table("stat_pages").
		Select(`
			COUNT(DISTINCT ip) as ip_count,
			COUNT(DISTINCT session_id) as session_count,
			COUNT(*) as page_visit_count
		`).
		Where("created_at >= ? AND created_at < ?", utils.GetTimeHourOffset(-1), utils.GetTimeHourOffset(0)).
		Where("kb_id = ?", kbID).
		Find(&statPageHour).Error

	if err != nil {
		return nil, err
	}
	return &statPageHour, nil
}

func (r *StatRepository) GetGeCountOneHour(ctx context.Context, kbID string) (map[string]int64, error) {
	key := fmt.Sprintf("geo:%s:%s", kbID, time.Now().Add(-time.Duration(1)*time.Hour).Format("2006-01-02-15"))
	values, err := r.cache.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	geoCount := make(map[string]int64)
	for field, value := range values {
		valueInt, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse geo count failed: %w", err)
		}
		geoCount[field] += valueInt
	}

	return geoCount, nil
}

func (r *StatRepository) GetConversationDistributionOneHour(ctx context.Context, kbID string) (map[string]int64, error) {
	var cds []domain.ConversationDistribution
	if err := r.db.WithContext(ctx).
		Model(&domain.Conversation{}).
		Select("apps.type as app_type", "COUNT(*) as count").
		Joins("left join apps on apps.id=conversations.app_id").
		Where("conversations.kb_id = ?", kbID).
		Where("conversations.created_at >= ? AND conversations.created_at < ?", utils.GetTimeHourOffset(-1), utils.GetTimeHourOffset(0)).
		Group("apps.type").
		Find(&cds).Error; err != nil {
		return nil, err
	}

	if len(cds) == 0 {
		return make(map[string]int64), nil
	}

	dcCount := lo.SliceToMap(cds, func(cd domain.ConversationDistribution) (string, int64) {
		return strconv.Itoa(int(cd.AppType)), cd.Count
	})

	return dcCount, nil
}

func (r *StatRepository) GetHotRefererHostOneHour(ctx context.Context, kbID string) (map[string]int64, error) {
	var hotRefererHosts []*domain.HotRefererHost
	if err := r.db.WithContext(ctx).Model(&domain.StatPage{}).
		Where("kb_id = ?", kbID).
		Where("created_at >= ? AND created_at < ?", utils.GetTimeHourOffset(-1), utils.GetTimeHourOffset(0)).
		Group("referer_host").
		Select("referer_host, COUNT(*) as count").
		Order("count DESC").
		Limit(10).
		Find(&hotRefererHosts).Error; err != nil {
		return nil, err
	}

	if len(hotRefererHosts) == 0 {
		return make(map[string]int64), nil
	}

	refererHostCount := lo.SliceToMap(hotRefererHosts, func(item *domain.HotRefererHost) (string, int64) {
		return item.RefererHost, item.Count
	})

	return refererHostCount, nil
}

func (r *StatRepository) GetHotRefererHostsByHour(ctx context.Context, kbID string, startHour int64) (map[string]int64, error) {
	// 查询实时数据
	var hotRefererHosts []*domain.HotRefererHost
	if err := r.db.WithContext(ctx).Model(&domain.StatPage{}).
		Where("kb_id = ?", kbID).
		Where("referer_host != '' ").
		Where("created_at > ?", utils.GetTimeHourOffset(-24)).
		Group("referer_host").
		Select("referer_host, COUNT(*) as count").
		Order("count DESC").
		Limit(10).
		Find(&hotRefererHosts).Error; err != nil {
		return nil, err
	}

	// 查询小时统计表中的聚合数据
	statPageHours := make([]domain.StatPageHour, 0)
	if err := r.db.WithContext(ctx).Model(&domain.StatPageHour{}).
		Select("hot_referer_host").
		Where("kb_id = ?", kbID).
		Where("hour >= ? and hour < ?", utils.GetTimeHourOffset(-startHour), utils.GetTimeHourOffset(-24)).
		Find(&statPageHours).Error; err != nil {
		return nil, err
	}

	// 聚合小时统计数据
	refererHostCountMap := make(map[string]int64)
	for i := range statPageHours {
		for k, v := range statPageHours[i].HotRefererHost {
			refererHostCountMap[k] += v
		}
	}

	// 合并实时数据和聚合数据
	finalRefererHostCount := make(map[string]int64)
	for _, item := range hotRefererHosts {
		finalRefererHostCount[item.RefererHost] = item.Count
	}

	for host, count := range refererHostCountMap {
		if host != "" {
			finalRefererHostCount[host] += count
		}
	}

	return finalRefererHostCount, nil
}

func (r *StatRepository) CreateStatPageHour(ctx context.Context, statPageHour *domain.StatPageHour) error {
	return r.db.WithContext(ctx).Create(statPageHour).Error
}

// CheckStatPageHourExists 检查指定时间和知识库的小时统计数据是否已存在
func (r *StatRepository) CheckStatPageHourExists(ctx context.Context, kbID string, hour time.Time) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.StatPageHour{}).
		Where("kb_id = ? AND hour = ?", kbID, hour).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CleanupOldHourlyStats 清理90天前的小时统计数据
func (r *StatRepository) CleanupOldHourlyStats(ctx context.Context) error {
	return r.db.WithContext(ctx).Model(&domain.StatPageHour{}).
		Where("hour < NOW() - INTERVAL '90 days'").
		Delete(&domain.StatPageHour{}).Error
}

func (r *StatRepository) GetHotPagesOneHour(ctx context.Context, kbID string) (map[string]int64, error) {
	var hotPages []*domain.HotPage
	if err := r.db.WithContext(ctx).Model(&domain.StatPage{}).
		Where("kb_id = ?", kbID).
		Where("node_id != '' ").
		Where("scene = ?", domain.StatPageSceneNodeDetail).
		Where("created_at >= ? AND created_at < ?", utils.GetTimeHourOffset(-1), utils.GetTimeHourOffset(0)).
		Group("node_id").
		Select("node_id, COUNT(*) as count").
		Order("count DESC").
		Find(&hotPages).Error; err != nil {
		return nil, err
	}

	if len(hotPages) == 0 {
		return make(map[string]int64), nil
	}

	refererHostCount := lo.SliceToMap(hotPages, func(item *domain.HotPage) (string, int64) {
		return item.NodeID, item.Count
	})

	return refererHostCount, nil
}

func (r *StatRepository) GetHotPagesByHour(ctx context.Context, kbID string, startHour int64) (map[string]int64, error) {
	// 查询小时统计表中的聚合数据
	counts := make(map[string]int64)
	hotPageMaps := make([]domain.MapStrInt64, 0)
	if err := r.db.WithContext(ctx).Model(&domain.StatPageHour{}).
		Where("kb_id = ?", kbID).
		Where("hot_page != '{}'").
		Where("hour >= ? and hour < ?", utils.GetTimeHourOffset(-startHour), utils.GetTimeHourOffset(-24)).
		Pluck("hot_page", &hotPageMaps).Error; err != nil {
		return nil, err
	}
	for i := range hotPageMaps {
		for k, v := range hotPageMaps[i] {
			counts[k] += v
		}
	}

	return counts, nil
}

func (r *StatRepository) GetHotBrowsersOneHour(ctx context.Context, kbID string) (map[string]int64, error) {
	var browserCount []domain.BrowserCount

	query := r.db.WithContext(ctx).Model(&domain.StatPage{}).
		Where("kb_id = ?", kbID).
		Where("created_at >= ? AND created_at < ?", utils.GetTimeHourOffset(-1), utils.GetTimeHourOffset(0)).
		Group("browser_name").
		Select("browser_name as name, COUNT(*) as count")
	if err := query.Order("count DESC").Limit(10).Find(&browserCount).Error; err != nil {
		return nil, err
	}

	if len(browserCount) == 0 {
		return make(map[string]int64), nil
	}

	refererHostCount := lo.SliceToMap(browserCount, func(item domain.BrowserCount) (string, int64) {
		return item.Name, item.Count
	})

	return refererHostCount, nil
}

func (r *StatRepository) GetHotOSOneHour(ctx context.Context, kbID string) (map[string]int64, error) {
	var osCount []domain.BrowserCount

	query := r.db.WithContext(ctx).Model(&domain.StatPage{}).
		Where("kb_id = ?", kbID).
		Where("created_at >= ? AND created_at < ?", utils.GetTimeHourOffset(-1), utils.GetTimeHourOffset(0)).
		Group("browser_os").
		Select("browser_os as name, COUNT(*) as count")
	if err := query.Order("count DESC").Limit(10).Find(&osCount).Error; err != nil {
		return nil, err
	}

	if len(osCount) == 0 {
		return make(map[string]int64), nil
	}

	refererOSCount := lo.SliceToMap(osCount, func(item domain.BrowserCount) (string, int64) {
		return item.Name, item.Count
	})

	return refererOSCount, nil
}

func (r *StatRepository) GetStatPageCountByHour(ctx context.Context, kbID string, startHour int64) (*v1.StatCountResp, error) {
	var count v1.StatCountResp
	if err := r.db.WithContext(ctx).Model(&domain.StatPageHour{}).
		Select("SUM(ip_count) as ip_count, SUM(session_count) as session_count, SUM(page_visit_count) as page_visit_count, SUM(conversation_count) as conversation_count").
		Where("kb_id = ?", kbID).
		Where("hour >=  ? and hour < ?", utils.GetTimeHourOffset(-startHour), utils.GetTimeHourOffset(-24)).
		Scan(&count).Error; err != nil {
		return nil, err
	}
	return &count, nil
}

func (r *StatRepository) GetHotBrowsersByHour(ctx context.Context, kbID string, startHour int64) (*domain.HotBrowser, error) {

	var browserCount []domain.BrowserCount
	query := r.db.WithContext(ctx).Model(&domain.StatPage{}).
		Where("kb_id = ?", kbID).
		Where("created_at > ?", utils.GetTimeHourOffset(-24)).
		Where("browser_name != '' ").
		Group("browser_name").
		Select("browser_name as name, COUNT(*) as count")
	if err := query.Order("count DESC").Find(&browserCount).Error; err != nil {
		return nil, err
	}

	var osCount []domain.BrowserCount
	query = r.db.WithContext(ctx).Model(&domain.StatPage{}).
		Where("kb_id = ?", kbID).
		Where("created_at > ?", utils.GetTimeHourOffset(-24)).
		Where("browser_os != '' ").
		Group("browser_os").
		Select("browser_os as name, COUNT(*) as count")
	if err := query.Order("count DESC").Find(&osCount).Error; err != nil {
		return nil, err
	}

	statPageHours := make([]domain.StatPageHour, 0)
	if err := r.db.WithContext(ctx).Model(&domain.StatPageHour{}).
		Select("hot_os, hot_browser").
		Where("kb_id = ?", kbID).
		Where("hour >= ? and hour < ?", utils.GetTimeHourOffset(-startHour), utils.GetTimeHourOffset(-24)).
		Find(&statPageHours).Error; err != nil {
		return nil, err
	}
	hourBrowserCountMap := make(domain.MapStrInt64)
	hourOSCountMap := make(domain.MapStrInt64)

	for i := range statPageHours {
		for k, v := range statPageHours[i].HotOS {
			if k != "" {
				hourOSCountMap[k] += v
			}
		}

		for k, v := range statPageHours[i].HotBrowser {
			if k != "" {
				hourBrowserCountMap[k] += v
			}
		}
	}

	for i := range browserCount {
		hourBrowserCountMap[browserCount[i].Name] += browserCount[i].Count
	}

	for i := range osCount {
		hourOSCountMap[osCount[i].Name] += osCount[i].Count
	}

	browserCount = lo.MapToSlice(hourBrowserCountMap, func(k string, v int64) domain.BrowserCount {
		return domain.BrowserCount{
			Name:  k,
			Count: v,
		}
	})

	osCount = lo.MapToSlice(hourOSCountMap, func(k string, v int64) domain.BrowserCount {
		return domain.BrowserCount{
			Name:  k,
			Count: v,
		}
	})

	// Sort browserCount by count in descending order and take top 10
	sort.Slice(browserCount, func(i, j int) bool {
		return browserCount[i].Count > browserCount[j].Count
	})
	if len(browserCount) > 10 {
		browserCount = browserCount[:10]
	}

	// Sort osCount by count in descending order and take top 10
	sort.Slice(osCount, func(i, j int) bool {
		return osCount[i].Count > osCount[j].Count
	})
	if len(osCount) > 10 {
		osCount = osCount[:10]
	}

	return &domain.HotBrowser{
		Browser: browserCount,
		OS:      osCount,
	}, nil
}
