package module

type ICrudResponse interface {
	IModule
	Success(data interface{}) interface{}
	Error(err error) interface{}
	Pagenation(items interface{}, total int64, page int, size int) interface{}
}
