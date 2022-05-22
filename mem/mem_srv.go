package mem

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

type SingletonMemStorage interface {
	setStart()
	isStart() bool
	initData()
	getData() *sync.Map
}

type singletonMemStorage struct {
	sync.RWMutex
	data       sync.Map
	isSrvStart bool
}

var instance *singletonMemStorage
var once sync.Once

func GetMemInstance() SingletonMemStorage {
	once.Do(func() {
		instance = new(singletonMemStorage)
	})

	return instance
}

func (s *singletonMemStorage) setStart() {
	s.Lock()
	defer s.Unlock()
	s.isSrvStart = true
}

func (s *singletonMemStorage) isStart() bool {
	s.RLock()
	defer s.RUnlock()
	return s.isSrvStart
}

func (s *singletonMemStorage) initData() {
	s.Lock()
	defer s.Unlock()
	s.data = sync.Map{}
}

func (s *singletonMemStorage) getData() *sync.Map {
	s.RLock()
	defer s.RUnlock()
	return &(s.data)
}

func (m *memoryStorage) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:

		filename := r.FormValue(urlValue)

		data, ok := m.data.Load(filename)

		if !ok {
			writeResponse(w, http.StatusBadRequest, http.StatusText(500))
		}

		dataUnit, ok := data.(dataUnit)

		//copy the relevant headers. If you want to preserve the downloaded file name, extract it with go's url parser.
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%v", filename))

		//stream the body to the client without fully loading it into memory
		_, err := io.Copy(w, bytes.NewReader(dataUnit.bytes))
		if err != nil {
			log.Println(err)
		}

	default:
		writeResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func writeResponse(w http.ResponseWriter, code int, v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(`{"error":"Internal server error"}`))
		if err != nil {
			log.Println(err)
		}
		return
	}
	w.WriteHeader(code)
	_, err = w.Write(b)
	if err != nil {
		log.Println(err)
	}
}
