//go:build js

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"syscall/js"
)

const leadingParams = 2

var (
	contextType = reflect.TypeFor[context.Context]()
	errorType   = reflect.TypeFor[error]()
	bytesType   = reflect.TypeFor[[]byte]()

	errInvalidArgument = errors.New("invalid argument")
)

func exportRepositories(ctx context.Context, repos any) map[string]any {
	value := reflect.ValueOf(repos)
	exports := make(map[string]any, value.NumField())

	for i := range value.NumField() {
		repoType := value.Type().Field(i).Type
		methods := make(map[string]any, repoType.NumMethod())
		for j := range repoType.NumMethod() {
			method := repoType.Method(j)
			mustBeExportable(repoType, method)
			methods[method.Name] = asyncFunc(bindMethod(ctx, value.Field(i).Method(j), method.Type))
		}
		exports[repoType.Name()] = methods
	}

	return exports
}

func mustBeExportable(repoType reflect.Type, method reflect.Method) {
	signature := method.Type
	numOut := signature.NumOut()
	if signature.NumIn() < leadingParams || signature.In(0) != contextType || signature.In(1).Kind() != reflect.String ||
		numOut < 1 || numOut > 2 || signature.Out(numOut-1) != errorType {
		panic(fmt.Sprintf("%s.%s: expected func(context.Context, string, ...) ([T,] error)", repoType.Name(), method.Name))
	}
}

func bindMethod(ctx context.Context, method reflect.Value, signature reflect.Type) func([]js.Value) (any, error) {
	return func(args []js.Value) (any, error) {
		in := make([]reflect.Value, signature.NumIn())
		in[0] = reflect.ValueOf(ctx)
		in[1] = reflect.Zero(signature.In(1))

		for i := leadingParams; i < len(in); i++ {
			jsArg := js.Undefined()
			if i-leadingParams < len(args) {
				jsArg = args[i-leadingParams]
			}
			arg, err := fromJS(jsArg, signature.In(i))
			if err != nil {
				return nil, fmt.Errorf("%w %d: %w", errInvalidArgument, i-leadingParams, err)
			}
			in[i] = arg
		}

		out := method.Call(in)
		if err, _ := out[len(out)-1].Interface().(error); err != nil {
			return nil, err
		}
		if len(out) == 1 {
			return nil, nil
		}
		return out[0].Interface(), nil
	}
}

func fromJS(value js.Value, target reflect.Type) (reflect.Value, error) {
	if target == bytesType {
		if value.IsNull() || value.IsUndefined() {
			return reflect.Zero(target), nil
		}
		if !value.InstanceOf(js.Global().Get("Uint8Array")) {
			return reflect.Value{}, errors.New("expected Uint8Array")
		}
		data := make([]byte, value.Length())
		js.CopyBytesToGo(data, value)
		return reflect.ValueOf(data), nil
	}

	encoded := "null"
	if !value.IsUndefined() {
		encoded = js.Global().Get("JSON").Call("stringify", value).String()
	}

	decoded := reflect.New(target)
	if err := json.Unmarshal([]byte(encoded), decoded.Interface()); err != nil {
		return reflect.Value{}, err
	}
	return decoded.Elem(), nil
}

func toJS(value any) (js.Value, error) {
	if value == nil {
		return js.Undefined(), nil
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return js.Value{}, err
	}
	result := js.Global().Get("JSON").Call("parse", string(encoded))

	attachBytes(result, reflect.ValueOf(value))
	return result, nil
}

func attachBytes(target js.Value, value reflect.Value) {
	value = reflect.Indirect(value)
	if value.Kind() != reflect.Struct || target.Type() != js.TypeObject {
		return
	}

	for i := range value.NumField() {
		field := value.Type().Field(i)
		if field.Type != bytesType {
			continue
		}
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			continue
		}

		data := value.Field(i).Bytes()
		if data == nil {
			target.Set(name, js.Null())
			continue
		}
		array := js.Global().Get("Uint8Array").New(len(data))
		js.CopyBytesToJS(array, data)
		target.Set(name, array)
	}
}
