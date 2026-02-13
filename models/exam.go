package models

import "time"

type QuestionType string

const (
	MCQ QuestionType = "mcq"
	TF  QuestionType = "tf"
)

type Question struct {
	QID           string       `json:"qid" firestore:"qid"`
	Type          QuestionType `json:"type" firestore:"type"`
	QuestionText  string       `json:"question_text" firestore:"question_text"`
	ImageURL      string       `json:"image_url,omitempty" firestore:"image_url,omitempty"`
	Options       []string     `json:"options,omitempty" firestore:"options,omitempty"`
	CorrectAnswer string       `json:"correct_answer,omitempty" firestore:"correct_answer,omitempty"`
	Points        float64      `json:"points" firestore:"points"`
}

type Exam struct {
	ExamID           string    `json:"exam_id" firestore:"exam_id"`
	Title            string    `json:"title" firestore:"title"`
	Class            string    `json:"class" firestore:"class"`
	TimeLimitMinutes int       `json:"time_limit_minutes" firestore:"time_limit_minutes"`
	StartTime        time.Time `json:"start_time" firestore:"start_time"`
	EndTime          time.Time `json:"end_time" firestore:"end_time"`
	CreatedBy        string    `json:"created_by" firestore:"created_by"`
	CreatedAt        time.Time `json:"created_at" firestore:"created_at"`
	Released         bool      `json:"released" firestore:"released"`
}
type Answer struct {
	QID         string  `json:"qid" firestore:"qid"`
	Response    string  `json:"response" firestore:"response"`
	ImageURL    string  `json:"image_url,omitempty" firestore:"image_url,omitempty"`
	AutoScore   float64 `json:"auto_score,omitempty" firestore:"auto_score,omitempty"`
	ManualScore float64 `json:"manual_score,omitempty" firestore:"manual_score,omitempty"`
}

type Submission struct {
	StudentID   string            `json:"student_id" firestore:"student_id"`
	StartedAt   time.Time         `json:"started_at" firestore:"started_at"`
	SubmittedAt time.Time         `json:"submitted_at,omitempty" firestore:"submitted_at,omitempty"`
	Answers     map[string]Answer `json:"answers" firestore:"answers"`
	AutoScore   float64           `json:"auto_score" firestore:"auto_score"`
	FinalScore  float64           `json:"final_score" firestore:"final_score"`
	Released    bool              `json:"released" firestore:"released"`
}

type ResultSummary struct {
	ExamID      string    `json:"exam_id"`
	Title       string    `json:"title"`
	Date        time.Time `json:"date"`
	FinalScore  float64   `json:"final_score"`
	TotalPoints float64   `json:"total_points"`
	Percentage  float64   `json:"percentage"`
}

type ResultStats struct {
	TotalQuestions int     `json:"total_questions" firestore:"total_questions"`
	Correct        int     `json:"correct" firestore:"correct"`
	Wrong          int     `json:"wrong" firestore:"wrong"`
	TotalPoints    float64 `json:"total_points" firestore:"total_points"`
	FinalScore     float64 `json:"final_score" firestore:"final_score"`
	Percentage     float64 `json:"percentage" firestore:"percentage"`
}

type ResultDetail struct {
	Exam       Exam        `json:"exam" firestore:"exam"`
	Submission Submission  `json:"submission" firestore:"submission"`
	Stats      ResultStats `json:"stats" firestore:"stats"`
}
