// Package customsensor implements a sensor where all methods are unimplemented.
// It extends the built-in resource subtype sensor and implements methods to handle resource construction and attribute configuration.
// TODO: rename if needed (i.e., custommotor)
package custummotor

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	// TODO: update to the interface you'll implement
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/components/motor"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/operation"
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

const (
	winchCwPin            = "35"
	winchCcwPin           = "37"
	winchPwmFrequency     = 500
	winchStopPwmDutyCycle = 0
	winchSlowPwmDutyCycle = 0.2
	winchFastPwmDutyCycle = 1.0
)

type winchState string

const (
	raiseWinchState   = "raisingWinch"
	stoppedWinchState = "stoppedWinch"
	lowerWinchState   = "loweringWinch"
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
	Board          string `json:"board"`
	SensorLoadCell string `json:"sensor-load-cell"`
}

// Validate validates the config and returns implicit dependencies.
// TODO: Change the Validate function to validate any config variables.
func (cfg *Config) Validate(path string) ([]string, error) {

	if cfg.Board == "" {
		return nil, utils.NewConfigValidationFieldRequiredError(path, "board")
	}

	if cfg.SensorLoadCell == "" {
		return nil, utils.NewConfigValidationFieldRequiredError(path, "sensor-load-cell")
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
		opMgr:      operation.NewSingleOperationManager(),
	}

	// TODO: If your custom component has dependencies, perform any checks you need to on them.

	// The Reconfigure() method changes the values on the customSensor based on the attributes in the component config
	if err := m.Reconfigure(ctx, deps, rawConf); err != nil {
		logger.Error("Error configuring module with ", rawConf)
		return nil, err
	}

	m.resetWinch()
	m.ws = stoppedWinchState

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
	mu         sync.Mutex
	opMgr      *operation.SingleOperationManager

	b             board.Board
	lc            resource.Sensor
	ws            winchState
	emergencyStop bool
	powerPct      float64
}

// GoTo implements motor.Motor.
func (m *customMotor) GoTo(ctx context.Context, rpm float64, positionRevolutions float64, extra map[string]interface{}) error {
	return fmt.Errorf("GoTo not yet implemented")
}

// GoFor implements motor.Motor.
func (m *customMotor) GoFor(ctx context.Context, rpm float64, revolutions float64, extra map[string]interface{}) error {
	return fmt.Errorf("GoFor not yet implemented")
}

// IsMoving implements motor.Motor.
func (m *customMotor) IsMoving(context.Context) (bool, error) {
	return m.ws != stoppedWinchState, nil
}

// IsPowered implements motor.Motor.
func (m *customMotor) IsPowered(ctx context.Context, extra map[string]interface{}) (bool, float64, error) {
	isPowered := m.ws != stoppedWinchState
	powerPct := 0.0
	if isPowered {
		powerPct = m.powerPct
	}
	return isPowered, powerPct, nil
}

// Position implements motor.Motor.
func (m *customMotor) Position(ctx context.Context, extra map[string]interface{}) (float64, error) {
	return 0.0, fmt.Errorf("Position not yet implemented")
}

// Properties implements motor.Motor.
func (m *customMotor) Properties(ctx context.Context, extra map[string]interface{}) (motor.Properties, error) {
	return motor.Properties{}, fmt.Errorf("ResetZeroPosition not yet implemented")
}

// ResetZeroPosition implements motor.Motor.
func (m *customMotor) ResetZeroPosition(ctx context.Context, offset float64, extra map[string]interface{}) error {
	return fmt.Errorf("ResetZeroPosition not yet implemented")
}

func (m *customMotor) setPin(pinName string, high bool) {
	pin, err := m.b.GPIOPinByName(pinName)
	if err != nil {
		m.logger.Error(err)
		return
	}
	err = pin.Set(context.Background(), high, nil)
	if err != nil {
		m.logger.Error(err)
		return
	}
}

func (m *customMotor) setPwmFrequency(pinName string, freqHz uint) {
	pin, err := m.b.GPIOPinByName(pinName)
	if err != nil {
		m.logger.Error(err)
		return
	}
	err = pin.SetPWMFreq(m.cancelCtx, freqHz, nil)
	if err != nil {
		m.logger.Error(err)
		return
	}
}

