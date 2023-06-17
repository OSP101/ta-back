package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


type Subjects struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Term      string    `json:"term"`
	Year      string    `json:"year"`
	Image      string    `json:"image"`
	Section   []string `json:"section"`
}

type Checkname struct {
	Sid string    `json:"sid"`
	Name string    `json:"name"`
	Date time.Time `json:"date"`
	Status string `json:"status"`
	Section string `json:"section"`
	Passcode []string `json:"passcode"`
	Check []*Check `json:"check"`
}

type Check struct {
	Std string `json:"std"`
	Passcodecheck string `json:"passcodecheck"`
	Timestamp time.Time `json:"timestamp"`
}

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Subject   []*Subject  `json:"subject"`
}

type Subject struct{
	ID        string    `json:"id"`
	Section   string    `json:"section"`
	Type      string    `json:"type"`
	Image      string    `json:"image"`
}


var (
	client             *mongo.Client
	collection_user    *mongo.Collection
	collection_subject *mongo.Collection
	collection_checkname *mongo.Collection
	ctx                context.Context
)



func main() {

	
	ctx, _ = context.WithTimeout(context.Background(), 1000*time.Second)
	// Connect to MongoDB
	connectToDB()
	// Create router
	router := mux.NewRouter()


	// Define API routes
	router.HandleFunc("/users", getUsersHandler).Methods("GET")
	router.HandleFunc("/user/{email}", getUserHandler).Methods("GET")
	router.HandleFunc("/user/{email}", deleteUserHandler).Methods("DELETE")
	router.HandleFunc("/user", addUserHandler).Methods("POST")
	router.HandleFunc("/user/{id}", updateUserHandler).Methods("PUT")

	router.HandleFunc("/user/{email}/subject", addSubjectToUserHandler).Methods("POST")
	router.HandleFunc("/user/{email}/subject/{subject}", removeSubjectFromUserHandler).Methods("DELETE")

	router.HandleFunc("/subjects", getSubjectsHandler).Methods("GET")
	router.HandleFunc("/subject/{id}", getSubjectByIDHandler).Methods("GET")
	router.HandleFunc("/subject/{id}", updateSubjectHandler).Methods("PUT")
	router.HandleFunc("/subject", addSubjectHandler).Methods("POST")
	router.HandleFunc("/subject/{id}", deleteSubjectHandler).Methods("DELETE")

	router.HandleFunc("/check/{id}", getAllChecknamesHandler).Methods("GET")
	router.HandleFunc("/check", addChecknameHandler).Methods("POST")
	router.HandleFunc("/check/{name}", deleteChecknameHandler).Methods("DELETE")

	router.HandleFunc("/check/{name}/std", addCheckToChecknameHandler).Methods("POST")
	router.HandleFunc("/checkname/{checkname_id}/check/{check_id}", deleteCheckFromChecknameHandler).Methods("DELETE")
	router.HandleFunc("/checkname/{checkname_id}/check", getAllChecksFromChecknameHandler).Methods("GET")


	corsServer := enableCORS(router)

	log.Fatal(http.ListenAndServe(":8000", corsServer))
}

func connectToDB() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	collection_user = client.Database("TA-DB").Collection("Users")
	collection_subject = client.Database("TA-DB").Collection("Subject")
	collection_checkname = client.Database("TA-DB").Collection("Checkname")
}

func enableCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		handler.ServeHTTP(w, r)
	})
}




// เพิ่มผู้ใช้ใหม่
func addUserHandler(w http.ResponseWriter, r *http.Request) {
	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = collection_user.InsertOne(ctx, newUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("User added successfully")
}

// อัปเดตผู้ใช้
func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["id"]

	var updatedUser User
	err := json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}


	filter := bson.M{"id": userID}
	update := bson.M{
		"$set": bson.M{
			"name":       updatedUser.Name,
			"email":      updatedUser.Email,
			"subject":    updatedUser.Subject,
		},
	}

	_, err = collection_user.UpdateOne(ctx, filter, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("User updated successfully")
}

// ลบผู้ใช้
func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["id"]

	_, err := collection_user.DeleteOne(ctx, bson.M{"id": userID})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("User deleted successfully")
}

// ดึงข้อมูลผู้ใช้ทั้งหมด
func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	var users []User

	cur, err := collection_user.Find(ctx, bson.M{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var user User
		err := cur.Decode(&user)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		users = append(users, user)
	}

	if err := cur.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// ดึงข้อมูลผู้ใช้แต่ละคน
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["email"]

	var user User
	err := collection_user.FindOne(ctx, bson.M{"email": userID}).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}


// เพิ่มวิชาให้กับผู้ใช้
func addSubjectToUserHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userEmail := params["email"]

	var subject Subject
	err := json.NewDecoder(r.Body).Decode(&subject)
	if err != nil {
		log.Fatal(err)
	}

	// หาผู้ใช้จากอีเมล
	var user User
	err = collection_user.FindOne(ctx, bson.M{"email": userEmail}).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}

	// เพิ่มวิชาลงในอาร์เรย์ของผู้ใช้
	user.Subject = append(user.Subject, &subject)

	// อัปเดตผู้ใช้ในฐานข้อมูล
	update := bson.M{
		"$set": bson.M{"subject": user.Subject},
	}
	_, err = collection_user.UpdateOne(ctx, bson.M{"email": userEmail}, update)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Subject added to user successfully")
}

