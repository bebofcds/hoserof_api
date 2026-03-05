/*
================================================================================
HOSEROF_API - Exam Services
================================================================================

Description:
This package provides service-level functions for creating, managing, and
submitting exams within the HosErof system. It also handles retrieving
exam questions, managing submissions, releasing results, and calculating
student scores. All exam-related data is stored in Firebase Firestore.

Responsibilities:
1. CreateExam                    - Creates a new exam and its questions.
2. GetExamsForClass              - Retrieves active exams for a student.
3. GetAllExamsForAdmin            - Retrieves all exams for administrative purposes.
4. GetExamQuestions              - Fetches exam questions, optionally hiding correct answers for students.
5. SubmitExam                    - Records a student's submission, auto-grading MCQ and TF questions.
6. GetSubmission                  - Retrieves a specific student's submission.
7. GetAllSubmissions              - Retrieves all submissions for an exam.
8. ReleaseResults                 - Releases results for all students after exam ends.
9. DeleteExam                     - Deletes an exam along with its questions and submissions.
10. GetReleasedResult             - Retrieves detailed results for a student after release.
11. GetAllReleasedResultsForStudent - Retrieves summaries of all released exams for a student.

Usage Notes:
- All functions require a Gin context (`*gin.Context`) to extract services.
- Exams and submissions are stored in Firestore:
    - Collection: "exams"
    - Subcollections: "questions" and "submissions"
- Only MCQ and True/False questions are automatically graded.
- ReleaseResults must be called after the exam's EndTime to release grades.
- Timestamps are based on `time.Now()` in the local timezone.

Error Handling:
- Returns descriptive errors for invalid operations such as:
    - Exam not started / ended
    - Already submitted
    - Results not released yet
- Firestore errors are propagated for logging or client response.

Security Notes:
- Correct answers are hidden from students in `GetExamQuestions`.
- Results can only be retrieved after they are released.

================================================================================
*/

package services

import (
	"HOSEROF_API/config"
	"HOSEROF_API/models"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateExam creates a new exam document with associated questions in Firestore.
func CreateExam(exam models.Exam, questions []models.Question, c *gin.Context) (string, error) {
	services := config.GetServices(c)
	ctx := c.Request.Context()
	exams := services.Firebase.DB.Collection("exams")
	doc := exams.NewDoc()
	exam.ExamID = doc.ID
	exam.CreatedAt = time.Now()
	exam.Released = false

	_, err := doc.Set(ctx, exam)
	if err != nil {
		return "", err
	}
	// Create questions subcollection
	for _, q := range questions {
		if q.QID == "" {
			q.QID = doc.Collection("questions").NewDoc().ID
		}
		_, err := doc.Collection("questions").Doc(q.QID).Set(ctx, q)
		if err != nil {
			return "", err
		}
	}

	return doc.ID, nil
}

// GetExamsForClass returns active exams for a student in a class, excluding already submitted exams.
func GetExamsForClass(class string, studentID string, c *gin.Context) ([]models.Exam, error) {
	ctx := c.Request.Context()
	services := config.GetServices(c)

	q := services.Firebase.DB.Collection("exams").
		Where("class", "==", class)
	iter := q.Documents(ctx)
	var out []models.Exam
	now := time.Now()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var e models.Exam
		if err := doc.DataTo(&e); err != nil {
			return nil, err
		}
		// Filter exams by current time
		if now.Before(e.StartTime) {
			continue
		}
		if now.After(e.EndTime) {
			continue
		}
		// Skip if student already submitted
		subSnap, err := doc.Ref.Collection("submissions").Doc(studentID).Get(ctx)
		if err == nil && subSnap.Exists() {
			continue
		}

		out = append(out, e)
	}

	return out, nil
}

