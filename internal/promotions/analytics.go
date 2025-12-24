package promotions

import (
	"encoding/json"
	"net/http"
	"strconv"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"

	"github.com/gorilla/mux"
)

// GetPromotionAnalytics handles getting detailed analytics for a promotion
func (h *PromotionHandler) GetPromotionAnalytics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get promotion ID from URL
	vars := mux.Vars(r)
	promotionID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid promotion ID")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get promotion
	var promotion models.Promotion
	if err := h.db.First(&promotion, promotionID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "promotion not found")
		return
	}

	// Check authorization
	if promotion.OrganizerID != nil {
		var organizer models.Organizer
		if err := h.db.Where("id = ? AND account_id = ?", *promotion.OrganizerID, user.AccountID).First(&organizer).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
			return
		}
	}

	// Get unique users count
	var uniqueUsers int64
	h.db.Model(&models.PromotionUsage{}).
		Where("promotion_id = ?", promotionID).
		Distinct("account_id").
		Count(&uniqueUsers)

	// Calculate average order size
	var avgOrderSize models.Money
	if promotion.UsageCount > 0 {
		avgOrderSize = models.Money(int64(promotion.TotalRevenue) / int64(promotion.UsageCount))
	}

	// Calculate ROI (Return on Investment)
	roi := 0.0
	if promotion.TotalDiscount > 0 {
		roi = (float64(promotion.TotalRevenue) / float64(promotion.TotalDiscount)) * 100
	}

	// Get usage by day
	usageByDay := h.getUsageByDay(uint(promotionID))

	conversionRate := 0.0
	if promotion.ConversionRate != nil {
		conversionRate = *promotion.ConversionRate
	}

	response := PromotionAnalyticsResponse{
		PromotionID:      promotion.ID,
		Code:             promotion.Code,
		TotalUsage:       int64(promotion.UsageCount),
		UniqueUsers:      uniqueUsers,
		TotalRevenue:     promotion.TotalRevenue,
		TotalDiscount:    promotion.TotalDiscount,
		AverageOrderSize: avgOrderSize,
		ConversionRate:   conversionRate,
		ROI:              roi,
		UsageByDay:       usageByDay,
	}

	json.NewEncoder(w).Encode(response)
}

// GetPromotionROI handles getting return on investment for a promotion
func (h *PromotionHandler) GetPromotionROI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get promotion ID from URL
	vars := mux.Vars(r)
	promotionID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid promotion ID")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get promotion
	var promotion models.Promotion
	if err := h.db.First(&promotion, promotionID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "promotion not found")
		return
	}

	// Check authorization
	if promotion.OrganizerID != nil {
		var organizer models.Organizer
		if err := h.db.Where("id = ? AND account_id = ?", *promotion.OrganizerID, user.AccountID).First(&organizer).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
			return
		}
	}

	// Calculate ROI
	roi := 0.0
	netGain := int64(0)
	if promotion.TotalDiscount > 0 {
		netGain = int64(promotion.TotalRevenue - promotion.TotalDiscount)
		roi = (float64(netGain) / float64(promotion.TotalDiscount)) * 100
	}

	// Calculate cost per acquisition
	costPerAcquisition := 0.0
	averageDiscount := 0.0
	if promotion.UsageCount > 0 {
		costPerAcquisition = float64(promotion.TotalDiscount) / float64(promotion.UsageCount)
		averageDiscount = float64(promotion.TotalDiscount) / float64(promotion.UsageCount)
	}

	response := map[string]interface{}{
		"promotion_id":         promotion.ID,
		"code":                 promotion.Code,
		"total_revenue":        promotion.TotalRevenue,
		"total_discount":       promotion.TotalDiscount,
		"net_gain":             netGain,
		"roi_percentage":       roi,
		"usage_count":          promotion.UsageCount,
		"cost_per_acquisition": costPerAcquisition,
		"average_discount":     averageDiscount,
	}

	json.NewEncoder(w).Encode(response)
}

// GetConversionMetrics handles getting conversion metrics
func (h *PromotionHandler) GetConversionMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get promotion ID from URL
	vars := mux.Vars(r)
	promotionID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid promotion ID")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get promotion
	var promotion models.Promotion
	if err := h.db.First(&promotion, promotionID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "promotion not found")
		return
	}

	// Check authorization
	if promotion.OrganizerID != nil {
		var organizer models.Organizer
		if err := h.db.Where("id = ? AND account_id = ?", *promotion.OrganizerID, user.AccountID).First(&organizer).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
			return
		}
	}

	// Get unique users
	var uniqueUsers int64
	h.db.Model(&models.PromotionUsage{}).
		Where("promotion_id = ?", promotionID).
		Distinct("account_id").
		Count(&uniqueUsers)

	// Calculate repeat usage rate
	repeatRate := 0.0
	if uniqueUsers > 0 {
		repeatRate = (float64(promotion.UsageCount) / float64(uniqueUsers)) - 1.0
	}

	response := map[string]interface{}{
		"promotion_id":      promotion.ID,
		"code":              promotion.Code,
		"total_usage":       promotion.UsageCount,
		"unique_users":      uniqueUsers,
		"repeat_usage_rate": repeatRate,
		"conversion_rate":   promotion.ConversionRate,
	}

	json.NewEncoder(w).Encode(response)
}

