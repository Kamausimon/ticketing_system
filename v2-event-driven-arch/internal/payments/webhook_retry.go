package payments

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"ticketing_system/internal/models"
)

// WebhookRetryConfig defines retry behavior
type WebhookRetryConfig struct {
	MaxRetries     int           // Maximum number of retries
	InitialBackoff time.Duration // Initial backoff duration
	MaxBackoff     time.Duration // Maximum backoff duration
	BackoffFactor  float64       // Exponential backoff multiplier
}

// DefaultRetryConfig returns sensible defaults
func DefaultRetryConfig() WebhookRetryConfig {
	return WebhookRetryConfig{
		MaxRetries:     5,
		InitialBackoff: 1 * time.Minute,
		MaxBackoff:     4 * time.Hour,
		BackoffFactor:  3.0,
	}
}

// ProcessWebhookWithRetry processes a webhook with automatic retry logic
func (h *PaymentHandler) ProcessWebhookWithRetry(webhookLog *models.WebhookLog, retryConfig WebhookRetryConfig) error {
	// Parse the webhook event
	var event IntasendWebhookEvent
	if err := json.Unmarshal([]byte(webhookLog.Payload), &event); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// Attempt to process the webhook
	success, err := h.processIntasendWebhook(&event)

	now := time.Now()
	webhookLog.ProcessedAt = &now
	processingTime := int(time.Since(now).Milliseconds())
	webhookLog.ProcessingTime = &processingTime

	if success {
		webhookLog.Status = models.WebhookProcessed
		webhookLog.Success = true
		webhookLog.ErrorMessage = nil
		webhookLog.RetryCount = 0
		h.db.Save(webhookLog)
		return nil
	}

	// Log the error
	if err != nil {
		errMsg := err.Error()
		webhookLog.ErrorMessage = &errMsg
	}

	// Determine if we should retry
	if webhookLog.RetryCount >= retryConfig.MaxRetries {
		webhookLog.Status = models.WebhookFailed
		webhookLog.Success = false
		h.db.Save(webhookLog)
		return fmt.Errorf("webhook processing failed after %d retries: %w", retryConfig.MaxRetries, err)
	}

	// Schedule retry
	webhookLog.Status = models.WebhookRetrying
	webhookLog.RetryCount++
	webhookLog.LastRetryAt = &now
	h.db.Save(webhookLog)

	log.Printf("🔄 Webhook retry scheduled: attempt %d/%d (Event: %s)", webhookLog.RetryCount, retryConfig.MaxRetries, webhookLog.EventID)
	return nil
}

// RetryFailedWebhooksJob is a background job to retry failed webhooks
func (h *PaymentHandler) RetryFailedWebhooksJob(retryConfig WebhookRetryConfig) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		h.processFailedWebhooks(retryConfig)
	}
}

// processFailedWebhooks finds and retries eligible failed webhooks
func (h *PaymentHandler) processFailedWebhooks(retryConfig WebhookRetryConfig) {
	// Find webhooks that are retryable and have exceeded their backoff time
	var webhooks []models.WebhookLog
	now := time.Now()

	err := h.db.Where("status = ? AND retry_count < ? AND (last_retry_at IS NULL OR last_retry_at < ?)",
		models.WebhookFailed,
		retryConfig.MaxRetries,
		now.Add(-h.calculateBackoff(0, retryConfig))).
		Order("last_retry_at ASC").
		Limit(10).
		Find(&webhooks).Error

	if err != nil {
		log.Printf("❌ Error fetching failed webhooks: %v", err)
		return
	}

	for _, webhook := range webhooks {
		h.retryWebhookWithRecovery(&webhook, retryConfig)
	}
}