// GetAllExamsForAdmin returns all exams for admin purposes.
func GetAllExamsForAdmin(c *gin.Context) ([]models.Exam, error) {
	ctx := c.Request.Context()
	services := config.GetServices(c)

	iter := services.Firebase.DB.Collection("exams").Documents(ctx)
	var out []models.Exam
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var e models.Exam
		if err := doc.DataTo(&e); err != nil {
			return nil, err
		}

		out = append(out, e)
	}

	return out, nil
}

// GetExamQuestions retrieves exam questions; hides correct answers if forStudent is true.
func GetExamQuestions(examID string, forStudent bool, c *gin.Context) ([]models.Question, error) {
	ctx := c.Request.Context()
	services := config.GetServices(c)

	qIter := services.Firebase.DB.Collection("exams").
		Doc(examID).
		Collection("questions").
		OrderBy("qid", firestore.Asc).
		Documents(ctx)

	var qs []models.Question

	for {
		doc, err := qIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var q models.Question
		if err := doc.DataTo(&q); err != nil {
			return nil, err
		}
		// Hide correct answers for students
		if forStudent {
			q.CorrectAnswer = ""
		}

		qs = append(qs, q)
	}

	return qs, nil
}

// SubmitExam records student's answers, auto-grading MCQ and True/False questions.
func SubmitExam(examID string, studentID string, answers map[string]models.Answer, c *gin.Context) error {
	ctx := c.Request.Context()
	services := config.GetServices(c)

	examDoc := services.Firebase.DB.Collection("exams").Doc(examID)
	snap, err := examDoc.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return errors.New("exam not found")
		}
		return err
	}
	var exam models.Exam
	if err := snap.DataTo(&exam); err != nil {
		return err
	}

	now := time.Now()
	if now.Before(exam.StartTime) {
		return errors.New("exam not started")
	}
	if now.After(exam.EndTime) {
		return errors.New("exam ended")
	}

	subDoc := examDoc.Collection("submissions").Doc(studentID)
	existsSnap, err := subDoc.Get(ctx)
	if err == nil && existsSnap.Exists() {
		return errors.New("already submitted")
	}
	// Build map of questions for auto-scoring
	qIter := examDoc.Collection("questions").Documents(ctx)
	correctMap := make(map[string]models.Question)
	for {
		doc, err := qIter.Next()
		if err != nil {
			break
		}
		var q models.Question
		if err := doc.DataTo(&q); err == nil {
			correctMap[q.QID] = q
		}
	}

	var autoScore float64 = 0
	for qid, ans := range answers {
		q, ok := correctMap[qid]
		if !ok {
			continue
		}
		if q.Type == models.MCQ || q.Type == models.TF {
			studentAns := strings.TrimSpace(strings.ToLower(fmt.Sprint(ans.Response)))
			correctAns := strings.TrimSpace(strings.ToLower(fmt.Sprint(q.CorrectAnswer)))

			if studentAns == correctAns {
				autoScore += q.Points
				a := ans
				a.AutoScore = q.Points
				answers[qid] = a
			} else {
				a := ans
				a.AutoScore = 0
				answers[qid] = a
			}
		} else {
			a := ans
			a.AutoScore = 0
			answers[qid] = a
		}
	}

	submission := models.Submission{
		StudentID:   studentID,
		StartedAt:   now,
		SubmittedAt: now,
		Answers:     answers,
		FinalScore:  autoScore,
		Released:    false,
	}

	_, err = subDoc.Set(ctx, submission)
	if err != nil {
		return err
	}

	return nil
}

// GetSubmission retrieves a single student's submission for an exam.
func GetSubmission(examID, studentID string, c *gin.Context) (models.Submission, error) {
	ctx := c.Request.Context()
	services := config.GetServices(c)

	doc := services.Firebase.DB.Collection("exams").Doc(examID).Collection("submissions").Doc(studentID)
	snap, err := doc.Get(ctx)
	if err != nil {
		return models.Submission{}, err
	}
	var s models.Submission
	if err := snap.DataTo(&s); err != nil {
		return models.Submission{}, err
	}
	return s, nil
}

