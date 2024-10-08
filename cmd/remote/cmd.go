package main

import (
	"context"
	"os"
	"strconv"

	"go.viam.com/rdk/module"
	// TODO: update to the interface you'll implement
	"go.viam.com/rdk/components/motor"
	"go.viam.com/rdk/config"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	robotimpl "go.viam.com/rdk/robot/impl"
	"go.viam.com/rdk/robot/web"
	_ "go.viam.com/rdk/services/sensors/builtin"
	rdkutils "go.viam.com/rdk/utils"
	"go.viam.com/utils"

	// TODO: update to your project
	winchmotor "github.com/ChrisPullToRefresh/winchmotor"
)

func main() {
	// NewLoggerFromArgs will create a logging.Logger at "DebugLevel" if
	// "--log-level=debug" is an argument in os.Args and at "InfoLevel" otherwise.
	utils.ContextualMain(mainWithArgs, module.NewLoggerFromArgs("winchmotor"))
}

func mainWithArgs(ctx context.Context, args []string, logger logging.Logger) (err error) {

	netconfig := config.NetworkConfig{}
	netconfig.BindAddress = "0.0.0.0:8083"

	if err := netconfig.Validate(""); err != nil {
		return err
	}

	arg_2_converted, _ := strconv.Atoi(os.Args[2])

	// Update the Attributes and ConvertedAttributes with the attributes your modular resource should receive
	conf := &config.Config{
		Network: netconfig,
		Components: []resource.Config{
			{
				Name:  os.Args[1],
				API:   motor.API,
				Model: winchmotor.Model,
				Attributes: rdkutils.AttributeMap{
					"one":   arg_2_converted,
					"two":   os.Args[3],
					"board": os.Args[4],
				},
				ConvertedAttributes: &winchmotor.Config{
					ArgumentOne: arg_2_converted,
					ArgumentTwo: os.Args[3],
					Board:       os.Args[4],
				},
			},
		},
	}

	myRobot, err := robotimpl.New(ctx, conf, logger)
	if err != nil {
		return err
	}

	return web.RunWebWithConfig(ctx, myRobot, conf, logger)
}
