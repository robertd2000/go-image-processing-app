package transformation

import (
	transformDomain "github.com/robertd2000/go-image-processing-app/processor/internal/domain/transformation"
	"github.com/robertd2000/go-image-processing-app/processor/internal/port"
)

type Service struct {
	transformRepo transformDomain.Repository
	storage       port.Storage
	txManager     port.TxManager
}
