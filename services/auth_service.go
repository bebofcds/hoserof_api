package services

import (
	"HOSEROF_API/config"
	"HOSEROF_API/models"
	"context"
	"errors"
)

func CreateStudent(newUser models.NewUser) error {
	data := map[string]interface{}{
		"student_id":          newUser.NewStudentID,
		"student_name":        newUser.NewStudentName,
		"student_phonenumber": newUser.NewStudentPhoneNumber,
		"student_age":         newUser.NewStudentAge,
		"student_grade":       newUser.NewStudentGrade,
		"student_class":       newUser.NewStudentClass,
		"role":                newUser.NewStudentRole,
	}

	_, err := config.DB.Collection("students").
		Doc(newUser.NewStudentID).
		Set(context.Background(), data)

	if err != nil {
		return err
	}

	return nil
}

func CreateStaff(newStaff models.NewStaff) error {
	hashed, err := HashPassword(newStaff.Password)
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"id":          newStaff.ID,
		"name":        newStaff.Name,
		"phonenumber": newStaff.PhoneNumber,
		"class":       newStaff.Class,
		"password":    hashed,
		"role":        newStaff.Role,
	}

	_, err = config.DB.Collection("staff").
		Doc(newStaff.ID).
		Set(context.Background(), data)

	if err != nil {
		return err
	}

	return nil
}

func LoginUser(login models.UserLogin) (*models.UserDataResponse, error) {
	ctx := context.Background()
	studentDoc, err := config.DB.Collection("students").Doc(login.ID).Get(ctx)

	if err == nil {
		var fsUser models.UserFirestore
		studentDoc.DataTo(&fsUser)

		token, _ := jwtGenerator(fsUser.StudentID, fsUser.StudentClass, fsUser.Role, fsUser.StudentName)

		return &models.UserDataResponse{
			StudentToken: token,
			StudentId:    fsUser.StudentID,
			StudentName:  fsUser.StudentName,
			StudentClass: fsUser.StudentClass,
			Role:         fsUser.Role,
		}, nil
	}

	if login.Password == "" {
		return nil, errors.New("PASSWORD_REQUIRED")
	}

	staffDoc, err := config.DB.Collection("staff").Doc(login.ID).Get(ctx)
	if err != nil {
		return nil, errors.New("INVALID_LOGIN")
	}

	var staff models.StaffFirestore
	staffDoc.DataTo(&staff)

	if !CheckPasswordHash(staff.Password, login.Password) {
		return nil, errors.New("INVALID_LOGIN")
	}

	token, _ := jwtGenerator(staff.ID, staff.Class, staff.Role, staff.Name)

	return &models.UserDataResponse{
		StudentToken: token,
		StudentId:    staff.ID,
		StudentName:  staff.Name,
		StudentClass: staff.Class,
		Role:         staff.Role,
	}, nil

}