// GetAllSubmissions retrieves all submissions for a given exam.
func GetAllSubmissions(examID string, c *gin.Context) ([]models.Submission, error) {
	ctx := c.Request.Context()
	services := config.GetServices(c)

	iter := services.Firebase.DB.Collection("exams").
		Doc(examID).
		Collection("submissions").
		Documents(ctx)

	var out []models.Submission

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var s models.Submission
		if err := doc.DataTo(&s); err != nil {
			return nil, err
		}

		out = append(out, s)
	}

	return out, nil
}

// ReleaseResults marks an exam and all submissions as released after exam ends.
func ReleaseResults(examID string, c *gin.Context) error {
	ctx := c.Request.Context()
	services := config.GetServices(c)

	examRef := services.Firebase.DB.Collection("exams").Doc(examID)

	snap, err := examRef.Get(ctx)
	if err != nil {
		return err
	}

	var exam models.Exam
	if err := snap.DataTo(&exam); err != nil {
		return err
	}

	if time.Now().Before(exam.EndTime) {
		return errors.New("cannot release before exam ends")
	}
	// Mark exam released
	_, err = examRef.Update(ctx, []firestore.Update{
		{Path: "released", Value: true},
	})
	if err != nil {
		return err
	}
	// Mark all submissions released
	iter := examRef.Collection("submissions").Documents(ctx)
	for {
		d, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			break
		}
		d.Ref.Update(ctx, []firestore.Update{
			{Path: "released", Value: true},
		})
	}

	return nil
}

// DeleteExam deletes an exam, its questions, and submissions.
func DeleteExam(examID string, c *gin.Context) error {
	ctx := c.Request.Context()
	services := config.GetServices(c)

	examRef := services.Firebase.DB.Collection("exams").Doc(examID)

	// Delete questions
	qIter := examRef.Collection("questions").Documents(ctx)
	for {
		doc, err := qIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			break
		}
		doc.Ref.Delete(ctx)
	}

	// Delete submissions
	sIter := examRef.Collection("submissions").Documents(ctx)
	for {
		doc, err := sIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			break
		}
		doc.Ref.Delete(ctx)
	}

	_, err := examRef.Delete(ctx)
	return err
}

// GetReleasedResult returns a detailed result for a student after the exam is released.
func GetReleasedResult(examID, studentID string, c *gin.Context) (*models.ResultDetail, error) {
	ctx := c.Request.Context()
	services := config.GetServices(c)

	examDoc := services.Firebase.DB.Collection("exams").Doc(examID)
	examSnap, err := examDoc.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("exam not found: %w", err)
	}
	var exam models.Exam
	if err := examSnap.DataTo(&exam); err != nil {
		log.Print(err)
		return nil, err
	}

	if !exam.Released {
		return nil, errors.New("results not released yet")
	}

	subDoc := examDoc.Collection("submissions").Doc(studentID)
	subSnap, err := subDoc.Get(ctx)
	if err != nil {
		log.Print(err)
		return nil, fmt.Errorf("submission not found: %w", err)
	}
	var sub models.Submission
	if err := subSnap.DataTo(&sub); err != nil {
		return nil, err
	}

	if !sub.Released {
		return nil, errors.New("student result not released yet")
	}

	qIter := examDoc.Collection("questions").Documents(ctx)
	questions := make(map[string]models.Question)
	var totalPoints float64
	for {
		doc, err := qIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var q models.Question
		if err := doc.DataTo(&q); err != nil {
			continue
		}
		questions[q.QID] = q
		totalPoints += q.Points
	}

	correctCount := 0
	wrongCount := 0

	for qid, q := range questions {
		ans, ok := sub.Answers[qid]
		var studentResp string
		var awarded float64
		if ok {
			studentResp = ans.Response
			awarded = ans.AutoScore + ans.ManualScore
		} else {
			studentResp = ""
			awarded = 0
		}

		isCorrect := false
		switch q.Type {
		case models.MCQ, models.TF:
			stud := strings.TrimSpace(strings.ToLower(studentResp))
			corr := strings.TrimSpace(strings.ToLower(fmt.Sprint(q.CorrectAnswer)))
			if stud != "" && stud == corr {
				isCorrect = true
			}
		default:
			if awarded >= q.Points {
				isCorrect = true
			}
		}

		if isCorrect {
			correctCount++
		} else {
			wrongCount++
		}

	}

	finalScore := sub.AutoScore
	var percentage float64
	if totalPoints > 0 {
		percentage = (finalScore / totalPoints) * 100
	}

	stats := models.ResultStats{
		TotalQuestions: len(questions),
		Correct:        correctCount,
		Wrong:          wrongCount,
		TotalPoints:    totalPoints,
		FinalScore:     finalScore,
		Percentage:     percentage,
	}

	result := models.ResultDetail{
		Exam:       exam,
		Submission: sub,
		Stats:      stats,
	}

	return &result, nil
}

