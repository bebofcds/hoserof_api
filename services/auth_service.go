/*
================================================================================
HOSEROF_API - User Services
================================================================================

Description:
This package provides service-level functions for creating and managing
students and staff, as well as handling user login. It interacts with
Firebase Firestore for data storage and the JWT service for authentication.

Responsibilities:
1. CreateStudent   - Adds a new student to Firestore.
2. CreateStaff     - Adds a new staff member with hashed password.
3. LoginUser       - Authenticates a user (student or staff) and returns a JWT.

Usage Notes:
- All functions require a Gin context (`*gin.Context`) to extract services.
- Firestore collections used:
    - `students` → stores student documents
    - `staff`    → stores staff documents
- JWT is generated using `services.JWT` injected via context.
- Passwords for staff are hashed using `HashPassword` before storage.
- `LoginUser` checks both students and staff collections and returns errors
  if login is invalid.

Error Handling:
- Returns descriptive errors for missing password, invalid login, or user
  not found.
- Firestore errors are returned directly for logging or response handling.

Security Notes:
- Staff passwords must be stored hashed.
- JWT tokens should be sent to clients securely.

Date Format Reference:
- Firestore document IDs correspond to `student ID` or `staff ID`.
================================================================================
*/

package services

import (
	"HOSEROF_API/config"
	"HOSEROF_API/models"
	"context"
	"errors"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateStudent adds a new student to the Firestore `students` collection.
// Fields include student ID, name, phone number, age, grade, class, and role.
func CreateStudent(newUser models.NewUser, c *gin.Context) error {
	services := config.GetServices(c)

	data := map[string]interface{}{
		"student_id":           newUser.NewStudentID,
		"student_name":         newUser.NewStudentName,
		"student_phonenumber":  newUser.NewStudentPhoneNumber,
		"student_age":          newUser.NewStudentAge,
		"student_grade":        newUser.NewStudentGrade,
		"student_class":        newUser.NewStudentClass,
		"role":                 newUser.NewStudentRole,
		"attended_days":        0,
		"last_attendance_date": "",
		"total_days":           0,
	}

	_, err := services.Firebase.DB.Collection("students").
		Doc(newUser.NewStudentID).
		Set(context.Background(), data)

	if err != nil {
		return err
	}

	return nil
}

// CreateStaff adds a new staff member to the Firestore `staff` collection.
// Staff password is hashed before saving.
func CreateStaff(newStaff models.NewStaff, c *gin.Context) error {
	services := config.GetServices(c)

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

	_, err = services.Firebase.DB.Collection("staff").
		Doc(newStaff.ID).
		Set(context.Background(), data)

	if err != nil {
		return err
	}

	return nil
}

// LoginUser authenticates a student or staff and returns a JWT token.
// - For students: token is generated directly, no password check.
// - For staff: password is required and checked against hashed value.
func LoginUser(login models.UserLogin, c *gin.Context) (*models.UserDataResponse, error) {
	services := config.GetServices(c)
	ctx := c.Request.Context()

	// Check if user is a student
	studentDoc, err := services.Firebase.DB.Collection("students").Doc(login.ID).Get(ctx)
	if err == nil && studentDoc.Exists() {
		var fsUser models.UserFirestore
		studentDoc.DataTo(&fsUser)

		jwtService := services.JWT.(*JWTService)
		token, err := jwtService.GenerateToken(fsUser.StudentID, fsUser.StudentClass, fsUser.Role, fsUser.StudentName)
		if err != nil {
			return nil, err
		}

		return &models.UserDataResponse{
			Token: token,
			Id:    fsUser.StudentID,
			Name:  fsUser.StudentName,
			Class: fsUser.StudentClass,
			Role:  fsUser.Role,
		}, nil
	}

	// Check if user is staff
	staffDoc, err := services.Firebase.DB.Collection("staff").Doc(login.ID).Get(ctx)
	if err == nil && staffDoc.Exists() {
		if login.Password == "" {
			return nil, errors.New("PASSWORD_REQUIRED")
		}

		var staff models.StaffFirestore
		staffDoc.DataTo(&staff)

		if !CheckPasswordHash(staff.Password, login.Password) {
			return nil, errors.New("INVALID_LOGIN")
		}

		jwtService := services.JWT.(*JWTService)
		token, err := jwtService.GenerateToken(staff.ID, staff.Class, staff.Role, staff.Name)
		if err != nil {
			return nil, err
		}

		return &models.UserDataResponse{
			Token: token,
			Id:    staff.ID,
			Name:  staff.Name,
			Class: staff.Class,
			Role:  staff.Role,
		}, nil
	}

	return nil, errors.New("USER_DOES_NOT_EXIST")

}

// UpdateStudent updates an existing student's information in Firestore.
func UpdateStudent(studentID string, updateData models.UpdateStudent, c *gin.Context) error {
	services := config.GetServices(c)

	var updates []firestore.Update

	if updateData.StudentName != "" {
		updates = append(updates, firestore.Update{Path: "student_name", Value: updateData.StudentName})
	}
	if updateData.StudentPhoneNumber != "" {
		updates = append(updates, firestore.Update{Path: "student_phonenumber", Value: updateData.StudentPhoneNumber})
	}
	if updateData.StudentAge != "" {
		updates = append(updates, firestore.Update{Path: "student_age", Value: updateData.StudentAge})
	}
	if updateData.StudentGrade != "" {
		updates = append(updates, firestore.Update{Path: "student_grade", Value: updateData.StudentGrade})
	}
	if updateData.StudentClass != "" {
		updates = append(updates, firestore.Update{Path: "student_class", Value: updateData.StudentClass})
	}

	if len(updates) == 0 {
		return errors.New("NO_FIELDS_TO_UPDATE")
	}

	_, err := services.Firebase.DB.Collection("students").
		Doc(studentID).
		Update(context.Background(), updates)

	if err != nil {
		return err
	}

	return nil
}

// DeleteStudent removes a student from Firestore.
func DeleteStudent(studentID string, c *gin.Context) error {
	services := config.GetServices(c)

	_, err := services.Firebase.DB.Collection("students").
		Doc(studentID).
		Delete(context.Background())

	if err != nil {
		return err
	}

	return nil
}

// UpdateStaff updates an existing staff member's information in Firestore.
// If password is provided, it will be hashed before saving.
func UpdateStaff(staffID string, updateData models.UpdateStaff, c *gin.Context) error {
	services := config.GetServices(c)

	var updates []firestore.Update

	if updateData.Name != "" {
		updates = append(updates, firestore.Update{Path: "name", Value: updateData.Name})
	}
	if updateData.PhoneNumber != "" {
		updates = append(updates, firestore.Update{Path: "phonenumber", Value: updateData.PhoneNumber})
	}
	if updateData.Class != "" {
		updates = append(updates, firestore.Update{Path: "class", Value: updateData.Class})
	}
	if updateData.Password != "" {
		hashed, err := HashPassword(updateData.Password)
		if err != nil {
			return err
		}
		updates = append(updates, firestore.Update{Path: "password", Value: hashed})
	}

	if len(updates) == 0 {
		return errors.New("NO_FIELDS_TO_UPDATE")
	}

	_, err := services.Firebase.DB.Collection("staff").
		Doc(staffID).
		Update(context.Background(), updates)

	if err != nil {
		return err
	}

	return nil
}

// DeleteStaff removes a staff member from Firestore.
func DeleteStaff(staffID string, c *gin.Context) error {
	services := config.GetServices(c)

	_, err := services.Firebase.DB.Collection("staff").
		Doc(staffID).
		Delete(context.Background())

	if err != nil {
		return err
	}

	return nil
}

// GetStudentByID retrieves a single user by their Firestore ID.
func GetStudentByID(userID string, c *gin.Context) (models.UserFirestore, error) {
	ctx := c.Request.Context()
	services := config.GetServices(c)

	doc, err := services.Firebase.DB.Collection("students").Doc(userID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return models.UserFirestore{}, errors.New("user not found")
		}
		return models.UserFirestore{}, err
	}

	var user models.UserFirestore
	if err := doc.DataTo(&user); err != nil {
		return models.UserFirestore{}, err
	}

	user.StudentID = doc.Ref.ID
	return user, nil
}

// GetStaffByID retrieves a single user by their Firestore ID.
func GetStaffByID(userID string, c *gin.Context) (models.Staff, error) {
	ctx := c.Request.Context()
	services := config.GetServices(c)

	doc, err := services.Firebase.DB.Collection("staff").Doc(userID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return models.Staff{}, errors.New("user not found")
		}
		return models.Staff{}, err
	}

	var user models.Staff
	if err := doc.DataTo(&user); err != nil {
		return models.Staff{}, err
	}

	user.ID = doc.Ref.ID
	return user, nil
}

// GetAllStaff retrieves all staff.
func GetAllStaff(c *gin.Context) ([]models.StaffList, error) {
	ctx := c.Request.Context()
	services := config.GetServices(c)

	iter := services.Firebase.DB.Collection("staff").
		Where("role", "==", "teacher").
		Documents(ctx)

	staff := []models.StaffList{}

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var s models.StaffList
		if err := doc.DataTo(&s); err != nil {
			return nil, err
		}
		s.ID = doc.Ref.ID
		staff = append(staff, s)

	}

	return staff, nil
}
