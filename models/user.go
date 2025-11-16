package models

type UserLogin struct {
	StudentId       string `json:"user_ID"`
	StudentPassword string `json:"user_password"`
}

type NewUser struct {
	NewStudentID          string `json:"student_id"`
	NewStudentName        string `json:"student_name"`
	NewStudentPhoneNumber string `json:"student_phonenumber"`
	NewStudentPassword    string `json:"student_password"`
	NewStudentAge         string `json:"student_age"`
	NewStudentGrade       string `json:"student_grade"`
	NewStudentClass       string `json:"student_class"`
}

type UserFirestore struct {
	StudentID       string `firestore:"student_id"`
	StudentPassword string `firestore:"student_password"`
	StudentName     string `firestore:"student_name"`
	StudentClass    string `firestore:"student_class"`
	StudentPhone    string `firestore:"student_phonenumber"`
	StudentAge      string `firestore:"student_age"`
	StudentGrade    string `firestore:"student_grade"`
}
type UserDataResponse struct {
	StudentToken string `json:"student_token"`
	StudentId    string `json:"student_id"`
	StudentName  string `json:"student_name"`
	StudentClass string `json:"student_class"`
}
