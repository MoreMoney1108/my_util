package middleware

import (
	"log"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

const (
	RequestParamKey = "request_param_key"
)

const (
	ParamTypeStr   = "param_type"
	ParamTypeUri   = "uri"
	ParamTypeQuery = "query"
	ParamTypeBody  = "body"
)

func BindRequestParam[T any]() func(*gin.Context) {
	return func(c *gin.Context) {
		req := new(T)

		value := reflect.ValueOf(req)
		for value.Kind() == reflect.Pointer {
			if value.IsNil() {
				value = reflect.New(value.Type().Elem())
			}

			value = value.Elem()
		}

		reqInterface := value.Addr().Interface()

		if value.Kind() != reflect.Struct {
			panic("not is struct")
		}

		for i := 0; i < value.NumField(); i++ {
			fieldType := value.Type().Field(i)
			fieldValue := value.Field(i)
			if !fieldType.Anonymous {
				continue
			}

			paramType := fieldType.Tag.Get(ParamTypeStr)
			if len(paramType) < 1 {
				continue
			}

			var bindFunc func(any) error
			switch paramType {
			case ParamTypeUri:
				bindFunc = c.BindUri

			case ParamTypeQuery:
				bindFunc = c.BindQuery

			case ParamTypeBody:
				bindFunc = c.BindJSON

			default:
				panic("not support param type")
			}

			if err := bindFunc(fieldValue.Addr().Interface()); err != nil {
				log.Printf("failed to bind request param, err: %v", err)
				_ = c.AbortWithError(http.StatusBadRequest, err)
				return
			}
		}

		c.Set(RequestParamKey, reqInterface)
		c.Next()
	}
}
