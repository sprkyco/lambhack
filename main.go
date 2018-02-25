package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	sparta "github.com/mweagle/Sparta"
	runner "github.com/wickett/lambhack/runner"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	spartaAPIG "github.com/mweagle/Sparta/aws/events"
	gocf "github.com/mweagle/go-cloudformation"
)

////////////////////////////////////////////////////////////////////////////////
type lambhackResponse struct {
	Message string
	Request spartaAPIG.APIGatewayRequest
}

////////////////////////////////////////////////////////////////////////////////
// lambhack event handler
func lambhack(ctx context.Context,
	gatewayEvent spartaAPIG.APIGatewayRequest) (lambhackResponse, error) {

	logger, loggerOk := ctx.Value(sparta.ContextKeyLogger).(*logrus.Logger)
	if loggerOk {
		logger.Info("Lambhack structured log message")
	}

	command := gatewayEvent.QueryParams["command"]
	commandOutput := runner.Run(command)
	// Return a message, together with the incoming input...
	return lambhackResponse{
		Message: fmt.Sprintf("Welcome to lambhack!" + commandOutput),
		//Request: gatewayEvent,
	}, nil
}

func lambhackLambdaFunctions(api *sparta.API) []*sparta.LambdaAWSInfo {
	var lambdaFunctions []*sparta.LambdaAWSInfo
	lambdaFn := sparta.HandleAWSLambda(sparta.LambdaName(lambhack),
		lambhack,
		sparta.IAMRoleDefinition{})

	if nil != api {
		apiGatewayResource, _ := api.NewResource("/lambhack", lambdaFn)

		// We only return http.StatusOK
		apiMethod, apiMethodErr := apiGatewayResource.NewMethod("GET",
			http.StatusOK,
			http.StatusOK)
		if nil != apiMethodErr {
			panic("Failed to create /lambhack resource: " + apiMethodErr.Error())
		}
		// The lambda resource only supports application/json Unmarshallable
		// requests.
		apiMethod.SupportedRequestContentTypes = []string{"application/json"}
	}
	return append(lambdaFunctions, lambdaFn)
}

////////////////////////////////////////////////////////////////////////////////
// Main
func main() {

	// Register the function with the API Gateway
	apiStage := sparta.NewStage("prod")
	apiGateway := sparta.NewAPIGateway("lambhack", apiStage)

	// Deploy it
	stackName := spartaCF.UserScopedStackName("lambhack")
	sparta.Main(stackName,
		fmt.Sprintf("Sacrificial lambs"),
		lambhackLambdaFunctions(apiGateway),
		apiGateway,
		s3Site)
}
