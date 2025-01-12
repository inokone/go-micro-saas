package common

import (
	"context"
	"fmt"

	recaptcha "cloud.google.com/go/recaptchaenterprise/v2/apiv1"
	recaptchapb "cloud.google.com/go/recaptchaenterprise/v2/apiv1/recaptchaenterprisepb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

const captchaThreshold = 0.5

// RecaptchaValidator is a struct for validating captcha using Google's ReCapthca validator
type RecaptchaValidator struct {
	projectID    string
	recaptchaKey string
	client       *recaptcha.Client
	ctx          context.Context
}

// NewRecaptchaValidator is a function creating a new `RecaptchaValidator` based on the API secret
func NewRecaptchaValidator(projectID string, recaptchaKey string, keyFile string) (*RecaptchaValidator, error) {
	ctx := context.Background()
	client, err := recaptcha.NewClient(ctx, option.WithCredentialsFile(keyFile))

	if err != nil {
		return nil, fmt.Errorf("error creating reCAPTCHA client: %v", err)
	}

	return &RecaptchaValidator{
		projectID:    projectID,
		recaptchaKey: recaptchaKey,
		client:       client,
		ctx:          ctx,
	}, nil
}

// Verify is a method of `RecaptchaValidator` verifying the captch token from the frontend
func (v RecaptchaValidator) Verify(token string) error {
	// Create an assessment request
	event := &recaptchapb.Event{
		Token:   token,
		SiteKey: v.recaptchaKey,
	}
	assessment := &recaptchapb.Assessment{
		Event: event,
	}
	request := &recaptchapb.CreateAssessmentRequest{
		Assessment: assessment,
		Parent:     fmt.Sprintf("projects/%s", v.projectID),
	}

	// Send the request and get the response
	response, err := v.client.CreateAssessment(v.ctx, request)
	if err != nil {
		return fmt.Errorf("error creating assessment: %v", err)
	}

	// Check if the token is valid
	if !response.TokenProperties.Valid {
		return fmt.Errorf("the reCAPTCHA token is invalid: %v", response.TokenProperties.InvalidReason)
	}

	// Get the risk score and reasons
	if response.RiskAnalysis.Score < captchaThreshold {
		err := fmt.Errorf("the reCAPTCHA risk score is too low: %v", response.RiskAnalysis.Score)
		log.WithError(err).WithField("reasons", response.RiskAnalysis.Reasons).Error("ReCAPTCHA failed")
		return err
	}
	return nil
}

func (v RecaptchaValidator) Close() {
	v.client.Close()
}