// GetRevenueImpact handles getting revenue impact analysis
func (h *PromotionHandler) GetRevenueImpact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get promotion ID from URL
	vars := mux.Vars(r)
	promotionID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid promotion ID")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get promotion
	var promotion models.Promotion
	if err := h.db.First(&promotion, promotionID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "promotion not found")
		return
	}

	// Check authorization
	if promotion.OrganizerID != nil {
		var organizer models.Organizer
		if err := h.db.Where("id = ? AND account_id = ?", *promotion.OrganizerID, user.AccountID).First(&organizer).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
			return
		}
	}

	// Calculate metrics
	revenueWithPromo := promotion.TotalRevenue
	revenueWithoutPromo := promotion.TotalRevenue + promotion.TotalDiscount
	discountRate := 0.0
	if revenueWithoutPromo > 0 {
		discountRate = (float64(promotion.TotalDiscount) / float64(revenueWithoutPromo)) * 100
	}

	averageDiscountPerOrder := 0.0
	if promotion.UsageCount > 0 {
		averageDiscountPerOrder = float64(promotion.TotalDiscount) / float64(promotion.UsageCount)
	}

	response := map[string]interface{}{
		"promotion_id":               promotion.ID,
		"code":                       promotion.Code,
		"revenue_with_promo":         revenueWithPromo,
		"revenue_without_promo":      revenueWithoutPromo,
		"total_discount":             promotion.TotalDiscount,
		"discount_rate":              discountRate,
		"net_revenue_impact":         revenueWithPromo,
		"average_discount_per_order": averageDiscountPerOrder,
	}

	json.NewEncoder(w).Encode(response)
}

// GetOrganizerPromotionStats handles getting overall promotion stats for organizer
func (h *PromotionHandler) GetOrganizerPromotionStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get organizer record for this user's account
	var organizer models.Organizer
	if err := h.db.Where("account_id = ?", user.AccountID).First(&organizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "organizer profile not found")
		return
	}

	// Count promotions by status
	var totalPromotions, activePromotions, expiredPromotions int64
	h.db.Model(&models.Promotion{}).Where("organizer_id = ?", organizer.ID).Count(&totalPromotions)
	h.db.Model(&models.Promotion{}).Where("organizer_id = ? AND status = ?", organizer.ID, models.PromotionActive).Count(&activePromotions)
	h.db.Model(&models.Promotion{}).Where("organizer_id = ? AND status = ?", organizer.ID, models.PromotionExpired).Count(&expiredPromotions)

	// Get total usage
	var totalUsage int64
	h.db.Model(&models.PromotionUsage{}).
		Joins("JOIN promotions ON promotions.id = promotion_usages.promotion_id").
		Where("promotions.organizer_id = ?", organizer.ID).
		Count(&totalUsage)

	// Get revenue and discount totals
	var stats struct {
		TotalRevenue  int64
		TotalDiscount int64
	}
	h.db.Model(&models.Promotion{}).
		Select("SUM(total_revenue) as total_revenue, SUM(total_discount) as total_discount").
		Where("organizer_id = ?", organizer.ID).
		Scan(&stats)

	// Calculate averages
	avgDiscount := 0.0
	conversionRate := 0.0
	if totalUsage > 0 {
		avgDiscount = float64(stats.TotalDiscount) / float64(totalUsage)
	}

	response := PromotionStatsResponse{
		TotalPromotions:   totalPromotions,
		ActivePromotions:  activePromotions,
		ExpiredPromotions: expiredPromotions,
		TotalUsage:        totalUsage,
		TotalRevenue:      models.Money(stats.TotalRevenue),
		TotalDiscount:     models.Money(stats.TotalDiscount),
		AverageDiscount:   avgDiscount,
		ConversionRate:    conversionRate,
	}

	json.NewEncoder(w).Encode(response)
}

// Helper function to get usage by day
func (h *PromotionHandler) getUsageByDay(promotionID uint) []DailyUsage {
	var results []struct {
		Date     string
		Count    int64
		Revenue  int64
		Discount int64
	}

	// Get usage grouped by day for last 30 days
	h.db.Model(&models.PromotionUsage{}).
		Select("DATE(used_at) as date, COUNT(*) as count, SUM(final_amount) as revenue, SUM(discount_amount) as discount").
		Where("promotion_id = ? AND used_at >= ?", promotionID, time.Now().AddDate(0, 0, -30)).
		Group("DATE(used_at)").
		Order("date ASC").
		Scan(&results)

	dailyUsage := make([]DailyUsage, len(results))
	for i, result := range results {
		dailyUsage[i] = DailyUsage{
			Date:     result.Date,
			Count:    result.Count,
			Revenue:  result.Revenue,
			Discount: result.Discount,
		}
	}

	return dailyUsage
}
