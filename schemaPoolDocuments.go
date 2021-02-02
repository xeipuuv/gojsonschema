package gojsonschema

import "sync"

type schemaPoolDocuments struct {
	mp sync.Map
}

func (spd *schemaPoolDocuments) Load(key string) (*schemaPoolDocument, bool) {
	val, ok := spd.mp.Load(key)
	if !ok {
		return nil, false
	}
	return val.(*schemaPoolDocument), true
}

func (spd *schemaPoolDocuments) Delete(key string) {
	spd.mp.Delete(key)

}

func (spd *schemaPoolDocuments) LoadAndDelete(key string) (*schemaPoolDocument, bool) {
	val, ok := spd.mp.LoadAndDelete(key)
	if !ok {
		return nil, false
	}
	return val.(*schemaPoolDocument), true
}

func (spd *schemaPoolDocuments) LoadOrStore(key string, value *schemaPoolDocument) (*schemaPoolDocument, bool) {
	actual, ok := spd.mp.LoadOrStore(key, value)
	if !ok {
		return nil, false
	}
	return actual.(*schemaPoolDocument), true
}

func (spd *schemaPoolDocuments) Range(f func(key string, value *schemaPoolDocument) bool) {
	spd.mp.Range(func(key, value interface{}) bool {
		typedKey := key.(string)
		typedVal := value.(*schemaPoolDocument)
		return f(typedKey, typedVal)
	})
}

func (spd *schemaPoolDocuments) Store(key string, value *schemaPoolDocument) {
	spd.mp.Store(key, value)
}
