package services

import (
	"HOSEROF_API/config"
	"HOSEROF_API/models"
	"context"
	"errors"
	"fmt"
)

func SignupUser(newUser models.NewUser) error {
	hashed, err := HashPassword(newUser.NewStudentPassword)

	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"student_id":          newUser.NewStudentID,
		"student_name":        newUser.NewStudentName,
		"student_phonenumber": newUser.NewStudentPhoneNumber,
		"student_password":    hashed,
		"student_age":         newUser.NewStudentAge,
		"student_grade":       newUser.NewStudentGrade,
		"student_class":       newUser.NewStudentClass,
	}
	_, err = config.DB.Collection("students").Doc(newUser.NewStudentID).Set(context.Background(), data)
	if err != nil {
		return err
	}

	return nil
}

func LoginUser(login models.UserLogin) (*models.UserDataResponse, error) {
	ctx := context.Background()
	docRef := config.DB.Collection("students").Doc(login.StudentId)
	docSnap, err := docRef.Get(ctx)

	if err != nil {
		return nil, errors.New("user not found")
	}

	var fsUser models.UserFirestore

	if err := docSnap.DataTo(&fsUser); err != nil {
		return nil, errors.New("failed to parse user data")
	}

	if !CheckPasswordHash(fsUser.StudentPassword, login.StudentPassword) {
		return nil, errors.New("invalid password")
	}
	token, err := jwtGenerator(fsUser.StudentID, fsUser.StudentClass, fsUser.StudentName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	resp := &models.UserDataResponse{
		StudentToken: token,
		StudentId:    fsUser.StudentID,
		StudentName:  fsUser.StudentName,
		StudentClass: fsUser.StudentClass,
	}
	return resp, nil
}