// GetAllReleasedResultsForStudent returns summaries of all released results for a student.
func GetAllReleasedResultsForStudent(studentID string, c *gin.Context) ([]models.ResultSummary, error) {
	ctx := c.Request.Context()
	services := config.GetServices(c)

	examsSnap, err := services.Firebase.DB.Collection("exams").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	var results []models.ResultSummary

	for _, examDoc := range examsSnap {
		examID := examDoc.Ref.ID

		var exam models.Exam
		if err := examDoc.DataTo(&exam); err != nil {
			continue
		}

		subSnap, err := services.Firebase.DB.Collection("exams").
			Doc(examID).
			Collection("submissions").
			Doc(studentID).
			Get(ctx)

		if err != nil {
			continue
		}

		var sub models.Submission
		if err := subSnap.DataTo(&sub); err != nil {
			continue
		}

		if !sub.Released {
			continue
		}

		totalPoints := 0.0
		correct := 0
		wrong := 0

		qsSnap, _ := services.Firebase.DB.Collection("exams").Doc(examID).Collection("questions").Documents(ctx).GetAll()
		for _, q := range qsSnap {
			var qq models.Question
			q.DataTo(&qq)

			totalPoints += qq.Points

			ans, ok := sub.Answers[qq.QID]
			if !ok {
				wrong++
				continue
			}

			if ans.Response == qq.CorrectAnswer {
				correct++
			} else {
				wrong++
			}

		}

		percentage := (sub.AutoScore / totalPoints) * 100.0

		results = append(results, models.ResultSummary{
			ExamID:      examID,
			Title:       exam.Title,
			Date:        exam.StartTime,
			FinalScore:  sub.AutoScore,
			TotalPoints: totalPoints,
			Percentage:  percentage,
		})
	}

	return results, nil
}

// SubmittedExamWithAnswers represents an exam submission with detailed answer analysis
type SubmittedExamWithAnswers struct {
	Exam           models.Exam             `json:"exam"`
	Submission     models.Submission       `json:"submission"`
	Questions      []models.Question       `json:"questions"`
	AnswerAnalysis map[string]AnswerDetail `json:"answer_analysis"`
	Stats          models.ResultStats      `json:"stats"`
}

// AnswerDetail provides information about a student's answer
type AnswerDetail struct {
	Question      models.Question `json:"question"`
	StudentAnswer string          `json:"student_answer"`
	CorrectAnswer string          `json:"correct_answer"`
	IsCorrect     bool            `json:"is_correct"`
	AutoScore     float64         `json:"auto_score"`
	ManualScore   float64         `json:"manual_score"`
}

// SubmittedExamSummary represents a simple exam submission summary without detailed answers
type SubmittedExamSummary struct {
	Exam       models.Exam       `json:"exam"`
	Submission models.Submission `json:"submission"`
}

