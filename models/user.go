package models

type UserLogin struct {
	ID       string `json:"user_ID"`
	Password string `json:"password"`
}

type NewUser struct {
	NewStudentID          string `json:"student_id"`
	NewStudentName        string `json:"student_name"`
	NewStudentPhoneNumber string `json:"student_phonenumber"`
	NewStudentAge         string `json:"student_age"`
	NewStudentGrade       string `json:"student_grade"`
	NewStudentClass       string `json:"student_class"`
	NewStudentRole        string `json:"role"`
}

type UserFirestore struct {
	StudentID          string `firestore:"student_id"`
	StudentName        string `firestore:"student_name"`
	StudentClass       string `firestore:"student_class"`
	StudentPhone       string `firestore:"student_phonenumber"`
	StudentAge         string `firestore:"student_age"`
	StudentGrade       string `firestore:"student_grade"`
	Role               string `firestore:"role"`
	TotalDays          int    `firestore:"total_days"`
	AttendedDays       int    `firestore:"attended_days"`
	AbsentDays         int    `firestore:"absent_days"`
	LastAttendanceDate string `firestore:"last_attendance_date"`
}

type UserClassList struct {
	StudentID          string `firestore:"student_id"`
	StudentName        string `firestore:"student_name"`
	StudentPhoneNumber string `firestore:"student_phonenumber"`
}

type UserDataResponse struct {
	Token string `json:"token"`
	Id    string `json:"id"`
	Name  string `json:"name"`
	Class string `json:"class"`
	Role  string `json:"role"`
}

type Staff struct {
	ID    string `firestore:"id"`
	Name  string `firestore:"name"`
	Class string `firestore:"class"`
	Phone string `firestore:"phonenumber"`
	Role  string `firestore:"role"`
}

type NewStaff struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phoneNumber"`
	Class       string `json:"class"`
	Role        string `json:"role"`
	Password    string `json:"password"`
}

type StaffFirestore struct {
	ID       string `firestore:"id"`
	Name     string `firestore:"name"`
	Class    string `firestore:"class"`
	Phone    string `firestore:"phonenumber"`
	Password string `firestore:"password"`
	Role     string `firestore:"role"`
}

type StaffList struct {
	ID    string `firestore:"id"`
	Name  string `firestore:"name"`
	Class string `firestore:"class"`
}

type UpdateStudent struct {
	StudentName        string `json:"student_name"`
	StudentPhoneNumber string `json:"student_phonenumber"`
	StudentAge         string `json:"student_age"`
	StudentGrade       string `json:"student_grade"`
	StudentClass       string `json:"student_class"`
}

type UpdateStaff struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phoneNumber"`
	Class       string `json:"class"`
	Password    string `json:"password"`
}
