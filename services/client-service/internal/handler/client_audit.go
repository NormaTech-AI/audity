package handler

import (
    "net/http"
    "fmt"

    "github.com/NormaTech-AI/audity/packages/go/auth"
    "github.com/labstack/echo/v4"
)

// ListClientAudits proxies to tenant-service to list audits for the authenticated user
func (h *Handler) ListClientAudits(c echo.Context) error {
    // Ensure user is authenticated
    user, err := auth.GetUserFromContext(c)
    if err != nil {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
    }

    // Forward request to tenant-service
    url := fmt.Sprintf("%s/api/client-audit", h.getTenantBaseURL())
    req, err := http.NewRequest(http.MethodGet, url, nil)
    if err != nil {
        h.logger.Errorw("Failed to create request to tenant-service", "error", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create upstream request"})
    }

    // Copy auth headers and cookies
    req.Header.Set("Authorization", c.Request().Header.Get("Authorization"))
    for _, cookie := range c.Request().Cookies() {
        req.AddCookie(cookie)
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        h.logger.Errorw("Failed to call tenant-service", "error", err, "user_id", user.UserID)
        return c.JSON(http.StatusBadGateway, map[string]string{"error": "Upstream service unavailable"})
    }
    defer resp.Body.Close()

    return c.Stream(resp.StatusCode, resp.Header.Get("Content-Type"), resp.Body)
}

// GetClientAuditDetail proxies to tenant-service to fetch audit detail with questions
func (h *Handler) GetClientAuditDetail(c echo.Context) error {
    user, err := auth.GetUserFromContext(c)
    if err != nil {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
    }

    auditID := c.Param("auditId")
    if auditID == "" {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing auditId"})
    }

    url := fmt.Sprintf("%s/api/client-audit/%s", h.getTenantBaseURL(), auditID)
    req, err := http.NewRequest(http.MethodGet, url, nil)
    if err != nil {
        h.logger.Errorw("Failed to create request to tenant-service", "error", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create upstream request"})
    }

    req.Header.Set("Authorization", c.Request().Header.Get("Authorization"))
    for _, cookie := range c.Request().Cookies() {
        req.AddCookie(cookie)
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        h.logger.Errorw("Failed to call tenant-service", "error", err, "user_id", user.UserID)
        return c.JSON(http.StatusBadGateway, map[string]string{"error": "Upstream service unavailable"})
    }
    defer resp.Body.Close()

    return c.Stream(resp.StatusCode, resp.Header.Get("Content-Type"), resp.Body)
}

// SaveClientSubmission proxies to tenant-service to save draft submission
func (h *Handler) SaveClientSubmission(c echo.Context) error {
    user, err := auth.GetUserFromContext(c)
    if err != nil {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
    }

    url := fmt.Sprintf("%s/api/client-audit/submissions", h.getTenantBaseURL())
    req, err := http.NewRequest(http.MethodPost, url, c.Request().Body)
    if err != nil {
        h.logger.Errorw("Failed to create request to tenant-service", "error", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create upstream request"})
    }

    // Copy content-type and auth
    req.Header.Set("Content-Type", c.Request().Header.Get("Content-Type"))
    req.Header.Set("Authorization", c.Request().Header.Get("Authorization"))
    for _, cookie := range c.Request().Cookies() {
        req.AddCookie(cookie)
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        h.logger.Errorw("Failed to call tenant-service", "error", err, "user_id", user.UserID)
        return c.JSON(http.StatusBadGateway, map[string]string{"error": "Upstream service unavailable"})
    }
    defer resp.Body.Close()

    return c.Stream(resp.StatusCode, resp.Header.Get("Content-Type"), resp.Body)
}

// SubmitClientAnswer proxies to tenant-service to submit answer for review
func (h *Handler) SubmitClientAnswer(c echo.Context) error {
    user, err := auth.GetUserFromContext(c)
    if err != nil {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
    }

    submissionID := c.Param("submissionId")
    if submissionID == "" {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing submissionId"})
    }

    url := fmt.Sprintf("%s/api/client-audit/submissions/%s/submit", h.getTenantBaseURL(), submissionID)
    req, err := http.NewRequest(http.MethodPost, url, nil)
    if err != nil {
        h.logger.Errorw("Failed to create request to tenant-service", "error", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create upstream request"})
    }

    req.Header.Set("Authorization", c.Request().Header.Get("Authorization"))
    for _, cookie := range c.Request().Cookies() {
        req.AddCookie(cookie)
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        h.logger.Errorw("Failed to call tenant-service", "error", err, "user_id", user.UserID)
        return c.JSON(http.StatusBadGateway, map[string]string{"error": "Upstream service unavailable"})
    }
    defer resp.Body.Close()

    return c.Stream(resp.StatusCode, resp.Header.Get("Content-Type"), resp.Body)
}