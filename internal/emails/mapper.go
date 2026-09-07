package emails

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"email-job/internal/domains"
	"email-job/internal/infrastructures/sqlite"

	"github.com/google/uuid"
)

// Email Storage <-> Entity

func ToEmailStorage(e *domains.EmailEntity) *sqlite.Email {
	if e == nil {
		return nil
	}

	return &sqlite.Email{
		ID:        e.ID,
		IpAddress: e.IpAddress,
		Name:      e.Name,
		ToEmail:   e.ToEmail,
		Type:      e.Type,
		Status:    e.Status,
		Project:   e.Project,
		Attempt:   e.Attempt,
		LastError: e.LastError,
		SentAt:    e.SentAt,
	}
}

func ToEmailEntity(s *sqlite.Email) *domains.EmailEntity {
	if s == nil {
		return nil
	}

	return &domains.EmailEntity{
		ID:        s.ID,
		IpAddress: s.IpAddress,
		Name:      s.Name,
		ToEmail:   s.ToEmail,
		Type:      s.Type,
		Status:    s.Status,
		Project:   s.Project,
		Attempt:   s.Attempt,
		LastError: s.LastError,
		SentAt:    s.SentAt,
	}
}

func ToEmailListEntities(list []sqlite.Email) []domains.EmailEntity {
	result := make([]domains.EmailEntity, len(list))

	for i := range list {
		result[i] = *ToEmailEntity(&list[i])
	}

	return result
}

// To Email Queeu Entity

func ToContactQueueEntity(id, ip, ownerTo string, req SendMessageEmailRequest) (domains.EmailQueueEntity, error) {
	payload :=  domains.ContactQueueEntity{
		Name:    req.Name,
		Email:   req.Email,
		Message: req.Message,
	}

	return toEmailQueueEntity(id, ip, req.Name, "", domains.EmailTypeContact, ownerTo, toContactSubject(req.Name), payload)
}

func ToLiveDemoRequestQueueEntity(id, ip, ownerTo string, req SendLiveDemoRequest) (domains.EmailQueueEntity, error) {
	payload := domains.DemoRequestQueueEntity{
		Name:    req.Name,
		Email:   req.Email,
		Project: req.Project,
		Message: req.Message,
	}

	return toEmailQueueEntity(id, ip, req.Name, domains.Project(req.Project), domains.EmailTypeDemoRequest, ownerTo, toDemoRequestSubject(string(req.Project)), payload)
}

// To Email Data Entity

func ToContactEmailDataEntity(e domains.EmailQueueEntity) (domains.ContactEmailDataEntity, error) {
	var p domains.ContactQueueEntity
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		return domains.ContactEmailDataEntity{}, fmt.Errorf("email: unmarshal contact payload: %w", err)
	}
	return domains.ContactEmailDataEntity{
		Name:      p.Name,
		FromEmail: p.Email,
		Message:   p.Message,
	}, nil
}

func ToDemoRequestEmailDataEntity(e domains.EmailQueueEntity) (domains.DemoRequestEmailDataEntity, error) {
	var p domains.DemoRequestQueueEntity
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		return domains.DemoRequestEmailDataEntity{}, fmt.Errorf("email: unmarshal demo_request payload: %w", err)
	}
	return domains.DemoRequestEmailDataEntity{
		FromEmail: p.Email,
		Project:   p.Project,
		Message:   p.Message,
	}, nil
}

func ToLiveDemoReadyEmailDataEntity(link string, e *domains.EmailEntity) domains.EmailDataEntity {
	return domains.EmailDataEntity{
		Name:        e.Name,
		TestingLink: link,
	}
}

// To Email Response

func ToEmailResponse(e *domains.EmailEntity) (*EmailResponse, error) {
	id, err := uuidToBase64(e.ID)
	if err != nil {
		return nil, err
	}

	return &EmailResponse{
		ID:        id,
		IpAddress: e.IpAddress,
		Name:      e.Name,
		ToEmail:   e.ToEmail,
		Type:      e.Type,
		Status:    e.Status,
		Project:   e.Project,
		Attempt:   e.Attempt,
		LastError: e.LastError,
		SentAt:    e.SentAt,
	}, nil
}

func ToEmailListResponse(list []domains.EmailEntity) ([]EmailResponse, error) {
    result := make([]EmailResponse, len(list))

    for i := range list {
        res, err := ToEmailResponse(&list[i])
        if err != nil {
            return nil, err
        }
        result[i] = *res
    }

    return result, nil
}

func ToPaginatedEmailResponse(list []domains.EmailEntity, page, limit int, totalData int64) (*PaginatedEmailResponse, error) {
    listResponse, err := ToEmailListResponse(list)
    if err != nil {
        return nil, err
    }

    if limit <= 0 {
        limit = 10
    }

    totalPages := int((totalData + int64(limit) - 1) / int64(limit))

    return &PaginatedEmailResponse{
        Data: listResponse,
        Meta: PaginationMeta{
            CurrentPage: page,
            TotalPages:  totalPages,
            PageSize:    limit,
            TotalData:   totalData,
        },
    }, nil
}

// Internal Helpers

func ToRenderData(e domains.EmailQueueEntity) (tplSource string, data any, err error) {
	switch e.Type {
	case domains.EmailTypeContact:
		data, err = ToContactEmailDataEntity(e)
		return contactMessageTpl, data, err

	case domains.EmailTypeDemoRequest:
		data, err = ToDemoRequestEmailDataEntity(e)
		return liveDemoRequestTpl, data, err

	default:
		return "", nil, fmt.Errorf("email: unknown email type: %s", e.Type)
	}
}

func toContactSubject(name string) string {
	return "New message from " + name
}

func toDemoRequestSubject(project string) string {
	return "Live demo request [" + project + "]"
}

func toEmailQueueEntity(id, ip, name string, project domains.Project, typ domains.EmailType, to, subject string, payload any) (domains.EmailQueueEntity, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return domains.EmailQueueEntity{}, fmt.Errorf("email: marshal %s payload: %w", typ, err)
	}

	return domains.EmailQueueEntity{
		ID:        id,
		IpAddress: ip,
		Name:      name,
		Project:   &project,
		Type:      typ,
		To:        to,
		Subject:   subject,
		Payload:   raw,
	}, nil
}

func uuidToBase64(uuidStr string) (string, error) {
	parsed, err := uuid.Parse(uuidStr)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(parsed[:]), nil
}