// Package customsensor implements a sensor where all methods are unimplemented.
// It extends the built-in resource subtype sensor and implements methods to handle resource construction and attribute configuration.
// TODO: rename if needed (i.e., custommotor)
package custummotor

import (
	"context"
	"errors"

	"fmt"

	// TODO: update to the interface you'll implement
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/components/motor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"go.viam.com/utils"
)

// Here is where we define your new model's colon-delimited-triplet (viam-labs:go-module-templates-sensor:customsensor)
// viam-labs = namespace, go-module-templates-sensor = repo-name, customsensor = model name.
// TODO: Change model namespace, family (often the repo-name), and model. For more information see https://docs.viam.com/registry/create/#name-your-new-resource-model
var (
	Model            = resource.NewModel("pulltorefresh", "winchmotor", "winchmotor")
	errUnimplemented = errors.New("unimplemented")
)

func init() {
	resource.RegisterComponent(motor.API, Model,
		// TODO: update to the interface you'll implement
		resource.Registration[motor.Motor, *Config]{
			Constructor: newCustomMotor,
		},
	)
}

// TODO: Change the Config struct to contain any values that you would like to be able to configure from the attributes field in the component
// configuration. For more information see https://docs.viam.com/build/configure/#components
type Config struct {
	ArgumentOne int    `json:"one"`
	ArgumentTwo string `json:"two"`
	Board       string `json:"board"`
}

// Validate validates the config and returns implicit dependencies.
// TODO: Change the Validate function to validate any config variables.
func (cfg *Config) Validate(path string) ([]string, error) {
	if cfg.ArgumentOne == 0 {
		return nil, utils.NewConfigValidationFieldRequiredError(path, "one")
	}

	if cfg.ArgumentTwo == "" {
		return nil, utils.NewConfigValidationFieldRequiredError(path, "two")
	}

	if cfg.Board == "" {
		return nil, utils.NewConfigValidationFieldRequiredError(path, "board")
	}

	// TODO: return implicit dependencies if needed as the first value
	return []string{}, nil
}

// Constructor for a custom sensor that creates and returns a customSensor.
// TODO: update the customSensor struct and the initialization, and rename it
// if needed (i.e., newCustomMotor)
func newCustomMotor(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (motor.Motor, error) {
	// This takes the generic resource.Config passed down from the parent and converts it to the
	// model-specific (aka "native") Config structure defined above, making it easier to directly access attributes.
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	// Create a cancelable context for custom sensor
	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	m := &customMotor{
		name:       rawConf.ResourceName(),
		logger:     logger,
		cfg:        conf,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
	}

	// TODO: If your custom component has dependencies, perform any checks you need to on them.

	// The Reconfigure() method changes the values on the customSensor based on the attributes in the component config
	if err := m.Reconfigure(ctx, deps, rawConf); err != nil {
		logger.Error("Error configuring module with ", rawConf)
		return nil, err
	}

	return m, nil
}

// TODO: update the customSensor struct with any fields you require and
// rename the struct as needed (i.e., customMotor)
type customMotor struct {
	name   resource.Name
	logger logging.Logger
	cfg    *Config

	cancelCtx  context.Context
	cancelFunc func()

	argumentOne int
	argumentTwo string

	b board.Board
}

// GoTo implements motor.Motor.
func (m *customMotor) GoTo(ctx context.Context, rpm float64, positionRevolutions float64, extra map[string]interface{}) error {
	panic("unimplemented")
}

// GoFor implements motor.Motor.
func (m *customMotor) GoFor(ctx context.Context, rpm float64, revolutions float64, extra map[string]interface{}) error {
	panic("unimplemented")
}

// IsMoving implements motor.Motor.
func (m *customMotor) IsMoving(context.Context) (bool, error) {
	panic("unimplemented")
}

// IsPowered implements motor.Motor.
func (m *customMotor) IsPowered(ctx context.Context, extra map[string]interface{}) (bool, float64, error) {
	panic("unimplemented")
}

// Position implements motor.Motor.
func (m *customMotor) Position(ctx context.Context, extra map[string]interface{}) (float64, error) {
	panic("unimplemented")
}

// Properties implements motor.Motor.
func (m *customMotor) Properties(ctx context.Context, extra map[string]interface{}) (motor.Properties, error) {
	panic("unimplemented")
}

// ResetZeroPosition implements motor.Motor.
func (m *customMotor) ResetZeroPosition(ctx context.Context, offset float64, extra map[string]interface{}) error {
	panic("unimplemented")
}

// SetPower implements motor.Motor.
func (m *customMotor) SetPower(ctx context.Context, powerPct float64, extra map[string]interface{}) error {
	panic("unimplemented")
}

// SetRPM implements motor.Motor.
func (m *customMotor) SetRPM(ctx context.Context, rpm float64, extra map[string]interface{}) error {
	panic("unimplemented")
}

// Stop implements motor.Motor.
func (m *customMotor) Stop(context.Context, map[string]interface{}) error {
	panic("unimplemented")
}

// TODO: rename as needed (i.e., m customMotor)
func (m *customMotor) Name() resource.Name {
	return m.name
}

// Reconfigures the model. Most models can be reconfigured in place without needing to rebuild. If you need to instead create a new instance of the sensor, throw a NewMustBuildError.
// TODO: rename as appropriate, i.e. m *customMotor
func (m *customMotor) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	// TODO: rename as appropriate (i.e., motorConfig)
	motorConfig, err := resource.NativeConfig[*Config](conf)
	if err != nil {
		m.logger.Warn("Error reconfiguring module with ", err)
		return err
	}

	m.argumentOne = motorConfig.ArgumentOne
	m.argumentTwo = motorConfig.ArgumentTwo

	m.name = conf.ResourceName()
	m.logger.Info("one is now configured to: ", m.argumentOne)
	m.logger.Info("two is now configured to ", m.argumentTwo)

	m.b, err = board.FromDependencies(deps, motorConfig.Board)
	if err != nil {
		return fmt.Errorf("unable to get motor %v for %v", motorConfig.Board, m.name)
	}

	return nil
}

// DoCommand is a place to add additional commands to extend the sensor API. This is optional.
// TODO: rename as appropriate (i.e., motorConfig)
func (m *customMotor) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	m.logger.Error("DoCommand method unimplemented")
	return nil, errUnimplemented
}

// Close closes the underlying generic.
// TODO: rename as appropriate (i.e., motorConfig)
func (m *customMotor) Close(ctx context.Context) error {
	m.cancelFunc()
	return nil
}
