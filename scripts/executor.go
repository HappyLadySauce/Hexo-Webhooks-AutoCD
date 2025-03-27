package scripts

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ExecutionResult 定义脚本执行结果
// 这个结构体用于存储脚本执行后的各种状态
type ExecutionResult struct {
	Output   string   `json:"output"`          // 脚本的输出内容
	ExitCode int      `json:"exit_code"`       // 脚本的退出码，0表示成功，非0表示失败
	Error    string   `json:"error,omitempty"` // 如果执行出错，这里存储错误信息
	Logs     []string `json:"logs"`            // 执行日志
}

// ScriptExecutor 定义脚本执行器接口
// 使用接口可以方便后续扩展不同的执行器实现（比如远程执行、容器内执行等）
type ScriptExecutor interface {
	Execute(event string, payload interface{}) (*ExecutionResult, error)
}

// ExecutorConfig 定义执行器配置
type ExecutorConfig struct {
	ScriptsPath   string        // 脚本所在目录
	Timeout       time.Duration // 脚本执行超时时间
	MaxConcurrent int           // 最大并发执行数
	DefaultEnv    []string      // 默认环境变量
}

// DefaultExecutor 默认的脚本执行器实现
type DefaultExecutor struct {
	config     ExecutorConfig       // 执行器配置
	semaphore  chan struct{}        // 信号量，用于控制并发执行数
	mu         sync.RWMutex         // 互斥锁，用于保护并发访问
	executions map[string]*exec.Cmd // 正在执行的脚本映射
}

// NewExecutor 创建新的执行器实例
func NewExecutor(config ExecutorConfig) ScriptExecutor {
	// 确保配置合理
	if config.MaxConcurrent <= 0 {
		config.MaxConcurrent = 5 // 默认最大并发数
	}
	if config.Timeout <= 0 {
		config.Timeout = 5 * time.Minute // 默认超时时间
	}

	return &DefaultExecutor{
		config:     config,
		semaphore:  make(chan struct{}, config.MaxConcurrent),
		executions: make(map[string]*exec.Cmd),
	}
}

// Execute 执行指定事件对应的脚本
// event: 触发事件的类型（如 push, release 等）
// payload: 事件的详细信息，会转换为环境变量传递给脚本
func (e *DefaultExecutor) Execute(event string, payload interface{}) (*ExecutionResult, error) {
	// 构建脚本路径
	scriptPath := filepath.Join(e.config.ScriptsPath, event)

	// 检查脚本是否存在
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("脚本不存在: %s", event)
	}

	// 获取执行许可
	e.semaphore <- struct{}{}        // 占用一个并发槽
	defer func() { <-e.semaphore }() // 释放并发槽

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), e.config.Timeout)
	defer cancel()

	// 准备命令
	cmd := exec.CommandContext(ctx, "/bin/bash", scriptPath)

	// 设置工作目录
	cmd.Dir = e.config.ScriptsPath

	// 设置环境变量
	// 首先添加系统环境变量
	env := os.Environ()

	// 然后添加配置中的默认环境变量
	if len(e.config.DefaultEnv) > 0 {
		env = append(env, e.config.DefaultEnv...)
	}

	// 设置最终的环境变量
	cmd.Env = env

	// 如果有payload，将其转换为环境变量
	if payload != nil {
		// TODO: 将 payload 转换为环境变量
		// 例如：GITHUB_EVENT=push, GITHUB_REPO=xxx 等
	}

	// 记录正在执行的命令
	e.mu.Lock()
	e.executions[event] = cmd
	e.mu.Unlock()

	// 清理函数
	defer func() {
		e.mu.Lock()
		delete(e.executions, event)
		e.mu.Unlock()
	}()

	// 创建管道用于实时获取输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("无法创建输出管道: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("无法创建错误输出管道: %v", err)
	}

	// 创建多路复用的输出
	var outputBuffer bytes.Buffer
	var logs []string

	// 启动命令
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动脚本失败: %v", err)
	}

	// 创建等待组
	var wg sync.WaitGroup
	wg.Add(2)

	// 处理标准输出
	go func() {
		defer wg.Done()
		reader := bufio.NewReader(stdout)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					logs = append(logs, fmt.Sprintf("读取输出错误: %v", err))
				}
				break
			}
			line = strings.TrimSpace(line)
			if line != "" {
				logs = append(logs, line)
				outputBuffer.WriteString(line + "\n")
			}
		}
	}()

	// 处理错误输出
	go func() {
		defer wg.Done()
		reader := bufio.NewReader(stderr)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					logs = append(logs, fmt.Sprintf("读取错误输出错误: %v", err))
				}
				break
			}
			line = strings.TrimSpace(line)
			if line != "" {
				logs = append(logs, "错误: "+line)
				outputBuffer.WriteString("错误: " + line + "\n")
			}
		}
	}()

	// 等待命令完成
	err = cmd.Wait()

	// 等待所有输出处理完成
	wg.Wait()

	// 准备执行结果
	result := &ExecutionResult{
		Output:   outputBuffer.String(),
		ExitCode: 0,
		Logs:     logs,
	}

	// 处理执行错误
	if err != nil {
		result.Error = err.Error()
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
	}

	// 检查是否超时
	if ctx.Err() == context.DeadlineExceeded {
		result.Error = "script execution timed out"
		result.ExitCode = -1
		logs = append(logs, "脚本执行超时")
	}

	return result, nil
}

// Stop 停止正在执行的脚本
func (e *DefaultExecutor) Stop(event string) error {
	e.mu.RLock()
	cmd, exists := e.executions[event]
	e.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no running script found for event: %s", event)
	}

	return cmd.Process.Kill()
}

// StopAll 停止所有正在执行的脚本
func (e *DefaultExecutor) StopAll() {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, cmd := range e.executions {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}
}
