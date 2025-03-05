package unit

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
)

func a() error {
	// 初始错误
	return errors.New("a error")
}

func b() error {
	err := a()
	if err != nil {
		// 包装新的错误并返回
		newErr := errors.WithMessage(err, "b error")
		// newErr := errors.Wrap(err, "b error")

		// 可以从包装后的错误中还原出初始错误
		//fmt.Printf("newErr cause == err: %t\n", errors.Cause(newErr) == err)

		return newErr
	}
	return nil
}

func test_error() {
	err := b()
	if err != nil {
		// %v 打印错误信息
		//fmt.Printf("%v\n", err)

		// %+v 打印错误信息和错误堆栈
		//fmt.Printf("%+v\n", err)
		//fmt.Println(fmt.Sprintf("%+v", err))
		// 打印错误根因
		fmt.Printf("`````%s\n", errors.Cause(err))
		return
	}
	fmt.Println("success")
}

func TestRunErr(t *testing.T) {
	test_error()
}
