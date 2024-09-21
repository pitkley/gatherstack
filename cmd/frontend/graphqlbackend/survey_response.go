package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type surveyResponseResolver struct {
	db             database.DB
	surveyResponse *types.SurveyResponse
}

func (s *surveyResponseResolver) ID() graphql.ID {
	return marshalSurveyResponseID(s.surveyResponse.ID)
}
func marshalSurveyResponseID(id int32) graphql.ID { return relay.MarshalID("SurveyResponse", id) }

func (s *surveyResponseResolver) User(ctx context.Context) (*UserResolver, error) {
	if s.surveyResponse.UserID != nil {
		user, err := UserByIDInt32(ctx, s.db, *s.surveyResponse.UserID)
		if err != nil && errcode.IsNotFound(err) {
			// This can happen if the user has been deleted, see issue #4888 and #6454
			return nil, nil
		}
		return user, err
	}
	return nil, nil
}

func (s *surveyResponseResolver) Email() *string {
	return s.surveyResponse.Email
}

func (s *surveyResponseResolver) Score() int32 {
	return s.surveyResponse.Score
}

func (s *surveyResponseResolver) Reason() *string {
	return s.surveyResponse.Reason
}

func (s *surveyResponseResolver) Better() *string {
	return s.surveyResponse.Better
}

func (s *surveyResponseResolver) OtherUseCase() *string {
	return s.surveyResponse.OtherUseCase
}

func (s *surveyResponseResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: s.surveyResponse.CreatedAt}
}

// SurveySubmissionInput contains a satisfaction (NPS) survey response.
type SurveySubmissionInput struct {
	// Emails is an optional, user-provided email address, if there is no
	// currently authenticated user. If there is, this value will not be used.
	Email *string
	// Score is the user's likelihood of recommending Sourcegraph to a friend, from 0-10.
	Score int32
	// OtherUseCase is the answer to "What do you use Sourcegraph for?".
	OtherUseCase *string
	// Better is the answer to "What can Sourcegraph do to provide a better product"
	Better *string
}

// SubmitSurvey records a new satisfaction (NPS) survey response by the current user.
// TODO(LIBRE): remove this type altogether
func (r *schemaResolver) SubmitSurvey(_ context.Context, _ *struct {
	Input *SurveySubmissionInput
}) (*EmptyResponse, error) {
	return &EmptyResponse{}, nil
}

// HappinessFeedbackSubmissionInput contains a happiness feedback response.
type HappinessFeedbackSubmissionInput struct {
	// Score is the user's happiness rating, from 1-4.
	Score int32
	// Feedback is the feedback text from the user.
	Feedback *string
	// The path that the happiness feedback was submitted from
	CurrentPath *string
}

// SubmitHappinessFeedback records a new happiness feedback response by the current user.
// TODO(LIBRE): remove this type altogether
func (r *schemaResolver) SubmitHappinessFeedback(_ context.Context, _ *struct {
	Input *HappinessFeedbackSubmissionInput
}) (*EmptyResponse, error) {
	return &EmptyResponse{}, nil
}
