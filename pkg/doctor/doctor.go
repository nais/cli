package doctor

import (
	"fmt"
	"io"

	nais_io_v1alpha1 "github.com/nais/liberator/pkg/apis/nais.io/v1alpha1"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var ErrSkip = fmt.Errorf("skip")

type Config struct {
	Application   *nais_io_v1alpha1.Application
	K8sClient     kubernetes.Interface
	DynamicClient dynamic.Interface
	Log           *logrus.Entry
	Out           io.Writer
}

type Error struct {
	Human string
	Err   error
}

func (e Error) Error() string {
	return e.Human
}

// ErrorMsg return a error message with the given message and the error message
// if err is nil, it will return nil
func ErrorMsg(err error, msg string) error {
	if err == nil {
		return nil
	}

	return &Error{
		Human: msg,
		Err:   err,
	}
}
