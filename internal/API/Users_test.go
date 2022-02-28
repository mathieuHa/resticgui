package API

// func TestQueryUsersRoute(t *testing.T) {
// 	Utils.TestConfig()
// 	Database.Migrate()
// 	db := Database.Connect()
// 	router := SetupRouter()
// 	Utils.CleanDatabaseTest(db, t, "users")

// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
// 	router.ServeHTTP(w, req)

// 	assert.Equal(t, 200, w.Code)
// 	assert.Equal(t, "[]", w.Body.String())
// }

// func TestGetUsersRoute(t *testing.T) {
// 	Utils.TestConfig()
// 	Database.Migrate()
// 	db := Database.Connect()
// 	router := SetupRouter()
// 	Utils.CleanDatabaseTest(db, t, "users")

// 	// Check invalid ID
// 	id := "4ed79fc4"
// 	url := fmt.Sprintf("/api/v1/users/%s", id)
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", url, nil)
// 	router.ServeHTTP(w, req)

// 	b := "{\n    \"error\": \"id not valid\"\n}"

// 	assert.Equal(t, http.StatusBadRequest, w.Code)
// 	assert.Equal(t, b, w.Body.String())

// 	// Check Unknown UUID
// 	id = "132319"
// 	url = fmt.Sprintf("/api/v1/users/%s", id)
// 	w = httptest.NewRecorder()
// 	req, _ = http.NewRequest("GET", url, nil)
// 	router.ServeHTTP(w, req)

// 	b = "{\n    \"error\": \"user not found\"\n}"

// 	// Check Valid ID get
// 	userEmail := "myEmail"
// 	user := Models.User{Email: userEmail}
// 	result := db.Create(&user)
// 	assert.Nil(t, result.Error)
// 	id2 := user.ID
// 	log.Debug().Msgf("%v", id2)

// 	url = fmt.Sprintf("/api/v1/users/%v", id2)
// 	w = httptest.NewRecorder()
// 	req, _ = http.NewRequest("GET", url, nil)
// 	router.ServeHTTP(w, req)

// 	log.Debug().Msg(w.Body.String())
// 	var userReceived Models.User
// 	err := json.Unmarshal(w.Body.Bytes(), &userReceived)
// 	log.Debug().Err(err)

// 	assert.Nil(t, err)
// 	assert.Equal(t, userReceived.ID, id2)
// 	assert.Equal(t, userReceived.Email, userEmail)

// 	assert.Equal(t, http.StatusOK, w.Code)
// }
