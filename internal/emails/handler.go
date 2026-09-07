package emails

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"email-job/internal/domains"
	"email-job/pkg"
)

type EmailHandler struct {
	emailSvc EmailService
}

func NewEmailHandler(emailSvc EmailService) *EmailHandler {
	return &EmailHandler{
		emailSvc: emailSvc,
	}
}

func httpError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, pkg.ErrNotFound):
		pkg.ErrorResponse(c, http.StatusNotFound, err)
	case errors.Is(err, pkg.ErrInvalidInput):
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
	default:
		pkg.ErrorResponse(c, http.StatusInternalServerError, err)
	}
}

// Emails Sending

func (h *EmailHandler) SendContactMessage(c *gin.Context) {
	ip := c.ClientIP()

	var req SendMessageEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	if err := h.emailSvc.SendContactMessage(c.Request.Context(), ip, req); err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusAccepted, "message has been received", nil)
}

func (h *EmailHandler) SendLiveDemoRequest(c *gin.Context) {
	ip := c.ClientIP()

	var req SendLiveDemoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	if err := h.emailSvc.SendLiveDemoRequest(c.Request.Context(), ip, req); err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusAccepted, "demo request received", nil)
}

func (h *EmailHandler) SendLiveDemoReady(c *gin.Context) {
	id := c.Param("id")

	var req SendLiveDemoManualRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	if err := h.emailSvc.SendLiveDemoReady(c.Request.Context(), id, req.TestingLink); err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "live demo email sent successfully", nil)
}

// Emails GET

func (h *EmailHandler) GetEmailByID(c *gin.Context) {
	id := c.Param("id")

	res, err := h.emailSvc.GetEmailByID(c.Request.Context(), id)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "get email successfully", map[string]interface{}{
		"email": res,
	})
}

func (h *EmailHandler) GetEmailByIP(c *gin.Context) {
	ip := c.Param("ip")

	res, err := h.emailSvc.GetEmailByIP(c.Request.Context(), ip)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "get email successfully", map[string]interface{}{
		"email": res,
	})
}

func (h *EmailHandler) GetEmails(c *gin.Context) {
    strPage  := c.DefaultQuery("page", "1")
    strLimit := c.DefaultQuery("limit", "10")
    status   := c.Query("status")

    page, err := strconv.Atoi(strPage)
    if err != nil {
        page = 1
    }

    limit, err := strconv.Atoi(strLimit)
    if err != nil {
        limit = 10
    }

    var (
        res any
    )

    if status != "" {
        statusQuery := domains.Status(status)

        if !statusQuery.IsValid() {
            httpError(c, pkg.ErrInvalidInput)
            return
        }

        res, err = h.emailSvc.GetEmailsByStatus(c.Request.Context(), status, page, limit)
    } else {
        res, err = h.emailSvc.GetEmails(c.Request.Context(), page, limit)
    }

    if err != nil {
        httpError(c, err)
        return
    }

    pkg.SuccessResponse(c, http.StatusOK, "get emails successfully", res)
}

func (h *EmailHandler) GetTodayEmails(c *gin.Context) {
	strPage   := c.DefaultQuery("page", "1")
	strLlimit := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(strPage)
	if err != nil {
		page = 1
	}

	limit, err := strconv.Atoi(strLlimit)
	if err != nil {
		limit = 10
	}

	res, err := h.emailSvc.GetTodayEmails(c.Request.Context(), page, limit)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "get today emails successfully", res)
}

func (h *EmailHandler) GetEmailsByLiveDemoRequest(c *gin.Context) {
	strPage   := c.DefaultQuery("page", "1")
	strLlimit := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(strPage)
	if err != nil {
		page = 1
	}

	limit, err := strconv.Atoi(strLlimit)
	if err != nil {
		limit = 10
	}

	typ := "demo_request"

	res, err := h.emailSvc.GetEmailsByLiveDemoRequest(c.Request.Context(), typ, page, limit)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "get live demo requests successfully", res)
}