package scripts

import (
	"Hexo-AutoCD/config"
	"fmt"
	"os/exec"
	"path/filepath"
)

// ExecutionResult 定义脚本执行结果
type ExecutionResult struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error,omitempty"`
}

// ScriptExecutor 定义脚本执行器接口
type ScriptExecutor interface {
	Execute(event string) (*ExecutionResult, error)
}

// DefaultExecutor 默认的脚本执行器实现
type DefaultExecutor struct {
	config *config.Config
}

// NewExecutor 创建新的执行器实例
func NewExecutor(cfg *config.Config) ScriptExecutor {
	return &DefaultExecutor{
		config: cfg,
	}
}

// Execute 执行指定事件对应的脚本
func (e *DefaultExecutor) Execute(event string) (*ExecutionResult, error) {
	scriptPath, ok := e.config.Scripts[event]
	if !ok {
		return nil, fmt.Errorf("no script configured for event: %s", event)
	}

	fullPath := filepath.Join(e.config.ScriptsPath, scriptPath)
	cmd := exec.Command("/bin/bash", fullPath)

	output, err := cmd.CombinedOutput()
	result := &ExecutionResult{
		Output:   string(output),
		ExitCode: 0,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
		result.Error = err.Error()
	}

	return result, nil
}
