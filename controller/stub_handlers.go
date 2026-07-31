package controller

import "github.com/gin-gonic/gin"

// GetSubscriptionPlans is a stub handler
func GetSubscriptionPlans(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"data":    []interface{}{},
	})
}

// GetUserSubscription is a stub handler
func GetUserSubscription(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"subscription": nil,
		},
	})
}

// GetTwoFAStatus is a stub handler
func GetTwoFAStatus(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"enabled": false,
		},
	})
}
