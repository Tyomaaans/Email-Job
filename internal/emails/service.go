package emails

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"

	"email-job/internal/domains"
	"email-job/internal/infrastructures/client/mailtrap"
	"email-job/pkg"
	jsonValidator "email-job/pkg"
)

const (
	dailyEmailLimit = 130
	redisCounterKey = "email:daily:count"
	queueName       = "email.queue"
	retryQueueName  = "email.retry.queue"
	retryExchange   = "email.retry.exchange"
	mainExchange    = "email.main.exchange"
)

// Interface

type EmailService interface {
	// Emails Sending
	SendContactMessage(ctx context.Context, ip string, req SendMessageEmailRequest) error
	SendLiveDemoRequest(ctx context.Context, ip string, req SendLiveDemoRequest) error
	SendLiveDemoReady(ctx context.Context, rawID, link string) error
	DeliverQueued(ctx context.Context, queueEmail domains.EmailQueueEntity) error

	// Emails GET
	GetEmailByID(ctx context.Context, rawID string) (*EmailResponse, error)
	GetEmailByIP(ctx context.Context, ip string) (*EmailResponse, error)
	GetEmails(ctx context.Context, page, limit int) (*PaginatedEmailResponse, error)
	GetEmailsByStatus(ctx context.Context, status string, page, limit int) (*PaginatedEmailResponse, error)
	GetEmailsByLiveDemoRequest(ctx context.Context, typ string, page, limit int) (*PaginatedEmailResponse, error)
	GetTodayEmails(ctx context.Context, page, limit int) (*PaginatedEmailResponse, error)
}

// Implementation

type emailService struct {
	sender    *mailtrap.Client
	redis     *redis.Client
	amqpCh    *amqp.Channel
	ownerTo   string
	emailRepo domains.EmailRepository
	validate  *validator.Validate
}

func NewEmailService(
	sender      *mailtrap.Client,
	redisClient *redis.Client,
	amqpCh      *amqp.Channel,
	ownerTo     string,
	emailRepo   domains.EmailRepository,
	validate    *validator.Validate,
) EmailService {
	return &emailService{
		sender:    sender,
		redis:     redisClient,
		amqpCh:    amqpCh,
		ownerTo:   ownerTo,
		emailRepo: emailRepo,
		validate:  validate,
	}
}

// Emails Sending

func (s *emailService) SendContactMessage(ctx context.Context, ip string, req SendMessageEmailRequest) error {
	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return err
	}
	
	contact, err := ToContactQueueEntity(uuid.NewString(), ip, s.ownerTo, req)
	if err != nil {
		return err
	}

	return s.enqueue(ctx, contact)
}

func (s *emailService) SendLiveDemoRequest(ctx context.Context, ip string, req SendLiveDemoRequest) error {
	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return err
	}

	liveDemo, err := ToLiveDemoRequestQueueEntity(uuid.NewString(), ip, s.ownerTo, req)
	if err != nil {
		return err
	}

	return s.enqueue(ctx, liveDemo)
}

func (s *emailService) SendLiveDemoReady(ctx context.Context, rawID, link string) error {
	id, err := parseOrDecodeUUID(rawID)
    if err != nil {
        return err
    }

	email, err := s.emailRepo.GetEmailByID(ctx, id)
	if err != nil {
		return err
	}

	data := ToLiveDemoReadyEmailDataEntity(link, email)

	return s.sendTemplatedEmail(ctx, email.ToEmail, "Your Live Demo Is Here — Let's Get Started", liveDemoReadyTpl, data)
}

func (s *emailService) DeliverQueued(ctx context.Context, queueEmail domains.EmailQueueEntity) error {
	allowed, err := s.checkAndIncrement(ctx)
	if err != nil {
		return err
	}

	if !allowed {
		log.Printf("mailtrap: daily limit reached! retrying for %s, type %s", queueEmail.To, queueEmail.Type)
		return s.enqueueRetry(ctx, queueEmail)
	}

	return s.deliver(ctx, queueEmail)
}

// Emails GET