// retryWebhookWithRecovery retries a webhook with panic recovery
func (h *PaymentHandler) retryWebhookWithRecovery(webhookLog *models.WebhookLog, retryConfig WebhookRetryConfig) {
	defer func() {
		if r := recover(); r != nil {
			stackTrace := string(debug.Stack())
			webhookLog.StackTrace = &stackTrace
			errMsg := fmt.Sprintf("panic during webhook retry: %v", r)
			webhookLog.ErrorMessage = &errMsg
			webhookLog.Status = models.WebhookFailed
			h.db.Save(webhookLog)
			log.Printf("❌ Panic in webhook retry (Event: %s): %v\nStack: %s", webhookLog.EventID, r, stackTrace)
		}
	}()

	// Parse event
	var event IntasendWebhookEvent
	if err := json.Unmarshal([]byte(webhookLog.Payload), &event); err != nil {
		errMsg := fmt.Sprintf("failed to parse webhook payload: %v", err)
		webhookLog.ErrorMessage = &errMsg
		webhookLog.Status = models.WebhookFailed
		h.db.Save(webhookLog)
		return
	}

	// Attempt processing
	success, err := h.processIntasendWebhook(&event)

	now := time.Now()
	webhookLog.ProcessedAt = &now
	processingTime := int(time.Since(now).Milliseconds())
	webhookLog.ProcessingTime = &processingTime

	if success {
		webhookLog.Status = models.WebhookProcessed
		webhookLog.Success = true
		webhookLog.ErrorMessage = nil
		webhookLog.StackTrace = nil
		webhookLog.RetryCount = 0
		h.db.Save(webhookLog)
		log.Printf("✅ Webhook retry succeeded (Event: %s, Attempts: %d)", webhookLog.EventID, webhookLog.RetryCount)
		return
	}

	// Handle retry failure
	webhookLog.RetryCount++
	if webhookLog.RetryCount >= retryConfig.MaxRetries {
		webhookLog.Status = models.WebhookFailed
		webhookLog.Success = false
		if err != nil {
			errMsg := err.Error()
			webhookLog.ErrorMessage = &errMsg
		}
		h.db.Save(webhookLog)
		log.Printf("❌ Webhook retry exhausted (Event: %s, Attempts: %d)", webhookLog.EventID, webhookLog.RetryCount)
		return
	}

	// Schedule next retry
	webhookLog.Status = models.WebhookRetrying
	webhookLog.LastRetryAt = &now
	if err != nil {
		errMsg := err.Error()
		webhookLog.ErrorMessage = &errMsg
	}
	h.db.Save(webhookLog)
	log.Printf("🔄 Webhook retry scheduled (Event: %s, Attempt: %d/%d)", webhookLog.EventID, webhookLog.RetryCount, retryConfig.MaxRetries)
}

// calculateBackoff calculates exponential backoff duration
func (h *PaymentHandler) calculateBackoff(retryCount int, config WebhookRetryConfig) time.Duration {
	if retryCount == 0 {
		return config.InitialBackoff
	}

	// Exponential backoff: initialBackoff * (backoffFactor ^ retryCount)
	backoff := config.InitialBackoff
	for i := 0; i < retryCount; i++ {
		backoff = time.Duration(float64(backoff) * config.BackoffFactor)
		if backoff > config.MaxBackoff {
			backoff = config.MaxBackoff
			break
		}
	}

	return backoff
}

// GetRetryableWebhooks returns webhooks eligible for retry
func (h *PaymentHandler) GetRetryableWebhooks() ([]models.WebhookLog, error) {
	var webhooks []models.WebhookLog
	err := h.db.Where("status = ? AND retry_count < ?", models.WebhookFailed, 5).
		Order("last_retry_at ASC").
		Find(&webhooks).Error

	return webhooks, err
}

// GetWebhookRetryStats returns statistics about webhook retries
func (h *PaymentHandler) GetWebhookRetryStats() map[string]interface{} {
	var totalFailed, totalRetrying, totalSuccessful, totalDuplicate int64

	h.db.Model(&models.WebhookLog{}).Where("status = ?", models.WebhookFailed).Count(&totalFailed)
	h.db.Model(&models.WebhookLog{}).Where("status = ?", models.WebhookRetrying).Count(&totalRetrying)
	h.db.Model(&models.WebhookLog{}).Where("status = ? AND success = ?", models.WebhookProcessed, true).Count(&totalSuccessful)
	h.db.Model(&models.WebhookLog{}).Where("status = ?", models.WebhookDuplicate).Count(&totalDuplicate)

	// Average retry count for failed webhooks
	var avgRetries float64
	h.db.Model(&models.WebhookLog{}).
		Where("status = ?", models.WebhookFailed).
		Select("AVG(retry_count)").
		Row().
		Scan(&avgRetries)

	// Success rate
	var totalProcessed int64
	h.db.Model(&models.WebhookLog{}).Count(&totalProcessed)

	successRate := 0.0
	if totalProcessed > 0 {
		successRate = float64(totalSuccessful) / float64(totalProcessed) * 100
	}

	return map[string]interface{}{
		"total_failed":     totalFailed,
		"total_retrying":   totalRetrying,
		"total_successful": totalSuccessful,
		"total_duplicate":  totalDuplicate,
		"total_processed":  totalProcessed,
		"success_rate":     fmt.Sprintf("%.2f%%", successRate),
		"avg_retry_count":  fmt.Sprintf("%.2f", avgRetries),
	}
}