// ลบวิชาออกจากผู้ใช้
func removeSubjectFromUserHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userEmail := params["email"]
	subjectID := params["subject"]

	// หาผู้ใช้จากอีเมล
	var user User
	err := collection_user.FindOne(ctx, bson.M{"email": userEmail}).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}

	// ลบวิชาออกจากอาร์เรย์ของผู้ใช้
	for i, subj := range user.Subject {
		if subj != nil && subj.ID == subjectID {
			user.Subject = append(user.Subject[:i], user.Subject[i+1:]...)
			break
		}
	}

	// อัปเดตผู้ใช้ในฐานข้อมูล
	update := bson.M{
		"$set": bson.M{"subject": user.Subject},
	}
	_, err = collection_user.UpdateOne(ctx, bson.M{"email": userEmail}, update)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Subject removed from user successfully")
}




// Subject
func getSubjectsHandler(w http.ResponseWriter, r *http.Request) {
	cur, err := collection_subject.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)

	var subjectss []Subjects
	for cur.Next(ctx) {
		var subject Subjects
		err := cur.Decode(&subject)
		if err != nil {
			log.Fatal(err)
		}
		subjectss = append(subjectss, subject)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subjectss)
}

func getSubjectByIDHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	subjectID := params["id"]

	var subjects Subjects
	err := collection_subject.FindOne(ctx, bson.M{"id": subjectID}).Decode(&subjects)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subjects)
}



func updateSubjectHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	subjectID := params["id"]
	var updatedSubject Subjects
	err := json.NewDecoder(r.Body).Decode(&updatedSubject)
	if err != nil {
		log.Fatal(err)
	}

	update := bson.M{
		"$set": updatedSubject,
	}
	_, err = collection_subject.UpdateOne(ctx, bson.M{"id": subjectID}, update)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Subject updated successfully")
}

func addSubjectHandler(w http.ResponseWriter, r *http.Request) {
	var newSubject Subjects
	err := json.NewDecoder(r.Body).Decode(&newSubject)
	if err != nil {
		log.Fatal(err)
	}


	_, err = collection_subject.InsertOne(ctx, newSubject)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Subject added successfully")
}

func deleteSubjectHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	subjectID := params["id"]
	result, err := collection_subject.DeleteOne(ctx, bson.M{"id": subjectID})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if result.DeletedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ดึงวิชาจากผู้ใช้

func getSubjectinUserHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userEmail := params["email"]

	var user User
	err := collection_user.FindOne(ctx, bson.M{"email": userEmail}).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(user.Subject)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user.Subject)
}


// Check
// func getAllChecknamesHandler(w http.ResponseWriter, r *http.Request) {
// 	cur, err := collection_checkname.Find(context.TODO(), bson.M{})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer cur.Close(context.TODO())

// 	var checknames []Checkname
// 	for cur.Next(context.TODO()) {
// 		var checkname Checkname
// 		err := cur.Decode(&checkname)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		checknames = append(checknames, checkname)
// 	}

// 	if err := cur.Err(); err != nil {
// 		log.Fatal(err)
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(checknames)
// }

// ดึงข้อมูล Check
func getAllChecknamesHandler(w http.ResponseWriter, r *http.Request) {
	// รับพารามิเตอร์ที่ต้องการจาก Query String
	params := mux.Vars(r)
	userID := params["id"]

	var newCheckname Checkname
	err := collection_checkname.FindOne(ctx, bson.M{"sid": userID}).Decode(&newCheckname)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newCheckname)
}




// เพิ่ม Check
func addChecknameHandler(w http.ResponseWriter, r *http.Request) {
	var newCheckname Checkname
	err := json.NewDecoder(r.Body).Decode(&newCheckname)
	if err != nil {
		log.Fatal(err)
	}

	_, err = collection_checkname.InsertOne(context.TODO(), newCheckname)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Checkname added successfully")
}

// ลบ Check
func deleteChecknameHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	checknameID := params["name"]

	result, err := collection_checkname.DeleteOne(context.TODO(), bson.M{"name": checknameID})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Checkname delete successfully")

}


func addCheckToChecknameHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	checknameID := params["name"]

	var newCheck Check
	err := json.NewDecoder(r.Body).Decode(&newCheck)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newCheck.Timestamp = time.Now().In(time.FixedZone("Asia/Bangkok", 7*60*60))

	// เช็คสถานะใน struct Checkname
	checkname := Checkname{}
	err = collection_checkname.FindOne(ctx, bson.M{"name": checknameID}).Decode(&checkname)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if checkname.Status == "off" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("ไม่รับการเช็คชื่อแล้ว")
		return
	}

	// เช็คว่า Passcode ตรงกับ Passcode ล่าสุดหรือไม่
	if len(checkname.Check) > 0 {
		latestCheck := checkname.Check[len(checkname.Check)-1]
		if len(latestCheck.Passcodecheck) > 0 {
			latestPasscode := latestCheck.Passcodecheck
			if newCheck.Passcodecheck != latestPasscode {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode("Passcode ไม่ถูกต้อง")
				return
			}
		}
	}

	update := bson.M{
		"$push": bson.M{"check": newCheck},
	}

	_, err = collection_checkname.UpdateOne(ctx, bson.M{"name": checknameID}, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Checkname added successfully")
}


// ลบการเช็คของคนๆนั้น
func deleteCheckFromChecknameHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	checknameID := params["checkname_id"]
	checkID := params["check_id"]

	update := bson.M{
		"$pull": bson.M{"check": bson.M{"std": checkID}},
	}
	_, err := collection_checkname.UpdateOne(ctx, bson.M{"name": checknameID}, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Checkname delete successfully")
}

// ดึงข้อมูลการเช็คทุกคน
func getAllChecksFromChecknameHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	checknameID := params["checkname_id"]

	var checkname Checkname

	err := collection_checkname.FindOne(ctx, bson.M{"name": checknameID}).Decode(&checkname)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Print(time.Now().In(time.FixedZone("Asia/Bangkok", 7*60*60)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(checkname.Check)
}