func (s *emailService) GetEmailByID(ctx context.Context, rawID string) (*EmailResponse, error) {
	id, err := parseOrDecodeUUID(rawID)
    if err != nil {
        return nil, err
    }

	email, err := s.emailRepo.GetEmailByID(ctx, id)
	if err != nil {
		return nil, err
	}

	res, err := ToEmailResponse(email)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *emailService) GetEmailByIP(ctx context.Context, ip string) (*EmailResponse, error) {
	email, err := s.emailRepo.GetEmailByIP(ctx, ip)
	if err != nil {
		return nil, err
	}

	res, err := ToEmailResponse(email)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *emailService) GetEmails(ctx context.Context, page, limit int) (*PaginatedEmailResponse, error) {
	emails, pages, err := s.emailRepo.GetEmails(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	res, err := ToPaginatedEmailResponse(emails, page, limit, pages)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *emailService) GetEmailsByStatus(ctx context.Context, status string, page, limit int) (*PaginatedEmailResponse, error) {
	emails, pages, err := s.emailRepo.GetEmailsByStatus(ctx, domains.Status(status), page, limit)
	if err != nil {
		return nil, err
	}

	res, err := ToPaginatedEmailResponse(emails, page, limit, pages)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *emailService) GetEmailsByLiveDemoRequest(ctx context.Context, typ string, page, limit int) (*PaginatedEmailResponse, error) {
	emails, pages, err := s.emailRepo.GetEmailsByLiveDemoRequest(ctx, typ, page, limit)
	if err != nil {
		return nil, err
	}

	res, err := ToPaginatedEmailResponse(emails, page, limit, pages)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *emailService) GetTodayEmails(ctx context.Context, page, limit int) (*PaginatedEmailResponse, error) {
	emails, pages, err := s.emailRepo.GetTodayEmails(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	res, err := ToPaginatedEmailResponse(emails, page, limit, pages)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Internal Helpers

func (s *emailService) enqueue(ctx context.Context, queueEmail domains.EmailQueueEntity) error {
	body, err := json.Marshal(queueEmail)
	if err != nil {
		return err
	}

	job := &domains.EmailEntity{
		ID:        queueEmail.ID,
		IpAddress: queueEmail.IpAddress,
		Name:      queueEmail.Name,
		Project:   queueEmail.Project,
		ToEmail:   queueEmail.To,
		Type:      queueEmail.Type,
		Status:    domains.Pending,
		Attempt:   0,
	}
	if err := s.emailRepo.CreateEmail(ctx, job); err != nil {
		return err
	}

	err = s.amqpCh.PublishWithContext(ctx, mainExchange, queueName, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Priority:     0,
		Body:         body,
	})
	if err != nil {
		return pkg.HandleRabbitError(err)
	}

	return nil
}

func (s *emailService) enqueueRetry(ctx context.Context, queueEmail domains.EmailQueueEntity) error {
	body, err := json.Marshal(queueEmail)
	if err != nil {
		return err
	}

	if err := s.emailRepo.UpdateEmail(ctx, &domains.EmailEntity{
		ID:      queueEmail.ID,
		Status:  domains.Retrying,
		Attempt: queueEmail.Attempt + 1,
	}); err != nil {
		return err
	}

	delay := s.delayUntilNextReset()

	err = s.amqpCh.PublishWithContext(ctx, retryExchange, retryQueueName, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Priority:     10,
		Expiration:   fmt.Sprintf("%d", delay.Milliseconds()),
		Body:         body,
	})
	if err != nil {
		return pkg.HandleRabbitError(err)
	}

	return nil
}

func (s *emailService) delayUntilNextReset() time.Duration {
	now := time.Now().UTC()
	nextReset := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	return time.Until(nextReset)
}

func (s *emailService) deliver(ctx context.Context, queueEmail domains.EmailQueueEntity) error {
	tplSource, data, err := ToRenderData(queueEmail)
	if err != nil {
		return err
	}

	body, err := renderEmailTemplate(tplSource, data)
	if err != nil {
		return err
	}

	if _, err := s.sender.SendEmail(ctx, queueEmail.To, queueEmail.Subject, body); err != nil {
		errMsg := err.Error()
		if err := s.emailRepo.UpdateEmail(ctx, &domains.EmailEntity{
			ID:        queueEmail.ID,
			Status:    domains.Failed,
			LastError: &errMsg,
		}); err != nil {
			log.Printf("sqlite: failed to persist job status %v", err)
		}
		return err
	}

	now := time.Now().UTC()
	if err := s.emailRepo.UpdateEmail(ctx, &domains.EmailEntity{
		ID:     queueEmail.ID,
		Status: domains.Sent,
		SentAt: &now,
	}); err != nil {
		log.Printf("sqlite: failed to persist sent job %v", err)
	}

	return nil
}

func (s *emailService) checkAndIncrement(ctx context.Context) (bool, error) {
	pipe := s.redis.TxPipeline()

	incr := pipe.Incr(ctx, redisCounterKey)

	now := time.Now().UTC()
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	ttl := time.Until(midnight)
	pipe.ExpireNX(ctx, redisCounterKey, ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		return false, pkg.HandleRedisError(err)
	}

	count := incr.Val()
	if count > int64(dailyEmailLimit) {
		if err := s.redis.Decr(ctx, redisCounterKey).Err(); err != nil {
			return false, pkg.HandleRedisError(err)
		}
		return false, nil
	}

	return true, nil
}

func (s *emailService) sendTemplatedEmail(ctx context.Context, to, subject, tplSource string, data any) error {
	body, err := renderEmailTemplate(tplSource, data)
	if err != nil {
		return err
	}

	if _, err := s.sender.SendEmail(ctx, to, subject, body); err != nil {
		return err
	}

	return nil
}

func renderEmailTemplate(tplSource string, data any) (string, error) {
	t, err := template.New("email").Parse(tplSource)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func parseOrDecodeUUID(input string) (string, error) {
    if input == "" {
        return "", pkg.ErrInvalidInput
    }

    if _, err := uuid.Parse(input); err == nil {
        return input, nil
    }

    decoded, err := base64.RawURLEncoding.DecodeString(input)
    if err != nil {
        return "", pkg.ErrInvalidInput
    }

    parsed, err := uuid.FromBytes(decoded)
    if err != nil {
        return "", pkg.ErrInvalidInput
    }

    return parsed.String(), nil
}