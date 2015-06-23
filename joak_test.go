package joak

import(
	//`log`
	`time`
	`regexp`
	`testing`
	`net/http`
	`appengine/aetest`
	`github.com/gorilla/mux`
	`google.golang.org/appengine`
	//`google.golang.org/appengine/datastore`
	`github.com/stretchr/testify/assert`
)

func Test_RouteLocalTest(t *testing.T){
	RouteLocalTest(mux.NewRouter(), nil, 300, `test`, &testEntity{}, nil, nil, nil)
}

func Test_RouteGaeProd(t *testing.T){
	c, _ := aetest.NewContext(nil)
	ctx := appengine.NewContext(c.Request().(*http.Request))
	dur1, _ := time.ParseDuration(`-1s`)

	err := RouteGaeProd(mux.NewRouter(), nil, 300, `test`, &testEntity{}, nil, nil, nil, dur1, dur1, ``, ctx, ``, ``, ``, ``)

	assert.Equal(t, `kind must not be an empty string`, err.Error(), `err should contain appropriate message`)

	err = RouteGaeProd(mux.NewRouter(), nil, 300, `test`, &testEntity{}, nil, nil, nil, dur1, dur1, `test`, ctx, ``, ``, ``, ``)

	assert.Equal(t, `deleteAfter must be a positive time.Duration`, err.Error(), `err should contain appropriate message`)

	dur2, _ := time.ParseDuration(`1s`)

	err = RouteGaeProd(mux.NewRouter(), nil, 300, `test`, &testEntity{}, nil, nil, nil, dur2, dur1, `test`, ctx, ``, ``, ``, ``)

	assert.Equal(t, `clearOutAfter must be a positive time.Duration`, err.Error(), `err should contain appropriate message`)

	err = RouteGaeProd(mux.NewRouter(), nil, 300, `test`, &testEntity{}, nil, nil, nil, dur2, dur2, `test`, ctx, ``, ``, ``, ``)

	assert.Nil(t, err, `err should contain appropriate message`)
}

func Test_MemoryStore(t *testing.T){
	re := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	s := newMemoryStore(func()Entity{return &testEntity{}})

	id, e, err := s.Create()

	assert.True(t, re.MatchString(id), `id should be a valid uuid`)
	assert.Equal(t, 0, e.GetVersion(), `entity Version should be 0`)
	assert.Nil(t, err, `err should be nil`)

	err = s.Update(`not an id`, e)

	assert.Equal(t, 0, e.GetVersion(), `entity Version should be 0`)
	assert.Equal(t, `Non extant error, inner error message: entity with id "not an id" does not exist`, err.Error(), `err should have appropriate message`)

	err = s.Update(id, e)

	assert.Equal(t, 1, e.GetVersion(), `entity Version should be 1`)
	assert.Nil(t, err, `err should be nil`)

	e, err = s.Read(`not an id`)

	assert.Nil(t, e, `entity should be nil`)
	assert.Equal(t, `Non extant error, inner error message: entity with id "not an id" does not exist`, err.Error(), `err should have appropriate message`)

	e, err = s.Read(id)

	assert.Equal(t, 1, e.GetVersion(), `entity Version should be 1`)
	assert.Nil(t, err, `err should be nil`)
}

func Test_GaeStore(t *testing.T){
	re := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	c, _ := aetest.NewContext(nil)
	ctx := appengine.NewContext(c.Request().(*http.Request))
	dur, _ := time.ParseDuration(`1s`)
	s, err := newGaeStore(`testEntity`, ctx, func()Entity{return &testEntity{}}, dur ,dur)

	assert.Nil(t, err, `err should be nil`)

	id, e, err := s.Create()

	assert.True(t, re.MatchString(id), `id should be a valid uuid`)
	assert.Equal(t, 0, e.GetVersion(), `entity Version should be 0`)
	assert.Nil(t, err, `err should be nil`)

	err = s.Update(`not an id`, e)

	assert.Equal(t, 0, e.GetVersion(), `entity Version should be 0`)
	assert.Equal(t, `Non extant error, inner error message: datastore: no such entity`, err.Error(), `err should have appropriate message`)

	err = s.Update(id, e)

	assert.Equal(t, 1, e.GetVersion(), `entity Version should be 1`)
	assert.Nil(t, err, `err should be nil`)

	e, err = s.Read(`not an id`)

	assert.Nil(t, e, `entity should be nil`)
	assert.Equal(t, `Non extant error, inner error message: datastore: no such entity`, err.Error(), `err should have appropriate message`)

	e, err = s.Read(id)

	assert.Equal(t, 1, e.GetVersion(), `entity Version should be 1`)
	assert.Nil(t, err, `err should be nil`)

	te := e.(*testEntity)
	for {
		if now().After(te.DeleteAfter) {
			break
		}
	}

	id2, e, err := s.Create()

	assert.True(t, re.MatchString(id2), `id should be a valid uuid`)
	assert.Equal(t, 0, e.GetVersion(), `entity Version should be 0`)
	assert.Nil(t, err, `err should be nil`)

	e, err = s.Read(id)

	assert.Nil(t, e, `entity should be nil`)
	assert.Equal(t, `Non extant error, inner error message: datastore: no such entity`, err.Error(), `err should have appropriate message`)

	e, err = s.Read(id2)

	assert.Equal(t, 0, e.GetVersion(), `entity Version should be 0`)
	assert.Nil(t, err, `err should be nil`)
}

type testEntity struct{
	Version 	int 		`datastore:",noindex"`
	DeleteAfter time.Time 	`datastore:""`
}

func (te *testEntity) GetVersion() int {
	return te.Version
}

func (te *testEntity) IncrementVersion() {
	te.Version++
}

func (te *testEntity) DecrementVersion() {
	te.Version--
}

func (te *testEntity) SetDeleteAfter(t time.Time) {
	te.DeleteAfter = t
}

func (te *testEntity) IsActive() bool {
	return true
}

func (te *testEntity) CreatedBy() string {
	return `created_by`
}

func (te *testEntity) RegisterNewUser() (string, error) {
	return `new_user`, nil
}

func (te *testEntity) UnregisterUser(userId string) error {
	return nil
}

func (te *testEntity) Kick() bool {
	return false
}