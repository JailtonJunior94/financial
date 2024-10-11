package middlewares

import (
	"context"
	"net/http"

	"github.com/JailtonJunior94/devkit-go/pkg/responses"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
)

func ErrorHandler(ctx context.Context, w http.ResponseWriter, err error) {
	if err != nil {
		responseError := httperrors.GetResponseError(err)
		responses.JSON(w, responseError.Code, responseError)
		return
	}
}