func (m *customMotor) setPwmDutyCycle(pinName string, dutyCyclePct float64) {
	pin, err := m.b.GPIOPinByName(pinName)
	if err != nil {
		m.logger.Error(err)
		return
	}
	err = pin.SetPWM(m.cancelCtx, dutyCyclePct, nil)
	if err != nil {
		m.logger.Error(err)
		return
	}
}

func (m *customMotor) resetWinch() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.setPin(winchCwPin, false)
	m.setPin(winchCcwPin, false)

	m.setPwmFrequency(winchCwPin, winchPwmFrequency)
	m.setPwmFrequency(winchCcwPin, winchPwmFrequency)
}

func (m *customMotor) stopWinch() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.setPwmDutyCycle(winchCwPin, winchStopPwmDutyCycle)
	m.setPwmDutyCycle(winchCcwPin, winchStopPwmDutyCycle)

	m.ws = stoppedWinchState
}

func iotaEqual(x, y float64) bool {
	iota := 0.001
	return math.Abs(x-y) <= iota
}

// SetPower implements motor.Motor.
// powerPct > 0 == raise == cw
func (m *customMotor) SetPower(ctx context.Context, powerPct float64, extra map[string]interface{}) error {
	m.opMgr.CancelRunning(ctx)

	if iotaEqual(powerPct, 0.0) {
		return m.Stop(ctx, nil)
	}

	{
		m.mu.Lock()
		defer m.mu.Unlock()
		var pin string
		if powerPct > 0 {
			pin = winchCwPin
			m.ws = raiseWinchState
		} else {
			pin = winchCcwPin
			m.ws = lowerWinchState
		}
		newPowerPct := math.Abs(powerPct)
		m.setPwmDutyCycle(pin, newPowerPct)
		m.powerPct = newPowerPct
	}

	// If winch is being raised, continually check if
	// load cell value is too high
	if m.ws == raiseWinchState {
		ctx, done := m.opMgr.New(ctx)
		defer done()
		return m.raiseWinchCarefully(ctx)
	} else {
		return nil
	}
}

// All callers must register an operation via `m.opMgr.New`
func (m *customMotor) raiseWinchCarefully(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			time.Sleep(time.Millisecond * 5)
		}
	}
}

// SetRPM implements motor.Motor.
func (m *customMotor) SetRPM(ctx context.Context, rpm float64, extra map[string]interface{}) error {
	return fmt.Errorf("SetRPM not yet implemeented")
}

// Stop implements motor.Motor.
func (m *customMotor) Stop(ctx context.Context, extra map[string]interface{}) error {
	m.opMgr.CancelRunning(ctx)

	m.stopWinch()
	return nil
}

// TODO: rename as needed (i.e., m customMotor)
func (m *customMotor) Name() resource.Name {
	return m.name
}

// Reconfigures the model. Most models can be reconfigured in place without needing to rebuild. If you need to instead create a new instance of the sensor, throw a NewMustBuildError.
// TODO: rename as appropriate, i.e. m *customMotor
func (m *customMotor) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	m.opMgr.CancelRunning(ctx)

	m.mu.Lock()
	defer m.mu.Unlock()

	// TODO: rename as appropriate (i.e., motorConfig)
	motorConfig, err := resource.NativeConfig[*Config](conf)
	if err != nil {
		m.logger.Warn("Error reconfiguring module with ", err)
		return err
	}

	m.name = conf.ResourceName()

	m.b, err = board.FromDependencies(deps, motorConfig.Board)
	if err != nil {
		return fmt.Errorf("unable to get motor %v for %v", motorConfig.Board, m.name)
	}
	m.logger.Info("board is now configured to ", m.b.Name())

	m.lc, err = sensor.FromDependencies(deps, motorConfig.SensorLoadCell)
	if err != nil {
		return fmt.Errorf("unable to get load cell sensor %v for %v", motorConfig.SensorLoadCell, m.name)
	}
	m.logger.Info("load cell sensor is now configured")

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
	err := m.Stop(ctx, nil)
	m.cancelFunc()
	return err
}
