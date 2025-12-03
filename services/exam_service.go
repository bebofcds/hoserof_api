package services

import (
	"HOSEROF_API/config"
	"HOSEROF_API/models"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CreateExam(exam models.Exam, questions []models.Question) (string, error) {
	ctx := context.Background()
	exams := config.DB.Collection("exams")
	doc := exams.NewDoc()
	exam.ExamID = doc.ID
	exam.CreatedAt = time.Now()
	exam.Released = false

	_, err := doc.Set(ctx, exam)
	if err != nil {
		return "", err
	}

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

func GetExamsForClass(class string, studentID string) ([]models.Exam, error) {
	ctx := context.Background()

	q := config.DB.Collection("exams").
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

		if now.Before(e.StartTime) {
			continue
		}
		if now.After(e.EndTime) {
			continue
		}

		subSnap, err := doc.Ref.Collection("submissions").Doc(studentID).Get(ctx)
		if err == nil && subSnap.Exists() {
			continue
		}

		out = append(out, e)
	}

	return out, nil
}

func GetExamQuestions(examID string, forStudent bool) ([]models.Question, error) {
	ctx := context.Background()

	qIter := config.DB.Collection("exams").
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

		if forStudent {
			q.CorrectAnswer = ""
		}

		qs = append(qs, q)
	}

	return qs, nil
}

func SubmitExam(examID string, studentID string, answers map[string]models.Answer) error {
	ctx := context.Background()
	examDoc := config.DB.Collection("exams").Doc(examID)

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
	var manualScore float64 = 0
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

	final := autoScore + manualScore

	submission := models.Submission{
		StudentID:   studentID,
		StartedAt:   now,
		SubmittedAt: now,
		Answers:     answers,
		AutoScore:   autoScore,
		ManualScore: manualScore,
		FinalScore:  final,
		Graded:      false,
		Released:    false,
	}

	_, err = subDoc.Set(ctx, submission)
	if err != nil {
		return err
	}

	return nil
}

func GetSubmission(examID, studentID string) (models.Submission, error) {
	ctx := context.Background()
	doc := config.DB.Collection("exams").Doc(examID).Collection("submissions").Doc(studentID)
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

func GetAllSubmissions(examID string) ([]models.Submission, error) {
	ctx := context.Background()

	iter := config.DB.Collection("exams").
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

func GradeWrittenAnswer(examID, studentID, qid string, score float64) error {
	ctx := context.Background()
	subRef := config.DB.Collection("exams").Doc(examID).Collection("submissions").Doc(studentID)

	snap, err := subRef.Get(ctx)
	if err != nil {
		return err
	}
	var s models.Submission
	if err := snap.DataTo(&s); err != nil {
		return err
	}

	ans, ok := s.Answers[qid]
	if !ok {
		return errors.New("question not found in submission")
	}

	ans.ManualScore = score
	s.Answers[qid] = ans

	var manualSum float64
	for _, a := range s.Answers {
		manualSum += a.ManualScore
	}
	s.ManualScore = manualSum
	s.FinalScore = s.AutoScore + s.ManualScore
	s.Graded = true

	_, err = subRef.Set(ctx, s, firestore.MergeAll)
	return err
}

func ReleaseResults(examID string) error {
	ctx := context.Background()
	examRef := config.DB.Collection("exams").Doc(examID)
	_, err := examRef.Update(ctx, []firestore.Update{
		{Path: "released", Value: true},
	})
	if err != nil {
		return err
	}

	subs := examRef.Collection("submissions").Documents(ctx)
	for {
		d, err := subs.Next()
		if err != nil {
			break
		}
		_, _ = d.Ref.Update(ctx, []firestore.Update{
			{Path: "released", Value: true},
		})
	}

	return nil
}
func GetReleasedResult(examID, studentID string) (*models.ResultDetail, error) {
	ctx := context.Background()
	client := config.DB

	examDoc := client.Collection("exams").Doc(examID)
	examSnap, err := examDoc.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("exam not found: %w", err)
	}
	var exam models.Exam
	if err := examSnap.DataTo(&exam); err != nil {
		return nil, err
	}

	if !exam.Released {
		return nil, errors.New("results not released yet")
	}

	subDoc := examDoc.Collection("submissions").Doc(studentID)
	subSnap, err := subDoc.Get(ctx)
	if err != nil {
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

	reviews := make([]models.QuestionReview, 0, len(questions))
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

		var correctAnsForReturn string
		if !isCorrect {
			correctAnsForReturn = q.CorrectAnswer
		}

		reviews = append(reviews, models.QuestionReview{
			QID:           q.QID,
			Type:          string(q.Type),
			QuestionText:  q.QuestionText,
			Options:       q.Options,
			StudentAnswer: studentResp,
			CorrectAnswer: correctAnsForReturn,
			IsCorrect:     isCorrect,
			PointsAwarded: awarded,
			MaxPoints:     q.Points,
			ImageURL:      q.ImageURL,
		})
	}

	finalScore := sub.AutoScore + sub.ManualScore
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
		Reviews:    reviews,
		Stats:      stats,
	}

	return &result, nil
}

func GetAllReleasedResultsForStudent(studentID string) ([]models.ResultSummary, error) {
	ctx := context.Background()
	client := config.DB

	examsSnap, err := client.Collection("exams").Documents(ctx).GetAll()
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

		subSnap, err := client.Collection("exams").
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

		qsSnap, _ := client.Collection("exams").Doc(examID).Collection("questions").Documents(ctx).GetAll()
		for _, q := range qsSnap {
			var qq models.Question
			q.DataTo(&qq)

			totalPoints += qq.Points

			ans, ok := sub.Answers[qq.QID]
			if !ok {
				wrong++
				continue
			}

			if qq.Type == models.WRITTEN {
				if ans.AutoScore+ans.ManualScore > 0 {
					correct++
				} else {
					wrong++
				}
			} else {
				if ans.Response == qq.CorrectAnswer {
					correct++
				} else {
					wrong++
				}
			}
		}

		finalScore := sub.AutoScore + sub.ManualScore
		percentage := (finalScore / totalPoints) * 100.0

		results = append(results, models.ResultSummary{
			ExamID:      examID,
			Title:       exam.Title,
			Date:        exam.StartTime,
			FinalScore:  finalScore,
			TotalPoints: totalPoints,
			Percentage:  percentage,
		})
	}

	return results, nil
}