// GetStudentSubmittedExams retrieves exam metadata for all exams submitted by a student
func GetStudentSubmittedExams(studentID string, c *gin.Context) ([]models.Exam, error) {
	ctx := c.Request.Context()
	services := config.GetServices(c)

	// Get all exams
	examsIter := services.Firebase.DB.Collection("exams").Documents(ctx)
	var submittedExams []models.Exam

	for {
		examDoc, err := examsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var exam models.Exam
		if err := examDoc.DataTo(&exam); err != nil {
			continue
		}

		// Check if student submitted this exam
		subSnap, err := examDoc.Ref.Collection("submissions").Doc(studentID).Get(ctx)
		if err != nil || !subSnap.Exists() {
			continue // Student didn't submit this exam
		}

		// Only return exam metadata
		submittedExams = append(submittedExams, exam)
	}

	return submittedExams, nil
}

// GetStudentExamResultDetails retrieves detailed results for a specific exam submission
func GetStudentExamResultDetails(studentID, examID string, c *gin.Context) (*SubmittedExamWithAnswers, error) {
	ctx := c.Request.Context()
	services := config.GetServices(c)

	examDoc := services.Firebase.DB.Collection("exams").Doc(examID)
	examSnap, err := examDoc.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("exam not found: %w", err)
	}

	var exam models.Exam
	if err := examSnap.DataTo(&exam); err != nil {
		return nil, err
	}

	// Check if student submitted this exam
	subSnap, err := examDoc.Collection("submissions").Doc(studentID).Get(ctx)
	if err != nil || !subSnap.Exists() {
		return nil, errors.New("submission not found")
	}

	var submission models.Submission
	if err := subSnap.DataTo(&submission); err != nil {
		return nil, err
	}

	// Get all questions for this exam
	qIter := examDoc.Collection("questions").Documents(ctx)
	questions := make(map[string]models.Question)
	var questionList []models.Question
	var totalPoints float64

	for {
		qDoc, err := qIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			break
		}

		var q models.Question
		if err := qDoc.DataTo(&q); err != nil {
			continue
		}
		questions[q.QID] = q
		questionList = append(questionList, q)
		totalPoints += q.Points
	}

	// Analyze answers
	answerAnalysis := make(map[string]AnswerDetail)
	correctCount := 0
	wrongCount := 0

	for qid, q := range questions {
		ans, hasAnswer := submission.Answers[qid]
		studentResp := ""
		autoScore := 0.0
		manualScore := 0.0
		isCorrect := false

		if hasAnswer {
			studentResp = ans.Response
			autoScore = ans.AutoScore
			manualScore = ans.ManualScore
		}

		// Determine if answer is correct
		switch q.Type {
		case models.MCQ, models.TF:
			stud := strings.TrimSpace(strings.ToLower(studentResp))
			corr := strings.TrimSpace(strings.ToLower(fmt.Sprint(q.CorrectAnswer)))
			if stud != "" && stud == corr {
				isCorrect = true
			}
		default:
			// For other types, check if points were awarded
			if autoScore+manualScore >= q.Points {
				isCorrect = true
			}
		}

		if isCorrect {
			correctCount++
		} else {
			wrongCount++
		}

		answerAnalysis[qid] = AnswerDetail{
			Question:      q,
			StudentAnswer: studentResp,
			CorrectAnswer: q.CorrectAnswer,
			IsCorrect:     isCorrect,
			AutoScore:     autoScore,
			ManualScore:   manualScore,
		}
	}

	finalScore := submission.AutoScore
	var percentage float64
	if totalPoints > 0 {
		percentage = (finalScore / totalPoints) * 100
	}

	stats := models.ResultStats{
		TotalQuestions: len(questions),
		Correct:        correctCount,
		Wrong:          wrongCount,
		TotalPoints:    totalPoints,
		FinalScore:     finalScore,
		Percentage:     percentage,
	}

	return &SubmittedExamWithAnswers{
		Exam:           exam,
		Submission:     submission,
		Questions:      questionList,
		AnswerAnalysis: answerAnalysis,
		Stats:          stats,
	}, nil
}
